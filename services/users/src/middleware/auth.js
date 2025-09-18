import { verifyToken } from '../utils/jwt.js';
import User from '../models/User.js';

// Simple gateway-only authenticate middleware
export const authenticate = async (req, res, next) => {
  try {
    const userId = req.header('X-User-ID');
    const userEmail = req.header('X-User-Email');
    const userPlan = req.header('X-User-Plan');
    const userStatus = req.header('X-User-Status');
    const websiteId = req.header('X-Website-ID');



    if (!userId) {
      return res.status(401).json({
        success: false,
        message: 'Unauthorized - request must come through gateway'
      });
    }

    // Set user info from gateway headers
    req.userId = userId;
    req.userEmail = userEmail;
    req.userPlan = userPlan || 'free';
    req.userStatus = userStatus || 'active';
    req.websiteId = websiteId;

    // Create user object for backward compatibility
    req.user = {
      _id: userId,
      id: userId,
      email: userEmail,
      plan: userPlan || 'free',
      status: userStatus || 'active'
    };


    next();
  } catch (error) {
    console.error('Gateway authentication error:', error);
    return res.status(500).json({
      success: false,
      message: 'Authentication failed'
    });
  }
};

// Simple authorization middleware for open-source version
export const authorize = (requiredRole = null) => {
  return async (req, res, next) => {
    try {
      // For open-source version, just check if user is authenticated
      // All authenticated users have access
      if (!req.userId) {
        return res.status(401).json({
          success: false,
          message: 'Authentication required'
        });
      }

      next();
    } catch (error) {
      return res.status(500).json({
        success: false,
        message: 'Authorization error'
      });
    }
  };
};

// Check subscription usage limits using header data from gateway
export const checkUsageLimit = (limitType) => {
  return async (req, res, next) => {
    try {
      const userId = req.userId || req.user?.id;
      const userPlan = req.userPlan || req.headers['x-user-plan'] || 'free';
      const userStatus = req.userStatus || req.headers['x-user-status'] || 'active';
      
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

      // Import models here to avoid circular dependency
      const UserModel = (await import('../models/User.js')).default;
      const SubscriptionModel = (await import('../models/Subscription.js')).default;
      const { PLAN_LIMITS } = await import('../models/Subscription.js');

      // Get user with subscription for usage data
      const user = await UserModel.findById(userId).populate('subscriptionId');
      
      if (!user || !user.subscriptionId) {
        return res.status(404).json({
          success: false,
          message: 'User subscription not found'
        });
      }

      const subscription = user.subscriptionId;
      
      // Use plan from header if available, fallback to subscription plan
      const currentPlan = userPlan || subscription.plan;
      const planLimits = PLAN_LIMITS[currentPlan] || PLAN_LIMITS.free;

      // Check specific limit based on type
      let canCreate = false;
      let limitMessage = '';
      let currentUsage = 0;
      let limit = 0;

      switch (limitType) {
        case 'websites':
          currentUsage = subscription.currentUsage.websites;
          limit = planLimits.websites;
          canCreate = currentUsage < limit;
          limitMessage = 'Website limit reached for your current plan';
          break;
        case 'workflows':
          currentUsage = subscription.currentUsage.workflows;
          limit = planLimits.workflows;
          canCreate = currentUsage < limit;
          limitMessage = 'Workflow limit reached for your current plan';
          break;
        case 'funnels':
          currentUsage = subscription.currentUsage.funnels;
          limit = planLimits.funnels;
          canCreate = currentUsage < limit;
          limitMessage = 'Funnel limit reached for your current plan';
          break;
        default:
          return res.status(400).json({
            success: false,
            message: 'Invalid limit type'
          });
      }

      if (!canCreate) {
        return res.status(403).json({
          success: false,
          message: limitMessage,
          error: 'LIMIT_REACHED',
          data: {
            currentPlan: currentPlan,
            currentUsage: currentUsage,
            limit: limit,
            upgradeRequired: currentPlan !== 'pro'
          }
        });
      }

      // Store subscription in request for later use
      req.subscription = subscription;
      req.userPlan = currentPlan;
      next();
    } catch (error) {
      console.error('Usage limit check error:', error);
      return res.status(500).json({
        success: false,
        message: 'Usage limit check failed'
      });
    }
  };
};