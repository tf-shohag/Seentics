import express from 'express';
import { getAdminStats, getUsersList, getWebsitesList } from '../controllers/admin/adminController.js';

const router = express.Router();

// Admin routes for dashboard statistics
router.get('/stats', getAdminStats);
router.get('/users', getUsersList);
router.get('/websites', getWebsitesList);
router.get('/websites/stats', getWebsitesList); // Alias for compatibility

export default router;
