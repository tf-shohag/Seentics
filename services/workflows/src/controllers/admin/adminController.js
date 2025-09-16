import Workflow from '../../models/Workflow.js';

export const getAdminStats = async (req, res) => {
  try {
    const [totalWorkflows, recentWorkflows, activeWorkflows] = await Promise.all([
      Workflow.countDocuments(),
      Workflow.find()
        .sort({ createdAt: -1 })
        .limit(10)
        .select('name websiteId createdAt status triggerCount')
        .lean(),
      Workflow.countDocuments({ status: 'active' })
    ]);

    res.json({
      total: totalWorkflows,
      recent: recentWorkflows,
      active: activeWorkflows
    });
  } catch (error) {
    console.error('Error fetching workflow admin stats:', error);
    res.status(500).json({ error: 'Failed to fetch workflow statistics' });
  }
};
