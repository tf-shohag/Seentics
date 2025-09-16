'use client';

import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  BarChart3,
  Calendar,
  Eye,
  Filter,
  Globe,
  RefreshCw,
  Shield,
  TrendingUp,
  Users,
  Workflow
} from 'lucide-react';
import { useEffect, useState } from 'react';
import api from '@/lib/api';

interface AdminStats {
  total_users: number;
  total_websites: number;
  total_workflows: number;
  total_funnels: number;
  total_events: number;
  recent_users: any[];
  recent_websites: any[];
  recent_workflows: any[];
  recent_funnels: any[];
  system_health: any;
  timestamp: string;
}

interface UsersList {
  users: any[];
  pagination: {
    currentPage: number;
    totalPages: number;
    totalUsers: number;
    hasNextPage: boolean;
    hasPrevPage: boolean;
  };
}

interface WebsitesList {
  websites: any[];
  pagination: {
    currentPage: number;
    totalPages: number;
    totalWebsites: number;
    hasNextPage: boolean;
    hasPrevPage: boolean;
  };
}

export default function AdminPage() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [adminCode, setAdminCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [usersList, setUsersList] = useState<UsersList | null>(null);
  const [websitesList, setWebsitesList] = useState<WebsitesList | null>(null);
  const [activeTab, setActiveTab] = useState('overview');

  const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';

  useEffect(() => {
    // Check if admin code is stored in localStorage
    const storedAdminCode = localStorage.getItem('admin_code');
    if (storedAdminCode) {
      setAdminCode(storedAdminCode);
      setIsAuthenticated(true);
      fetchAdminData();
    }
  }, []);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    // Validate admin code with environment variable
    const expectedAdminCode = process.env.NEXT_PUBLIC_ADMIN_CODE || 'admin123';

    if (adminCode !== expectedAdminCode) {
      setError('Invalid admin code');
      setLoading(false);
      return;
    }

    // Store admin code in localStorage
    localStorage.setItem('admin_code', adminCode);
    setIsAuthenticated(true);
    fetchAdminData();
    setLoading(false);
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_code');
    setIsAuthenticated(false);
    setAdminCode('');
    setStats(null);
    setUsersList(null);
    setWebsitesList(null);
  };

  const fetchAdminData = async () => {
    setLoading(true);
    try {
      // Fetch admin statistics using global axios instance
      const statsResponse = await api.get('/admin/stats');
      if (statsResponse.status === 200) {
        setStats(statsResponse.data);
      }

      // Fetch users list using global axios instance
      const usersResponse = await api.get('/admin/users?page=1&limit=10');
      if (usersResponse.status === 200) {
        setUsersList(usersResponse.data);
      }

      // Fetch websites list using global axios instance
      const websitesResponse = await api.get('/admin/websites?page=1&limit=10');
      if (websitesResponse.status === 200) {
        setWebsitesList(websitesResponse.data);
      }
    } catch (err) {
      console.error('Admin data fetch error:', err);
      setError('Failed to fetch admin data');
    } finally {
      setLoading(false);
    }
  };

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat().format(num);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
              <Shield className="w-6 h-6 text-blue-600" />
            </div>
            <CardTitle className="text-2xl">Admin Access</CardTitle>
            <CardDescription>
              Enter the admin code to access the dashboard
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleLogin} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="adminCode">Admin Code</Label>
                <Input
                  id="adminCode"
                  type="password"
                  value={adminCode}
                  onChange={(e) => setAdminCode(e.target.value)}
                  placeholder="Enter admin code"
                  required
                />
              </div>
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? (
                  <>
                    <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                    Verifying...
                  </>
                ) : (
                  'Access Dashboard'
                )}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <Shield className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-semibold text-gray-900">Admin Dashboard</h1>
                <p className="text-sm text-gray-500">Seentics Cloud Platform</p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <Button
                variant="outline"
                size="sm"
                onClick={fetchAdminData}
                disabled={loading}
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                Refresh
              </Button>
              <Button variant="outline" size="sm" onClick={handleLogout}>
                Logout
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="users">Users</TabsTrigger>
            <TabsTrigger value="websites">Websites</TabsTrigger>
            <TabsTrigger value="analytics">Analytics</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            {/* Stats Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Total Users</p>
                      <p className="text-2xl font-bold text-gray-900">
                        {stats ? formatNumber(stats.total_users) : '---'}
                      </p>
                    </div>
                    <Users className="w-8 h-8 text-blue-600" />
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Websites</p>
                      <p className="text-2xl font-bold text-gray-900">
                        {stats ? formatNumber(stats.total_websites) : '---'}
                      </p>
                    </div>
                    <Globe className="w-8 h-8 text-green-600" />
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Workflows</p>
                      <p className="text-2xl font-bold text-gray-900">
                        {stats ? formatNumber(stats.total_workflows) : '---'}
                      </p>
                    </div>
                    <Workflow className="w-8 h-8 text-purple-600" />
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Funnels</p>
                      <p className="text-2xl font-bold text-gray-900">
                        {stats ? formatNumber(stats.total_funnels) : '---'}
                      </p>
                    </div>
                    <Filter className="w-8 h-8 text-orange-600" />
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Total Events</p>
                      <p className="text-2xl font-bold text-gray-900">
                        {stats ? formatNumber(stats.total_events) : '---'}
                      </p>
                    </div>
                    <BarChart3 className="w-8 h-8 text-red-600" />
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Recent Activity */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center">
                    <Users className="w-5 h-5 mr-2" />
                    Recent Users
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {stats?.recent_users?.slice(0, 5).map((user, index) => (
                      <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                        <div>
                          <p className="font-medium text-sm">{user.email || user.username}</p>
                          <p className="text-xs text-gray-500">
                            {user.createdAt ? formatDate(user.createdAt) : 'N/A'}
                          </p>
                        </div>
                        <Badge variant="secondary">{user.subscription || 'Free'}</Badge>
                      </div>
                    )) || (
                        <p className="text-gray-500 text-center py-4">No recent users</p>
                      )}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center">
                    <Globe className="w-5 h-5 mr-2" />
                    Recent Websites
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {stats?.recent_websites?.slice(0, 5).map((website, index) => (
                      <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                        <div>
                          <p className="font-medium text-sm">{website.name || website.domain}</p>
                          <p className="text-xs text-gray-500">
                            {website.createdAt ? formatDate(website.createdAt) : 'N/A'}
                          </p>
                        </div>
                        <Badge variant="outline">
                          <Eye className="w-3 h-3 mr-1" />
                          Active
                        </Badge>
                      </div>
                    )) || (
                        <p className="text-gray-500 text-center py-4">No recent websites</p>
                      )}
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="users" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Users Management</CardTitle>
                <CardDescription>
                  Total users: {usersList?.pagination?.totalUsers || 0}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {usersList?.users?.map((user, index) => (
                    <div key={index} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                            <Users className="w-5 h-5 text-blue-600" />
                          </div>
                          <div>
                            <p className="font-medium">{user.email}</p>
                            <p className="text-sm text-gray-500">
                              Joined: {user.createdAt ? formatDate(user.createdAt) : 'N/A'}
                            </p>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <Badge variant={user.subscription === 'pro' ? 'default' : 'secondary'}>
                          {user.subscription || 'Free'}
                        </Badge>
                        {user.lastLogin && (
                          <Badge variant="outline">
                            <Calendar className="w-3 h-3 mr-1" />
                            {formatDate(user.lastLogin)}
                          </Badge>
                        )}
                      </div>
                    </div>
                  )) || (
                      <p className="text-gray-500 text-center py-8">No users found</p>
                    )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="websites" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Websites Management</CardTitle>
                <CardDescription>
                  Total websites: {websitesList?.pagination?.totalWebsites || 0}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {websitesList?.websites?.map((website, index) => (
                    <div key={index} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                            <Globe className="w-5 h-5 text-green-600" />
                          </div>
                          <div>
                            <p className="font-medium">{website.name || website.domain}</p>
                            <p className="text-sm text-gray-500">
                              Owner: {website.userId?.email || 'Unknown'}
                            </p>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <Badge variant="outline">
                          <TrendingUp className="w-3 h-3 mr-1" />
                          Active
                        </Badge>
                        <Badge variant="secondary">
                          {website.createdAt ? formatDate(website.createdAt) : 'N/A'}
                        </Badge>
                      </div>
                    </div>
                  )) || (
                      <p className="text-gray-500 text-center py-8">No websites found</p>
                    )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="analytics" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Analytics Overview</CardTitle>
                <CardDescription>
                  System analytics and performance metrics
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="text-center p-6 bg-blue-50 rounded-lg">
                    <BarChart3 className="w-12 h-12 text-blue-600 mx-auto mb-2" />
                    <p className="text-2xl font-bold text-blue-900">
                      {stats ? formatNumber(stats.total_events) : '---'}
                    </p>
                    <p className="text-sm text-blue-700">Total Events</p>
                  </div>
                  <div className="text-center p-6 bg-green-50 rounded-lg">
                    <Filter className="w-12 h-12 text-green-600 mx-auto mb-2" />
                    <p className="text-2xl font-bold text-green-900">
                      {stats ? formatNumber(stats.total_funnels) : '---'}
                    </p>
                    <p className="text-sm text-green-700">Active Funnels</p>
                  </div>
                  <div className="text-center p-6 bg-purple-50 rounded-lg">
                    <Workflow className="w-12 h-12 text-purple-600 mx-auto mb-2" />
                    <p className="text-2xl font-bold text-purple-900">
                      {stats ? formatNumber(stats.total_workflows) : '---'}
                    </p>
                    <p className="text-sm text-purple-700">Workflows</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
