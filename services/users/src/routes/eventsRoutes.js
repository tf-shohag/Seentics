import express from 'express';
import { syncEventsBatch, getUserEventsUsage, resetMonthlyEvents } from '../controllers/billing/eventsController.js';
import { authenticate } from '../middleware/auth.js';

const router = express.Router();

// Batch sync endpoint for gateway (requires API key)
router.post('/batch', authenticate, syncEventsBatch);

// Get user events usage
router.get('/usage/:userId?', authenticate, getUserEventsUsage);

// Reset monthly events (admin endpoint)
router.post('/reset', authenticate, resetMonthlyEvents);

export default router;
