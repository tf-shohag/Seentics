import express from 'express';
import { sendSupportEmail } from '../controllers/supportController.js';

const router = express.Router();

// Send support email (handled by gateway auth)
router.post('/contact', sendSupportEmail);

export default router;
