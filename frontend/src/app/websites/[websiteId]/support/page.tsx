'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { 
  HelpCircle, 
  Mail, 
  CheckCircle,
  Send,
  MessageSquare,
  Clock,
  Shield,
  Zap
} from 'lucide-react';
import { useAuth } from '@/stores/useAuthStore';
import api from '@/lib/api';

export default function SupportPage() {
  const [formData, setFormData] = useState({
    subject: '',
    message: '',
    email: ''
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const { user } = useAuth();

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  React.useEffect(() => {
    if (user?.email) {
      setFormData(prev => ({ ...prev, email: user.email }));
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.subject.trim() || !formData.message.trim() || !formData.email.trim()) {
      alert('Please fill in all fields');
      return;
    }

    setIsSubmitting(true);
    
    try {
      const response = await api.post('/user/support/contact', {
        subject: formData.subject,
        message: formData.message,
        email: formData.email
      });

      if (response.data.success) {
        setIsSubmitted(true);
        setFormData({ subject: '', message: '', email: user?.email || '' });
      } else {
        throw new Error(response.data.message || 'Failed to send support request');
      }
    } catch (error: any) {
      console.error('Support request error:', error);
      alert(error.response?.data?.message || error.message || 'Failed to send support request');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isSubmitted) {
    return (
      <div className="container mx-auto px-4 py-8 max-w-2xl">
        <Card>
          <CardContent className="p-8 text-center">
            <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
            <h2 className="text-2xl font-bold text-gray-900 mb-2">Thank You!</h2>
            <p className="text-gray-600 mb-6">
              Your support request has been sent successfully. We'll get back to you within 24 hours.
            </p>
            <Button 
              onClick={() => setIsSubmitted(false)}
              className="bg-blue-600 hover:bg-blue-700"
            >
              Send Another Message
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="w-full px-6 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 max-w-7xl mx-auto">
        {/* Contact Form */}
        <div>
          <Card className="shadow-lg h-fit">
            <CardHeader className="bg-gradient-to-r from-blue-600 to-blue-700 text-white rounded-t-lg">
              <CardTitle className="flex items-center gap-2 text-xl">
                <Mail className="w-6 h-6" />
                Send us a Message
              </CardTitle>
              <CardDescription className="text-blue-100">
                Get personalized help from our support team
              </CardDescription>
            </CardHeader>
            <CardContent className="p-8">
              <form onSubmit={handleSubmit} className="space-y-6">
                <div>
                  <Label htmlFor="email" className="text-sm font-medium text-gray-700">Email Address</Label>
                  <Input
                    id="email"
                    type="email"
                    value={formData.email}
                    onChange={(e) => handleInputChange('email', e.target.value)}
                    placeholder="your@email.com"
                    className="mt-1"
                    required
                  />
                </div>

                <div>
                  <Label htmlFor="subject" className="text-sm font-medium text-gray-700">Subject</Label>
                  <Input
                    id="subject"
                    value={formData.subject}
                    onChange={(e) => handleInputChange('subject', e.target.value)}
                    placeholder="What can we help you with?"
                    className="mt-1"
                    required
                  />
                </div>

                <div>
                  <Label htmlFor="message" className="text-sm font-medium text-gray-700">Message</Label>
                  <Textarea
                    id="message"
                    value={formData.message}
                    onChange={(e) => handleInputChange('message', e.target.value)}
                    placeholder="Please describe your question or issue in detail. Include any error messages, steps you've taken, and what you were trying to accomplish."
                    rows={8}
                    className="mt-1"
                    required
                  />
                </div>

                <Button 
                  type="submit" 
                  disabled={isSubmitting}
                  className="w-full bg-blue-600 hover:bg-blue-700 py-3 text-lg"
                >
                  {isSubmitting ? (
                    <>
                      <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
                      Sending Message...
                    </>
                  ) : (
                    <>
                      <Send className="w-5 h-5 mr-2" />
                      Send Message
                    </>
                  )}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>

        {/* Emergency Contact */}
        <div>
          <Card className="shadow-lg h-fit">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-xl">
                <Mail className="w-6 h-6 text-red-600" />
                Emergency Contact
              </CardTitle>
              <CardDescription>
                For urgent issues that need immediate attention
              </CardDescription>
            </CardHeader>
            <CardContent className="p-8">
              <div className="space-y-6">
                <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                  <h3 className="font-semibold text-red-900 mb-3">Critical Issues</h3>
                  <p className="text-red-700 text-sm mb-4">
                    If you're experiencing a critical issue that affects your service availability or data integrity, contact us immediately.
                  </p>
                  <div className="space-y-3">
                    <div className="flex items-center gap-3">
                      <Mail className="w-5 h-5 text-red-600" />
                      <div>
                        <p className="font-medium text-red-900">Email</p>
                        <p className="text-red-700 text-sm">shohagmiah2100@gmail.com</p>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
                  <h3 className="font-semibold text-blue-900 mb-3">Response Times</h3>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-blue-700 text-sm">Critical Issues:</span>
                      <span className="font-medium text-blue-900 text-sm">Within 2 hours</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-blue-700 text-sm">General Support:</span>
                      <span className="font-medium text-blue-900 text-sm">Within 24 hours</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-blue-700 text-sm">Feature Requests:</span>
                      <span className="font-medium text-blue-900 text-sm">Within 48 hours</span>
                    </div>
                  </div>
                </div>

                <div className="bg-gray-50 border border-gray-200 rounded-lg p-6">
                  <h3 className="font-semibold text-gray-900 mb-3">Before Contacting</h3>
                  <ul className="space-y-2 text-sm text-gray-700">
                    <li className="flex items-start gap-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full mt-2"></div>
                      <span>Include error messages and screenshots</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full mt-2"></div>
                      <span>Describe steps to reproduce the issue</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full mt-2"></div>
                      <span>Mention your website URL if relevant</span>
                    </li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
