import express from 'express';
import { authenticate } from '../middleware/auth.js';
import { apiKeyMiddleware } from '../middleware/apiKey.js';
import {
  getSubscription,
  createCheckoutSession,
  cancelSubscription,
  getBillingHistory,
  checkUsageLimits,
  incrementUsage,
  requireCloudFeatures
} from '../controllers/billing/billingController.js';

const router = express.Router();

// All billing routes require cloud features to be enabled
router.use(requireCloudFeatures);

// Get current subscription
router.get('/subscription', authenticate, getSubscription);

// Create checkout session for subscription upgrade
router.post('/checkout', authenticate, createCheckoutSession);

// Cancel subscription
router.post('/subscription/cancel', authenticate, cancelSubscription);

// Get billing history
router.get('/billing-history', authenticate, getBillingHistory);

// Check usage limits
router.get('/usage', authenticate, checkUsageLimits);

// Increment usage counter
router.post('/usage/increment', authenticate, incrementUsage);

export default router;
