import User from '../../models/User.js';
import Website from '../../models/Website.js';

// Get admin statistics for users
export const getAdminStats = async (req, res) => {
  try {
    const [totalUsers, recentUsers, activeUsersToday, totalSubscriptions] = await Promise.all([
      User.countDocuments(),
      User.find()
        .sort({ createdAt: -1 })
        .limit(10)
        .select('email username createdAt subscription lastLogin')
        .lean(),
      User.countDocuments({
        lastLogin: { $gte: new Date(Date.now() - 24 * 60 * 60 * 1000) }
      }),
      User.countDocuments({
        subscription: { $in: ['standard', 'pro'] }
      })
    ]);

    res.json({
      total: totalUsers,
      recent: recentUsers,
      active_today: activeUsersToday,
      total_subscriptions: totalSubscriptions
    });
  } catch (error) {
    console.error('Error fetching admin stats:', error);
    res.status(500).json({ error: 'Failed to fetch admin statistics' });
  }
};

// Get paginated list of users
export const getUsersList = async (req, res) => {
  try {
    const page = parseInt(req.query.page) || 1;
    const limit = parseInt(req.query.limit) || 20;
    const skip = (page - 1) * limit;

    const users = await User.find()
      .sort({ createdAt: -1 })
      .skip(skip)
      .limit(limit)
      .select('email username createdAt subscription lastLogin')
      .lean();

    const totalUsers = await User.countDocuments();
    const totalPages = Math.ceil(totalUsers / limit);

    res.json({
      users,
      pagination: {
        currentPage: page,
        totalPages,
        totalUsers,
        hasNextPage: page < totalPages,
        hasPrevPage: page > 1
      }
    });
  } catch (error) {
    console.error('Error fetching users list:', error);
    res.status(500).json({ error: 'Failed to fetch users list' });
  }
};

// Get websites statistics
export const getWebsitesList = async (req, res) => {
  try {
    const page = parseInt(req.query.page) || 1;
    const limit = parseInt(req.query.limit) || 20;
    const skip = (page - 1) * limit;

    const [websites, totalWebsites, activeWebsites] = await Promise.all([
      Website.find()
        .populate('userId', 'email username')
        .sort({ createdAt: -1 })
        .skip(skip)
        .limit(limit)
        .lean(),
      Website.countDocuments(),
      Website.countDocuments({ status: 'active' })
    ]);

    const totalPages = Math.ceil(totalWebsites / limit);

    // If this is a stats request, return stats format
    if (req.path.includes('/stats')) {
      const recentWebsites = await Website.find()
        .populate('userId', 'email username')
        .sort({ createdAt: -1 })
        .limit(10)
        .lean();

      return res.json({
        total: totalWebsites,
        recent: recentWebsites,
        active: activeWebsites
      });
    }

    res.json({
      websites,
      pagination: {
        currentPage: page,
        totalPages,
        totalWebsites,
        hasNextPage: page < totalPages,
        hasPrevPage: page > 1
      }
    });
  } catch (error) {
    console.error('Error fetching websites:', error);
    res.status(500).json({ error: 'Failed to fetch websites' });
  }
};
