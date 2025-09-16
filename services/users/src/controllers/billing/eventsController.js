import User from '../../models/User.js';
import Subscription from '../../models/Subscription.js';

// Handle batch events sync from gateway
export const syncEventsBatch = async (req, res) => {
  try {
    const { events } = req.body;
    
    if (!events || !Array.isArray(events)) {
      return res.status(400).json({
        success: false,
        message: 'Invalid events data format'
      });
    }

    const results = [];
    const errors = [];

    // Process each user's events
    for (const eventData of events) {
      try {
        const { user_id, count, month_year } = eventData;
        
        if (!user_id || !count || !month_year) {
          errors.push({
            user_id: user_id || 'unknown',
            error: 'Missing required fields'
          });
          continue;
        }

        // Find user and subscription
        const user = await User.findById(user_id).populate('subscriptionId');
        if (!user || !user.subscriptionId) {
          errors.push({
            user_id,
            error: 'User or subscription not found'
          });
          continue;
        }

        const subscription = user.subscriptionId;
        
        // Update monthly events usage
        subscription.currentUsage.monthlyEvents += count;
        
        // Update last usage timestamp
        subscription.lastUsageUpdate = new Date();
        
        await subscription.save();
        
        results.push({
          user_id,
          events_added: count,
          total_events: subscription.currentUsage.monthlyEvents,
          month_year
        });

      } catch (error) {
        console.error(`Error processing events for user ${eventData.user_id}:`, error);
        errors.push({
          user_id: eventData.user_id || 'unknown',
          error: error.message
        });
      }
    }

    res.status(200).json({
      success: true,
      message: `Processed ${results.length} users successfully`,
      data: {
        processed: results.length,
        errors: errors.length,
        results,
        errors
      }
    });

  } catch (error) {
    console.error('Batch events sync error:', error);
    res.status(500).json({
      success: false,
      message: 'Failed to sync events batch',
      error: error.message
    });
  }
};

// Get current monthly events usage for a user
export const getUserEventsUsage = async (req, res) => {
  try {
    const userId = req.userId || req.params.userId;
    
    if (!userId) {
      return res.status(400).json({
        success: false,
        message: 'User ID required'
      });
    }

    const user = await User.findById(userId).populate('subscriptionId');
    if (!user || !user.subscriptionId) {
      return res.status(404).json({
        success: false,
        message: 'User or subscription not found'
      });
    }

    const subscription = user.subscriptionId;
    const planLimits = subscription.getPlanLimits();
    
    res.status(200).json({
      success: true,
      data: {
        user_id: userId,
        plan: subscription.plan,
        current_usage: subscription.currentUsage.monthlyEvents,
        limit: planLimits.monthlyEvents,
        remaining: Math.max(0, planLimits.monthlyEvents - subscription.currentUsage.monthlyEvents),
        percentage_used: Math.round((subscription.currentUsage.monthlyEvents / planLimits.monthlyEvents) * 100),
        last_updated: subscription.lastUsageUpdate
      }
    });

  } catch (error) {
    console.error('Get events usage error:', error);
    res.status(500).json({
      success: false,
      message: 'Failed to get events usage',
      error: error.message
    });
  }
};

// Reset monthly events counter (called at month start)
export const resetMonthlyEvents = async (req, res) => {
  try {
    const { user_ids } = req.body;
    
    if (!user_ids || !Array.isArray(user_ids)) {
      return res.status(400).json({
        success: false,
        message: 'user_ids array required'
      });
    }

    const results = [];
    const errors = [];

    for (const userId of user_ids) {
      try {
        const user = await User.findById(userId).populate('subscriptionId');
        if (!user || !user.subscriptionId) {
          errors.push({
            user_id: userId,
            error: 'User or subscription not found'
          });
          continue;
        }

        const subscription = user.subscriptionId;
        subscription.resetMonthlyUsage();
        await subscription.save();
        
        results.push({
          user_id: userId,
          reset: true
        });

      } catch (error) {
        console.error(`Error resetting events for user ${userId}:`, error);
        errors.push({
          user_id: userId,
          error: error.message
        });
      }
    }

    res.status(200).json({
      success: true,
      message: `Reset ${results.length} users successfully`,
      data: {
        reset: results.length,
        errors: errors.length,
        results,
        errors
      }
    });

  } catch (error) {
    console.error('Reset monthly events error:', error);
    res.status(500).json({
      success: false,
      message: 'Failed to reset monthly events',
      error: error.message
    });
  }
};
