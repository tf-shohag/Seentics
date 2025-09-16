import express from 'express';
import { handleLemonSqueezyWebhook } from '../controllers/webhooks/lemonSqueezyWebhook.js';

const router = express.Router();

// Lemon Squeezy webhook endpoint
router.post('/lemonsqueezy', express.raw({ type: 'application/json' }), handleLemonSqueezyWebhook);

export default router;
