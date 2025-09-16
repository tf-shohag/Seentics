'use client';

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/stores/useAuthStore';
import api from '@/lib/api';

export interface UsageStatus {
  current: number;
  limit: number;
  canCreate: boolean;
}

export interface SubscriptionUsage {
  websites: UsageStatus;
  workflows: UsageStatus;
  funnels: UsageStatus;
  monthlyEvents: UsageStatus;
}

export interface SubscriptionData {
  id: string;
  plan: 'free' | 'standard' | 'pro';
  status: string;
  usage: SubscriptionUsage;
  features: string[];
  isActive: boolean;
  currentPeriodEnd?: string;
}

export interface UseSubscriptionReturn {
  subscription: SubscriptionData | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
  canCreateWebsite: boolean;
  canCreateWorkflow: boolean;
  canCreateFunnel: boolean;
  canTrackEvents: (count?: number) => boolean;
  getUsagePercentage: (type: keyof SubscriptionUsage) => number;
  isNearLimit: (type: keyof SubscriptionUsage, threshold?: number) => boolean;
  hasReachedLimit: (type: keyof SubscriptionUsage) => boolean;
}

export const useSubscription = (): UseSubscriptionReturn => {
  const [subscription, setSubscription] = useState<SubscriptionData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { user, isAuthenticated } = useAuth();

  const fetchSubscription = useCallback(async () => {
    if (!isAuthenticated || !user) {
      setSubscription(null);
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await api.get('/user/billing/usage');
      
      if (response.data.success) {
        setSubscription({
          id: user.id,
          plan: response.data.data.plan,
          status: 'active',
          usage: response.data.data.usage,
          features: response.data.data.features || [],
          isActive: true
        });
      } else {
        throw new Error(response.data.message || 'Failed to fetch subscription');
      }
    } catch (err: any) {
      console.error('Error fetching subscription:', err);
      setError(err.response?.data?.message || err.message || 'Unknown error occurred');
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated, user]);

  useEffect(() => {
    fetchSubscription();
  }, [fetchSubscription]);

  // Helper functions
  const canCreateWebsite = subscription?.usage?.websites?.canCreate ?? false;
  const canCreateWorkflow = subscription?.usage?.workflows?.canCreate ?? false;
  const canCreateFunnel = subscription?.usage?.funnels?.canCreate ?? false;
  
  const canTrackEvents = useCallback((count: number = 1): boolean => {
    if (!subscription?.usage?.monthlyEvents) return false;
    const { current, limit } = subscription.usage.monthlyEvents;
    return (current + count) <= limit;
  }, [subscription]);

  const getUsagePercentage = useCallback((type: keyof SubscriptionUsage): number => {
    if (!subscription?.usage?.[type]) return 0;
    const { current, limit } = subscription.usage[type];
    return Math.min((current / limit) * 100, 100);
  }, [subscription]);

  const isNearLimit = useCallback((type: keyof SubscriptionUsage, threshold: number = 80): boolean => {
    return getUsagePercentage(type) >= threshold;
  }, [getUsagePercentage]);

  const hasReachedLimit = useCallback((type: keyof SubscriptionUsage): boolean => {
    if (!subscription?.usage?.[type]) return false;
    const { current, limit } = subscription.usage[type];
    return current >= limit;
  }, [subscription]);

  return {
    subscription,
    loading,
    error,
    refetch: fetchSubscription,
    canCreateWebsite,
    canCreateWorkflow,
    canCreateFunnel,
    canTrackEvents,
    getUsagePercentage,
    isNearLimit,
    hasReachedLimit,
  };
};
