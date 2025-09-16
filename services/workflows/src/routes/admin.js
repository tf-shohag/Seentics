import express from 'express';
import { getAdminStats } from '../controllers/admin/adminController.js';

const router = express.Router();

// Admin routes for workflow statistics
router.get('/stats', getAdminStats);

export default router;
