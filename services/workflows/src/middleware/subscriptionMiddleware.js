import axios from 'axios';
import { config } from '../config/config.js';

// Plan limits configuration (mirrored from users service)
const PLAN_LIMITS = {
  free: {
    websites: 1,
    monthlyEvents: 1000,
    workflows: 1,
    funnels: 1,
    features: ['basic_analytics', 'email_support', 'privacy_compliant']
  },
  standard: {
    websites: 5,
    monthlyEvents: 100000,
    workflows: 10,
    funnels: 10,
    features: ['advanced_analytics', 'priority_support', 'custom_domains', 'api_access', 'advanced_integrations', 'ab_testing']
  },
  pro: {
    websites: 10,
    monthlyEvents: 500000,
    workflows: 30,
    funnels: 30,
    features: ['enterprise_analytics', '24_7_support', 'white_label', 'advanced_api', 'custom_integrations', 'advanced_ab_testing', 'team_collaboration', 'custom_reports']
  }
};

// Check subscription usage limits using header data from gateway
export const checkSubscriptionLimit = (limitType) => {
  return async (req, res, next) => {
    try {
      const userId = req.user?._id || req.user?.id || req.headers['x-user-id'];
      const userPlan = req.headers['x-user-plan'] || 'free';
      const userStatus = req.headers['x-user-status'] || 'active';
      
      if (!userId) {
        return res.status(401).json({
          success: false,
          message: 'User authentication required'
        });
      }

      // Check if user is active
      if (userStatus !== 'active') {
        return res.status(403).json({
          success: false,
          message: 'User account is not active'
        });
      }

      // Get plan limits from local configuration
      const planLimits = PLAN_LIMITS[userPlan] || PLAN_LIMITS.free;
      
      // For workflows service, we need to get current usage from users service
      // This is more efficient than getting full subscription data
      try {
        const response = await axios.get(`${config.usersServiceUrl}/api/v1/user/billing/usage`, {
          headers: {
            'X-API-Key': config.globalApiKey,
            'X-User-ID': userId,
            'X-User-Email': req.headers['x-user-email'],
            'X-User-Plan': userPlan,
            'X-User-Status': userStatus
          }
        });

        if (!response.data.success) {
          // Fail open for development - allow creation if API fails
          console.warn('Unable to verify usage limits, allowing creation');
          req.userPlan = userPlan;
          return next();
        }

        const usage = response.data.data.usage;
        const currentUsage = usage[limitType]?.current || 0;
        const limit = planLimits[limitType];

        // Check if user can create more of this resource
        const canCreate = currentUsage < limit;

        if (!canCreate) {
          return res.status(403).json({
            success: false,
            message: `${limitType.charAt(0).toUpperCase() + limitType.slice(1)} limit reached for your current plan`,
            error: 'LIMIT_REACHED',
            data: {
              currentPlan: userPlan,
              currentUsage: currentUsage,
              limit: limit,
              upgradeRequired: userPlan !== 'pro'
            }
          });
        }

        // Store info in request for later use
        req.subscriptionUsage = usage;
        req.userPlan = userPlan;
        next();
      } catch (apiError) {
        console.error('Usage check API error:', apiError.message);
        // Fail open for development - allow creation if API fails
        console.warn('Users service unavailable, allowing creation');
        req.userPlan = userPlan;
        next();
      }
    } catch (error) {
      console.error('Subscription limit check error:', error);
      return res.status(500).json({
        success: false,
        message: 'Subscription limit check failed'
      });
    }
  };
};

// Increment usage counter after successful creation
export const incrementUsageCounter = (limitType) => {
  return async (req, res, next) => {
    try {
      const userId = req.user?._id || req.user?.id || req.headers['x-user-id'];
      
      if (!userId) {
        return next(); // Skip if no user ID
      }

      // Call users service to increment usage
      const usersServiceUrl = config.usersServiceUrl;
      
      try {
        await axios.post(`${usersServiceUrl}/api/v1/user/billing/usage/increment`, {
          type: limitType,
          count: 1
        }, {
          headers: {
            'X-API-Key': config.globalApiKey,
            'X-User-ID': userId,
            'X-User-Email': req.headers['x-user-email'],
            'X-User-Plan': req.headers['x-user-plan'],
            'X-User-Status': req.headers['x-user-status']
          }
        });
      } catch (apiError) {
        console.error('Usage increment API error:', apiError.message);
        // Don't fail the request if usage increment fails
      }

      next();
    } catch (error) {
      console.error('Usage increment error:', error);
      // Don't fail the request if usage increment fails
      next();
    }
  };
};
