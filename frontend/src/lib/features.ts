/**
 * Feature flags for Open Source vs Cloud deployments
 * 
 * CLOUD_FEATURES_ENABLED=true  -> Full SaaS features (billing, support, usage limits)
 * CLOUD_FEATURES_ENABLED=false -> Open Source features only
 */

// Check if cloud features are enabled
export const CLOUD_FEATURES_ENABLED = 
  process.env.NEXT_PUBLIC_CLOUD_FEATURES_ENABLED === 'true';

// Helper functions
export const isOpenSource = () => !CLOUD_FEATURES_ENABLED;
export const isCloudVersion = () => CLOUD_FEATURES_ENABLED;

// Feature configuration
export const FEATURES = {
  // Core features (always available in both OSS and Cloud)
  ANALYTICS_DASHBOARD: true,
  EVENT_TRACKING: true,
  BASIC_REPORTS: true,
  PRIVACY_SETTINGS: true,
  WEBSITE_MANAGEMENT: true,
  WORKFLOW_BASIC: true,
  FUNNEL_BASIC: true,
  
  // Cloud-only features (hidden in OSS)
  BILLING_PAGE: CLOUD_FEATURES_ENABLED,
  SUBSCRIPTION_MANAGEMENT: CLOUD_FEATURES_ENABLED,
  USAGE_LIMITS_UI: CLOUD_FEATURES_ENABLED,
  TEAM_MANAGEMENT: CLOUD_FEATURES_ENABLED,
  SUPPORT_CHAT: CLOUD_FEATURES_ENABLED,
  ADVANCED_INTEGRATIONS: CLOUD_FEATURES_ENABLED,
  WHITE_LABEL_OPTIONS: CLOUD_FEATURES_ENABLED,
  STRIPE_INTEGRATION: CLOUD_FEATURES_ENABLED,
  USAGE_ANALYTICS: CLOUD_FEATURES_ENABLED,
  
  // OSS gets unlimited everything, Cloud has limits
  UNLIMITED_WEBSITES: isOpenSource(),
  UNLIMITED_EVENTS: isOpenSource(),
  UNLIMITED_WORKFLOWS: isOpenSource(),
  UNLIMITED_FUNNELS: isOpenSource(),
} as const;

// Check if a specific feature is enabled
export const hasFeature = (feature: keyof typeof FEATURES): boolean => {
  return FEATURES[feature] === true;
};

// Usage limits for display
export const LIMITS = {
  OSS: {
    websites: '∞',
    events: '∞ events/month',
    workflows: '∞',
    funnels: '∞',
    teamMembers: '∞',
    dataRetention: '∞',
  },
  CLOUD: {
    free: {
      websites: '1',
      events: '10K events/month',
      workflows: '1',
      funnels: '1',
      teamMembers: '1',
      dataRetention: '30 days',
    },
    pro: {
      websites: '10',
      events: '1M events/month',
      workflows: '50',
      funnels: '25',
      teamMembers: '5',
      dataRetention: '1 year',
    },
    enterprise: {
      websites: '∞',
      events: '∞ events/month',
      workflows: '∞',
      funnels: '∞',
      teamMembers: '∞',
      dataRetention: '∞',
    }
  }
};

// Get current limits based on deployment type
export const getCurrentLimits = () => {
  if (isOpenSource()) {
    return LIMITS.OSS;
  }
  // For cloud version, this would come from user's subscription
  // Default to free tier for now
  return LIMITS.CLOUD.free;
};

// Navigation items that should be hidden in OSS
export const CLOUD_ONLY_NAV_ITEMS = [
  'billing',
  'subscription',
  'team',
  'support',
  'integrations'
];

// Check if a navigation item should be shown
export const shouldShowNavItem = (itemKey: string): boolean => {
  if (CLOUD_ONLY_NAV_ITEMS.includes(itemKey)) {
    return CLOUD_FEATURES_ENABLED;
  }
  return true;
};
