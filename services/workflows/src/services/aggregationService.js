import cron from 'node-cron';
import { WorkflowEvent, DailyAggregation, WeeklyAggregation, MonthlyAggregation } from '../models/WorkflowEvent.js';
import Workflow from '../models/Workflow.js';
import { logger } from '../utils/logger.js';

class AggregationService {
  constructor() {
    this.isRunning = false;
  }

  // Start all cron jobs
  start() {
    logger.info('Starting aggregation service cron jobs');

    // Daily aggregation - runs every day at 1 AM
    cron.schedule('0 1 * * *', async () => {
      await this.runDailyAggregation();
    });

    // Weekly aggregation - runs every Monday at 2 AM
    cron.schedule('0 2 * * 1', async () => {
      await this.runWeeklyAggregation();
    });

    // Monthly aggregation - runs on 1st of every month at 3 AM
    cron.schedule('0 3 1 * *', async () => {
      await this.runMonthlyAggregation();
    });

    // Cleanup old events - runs every hour
    cron.schedule('0 * * * *', async () => {
      await this.cleanupOldEvents();
    });

    logger.info('Aggregation service cron jobs started');
  }

  // Run daily aggregation for yesterday's data
  async runDailyAggregation() {
    if (this.isRunning) {
      logger.warn('Daily aggregation already running, skipping');
      return;
    }

    this.isRunning = true;
    logger.info('Starting daily aggregation');

    try {
      const yesterday = new Date();
      yesterday.setDate(yesterday.getDate() - 1);
      yesterday.setHours(0, 0, 0, 0);

      const today = new Date(yesterday);
      today.setDate(today.getDate() + 1);

      // Get all workflows that had events yesterday
      const workflowsWithEvents = await WorkflowEvent.distinct('workflowId', {
        timestamp: { $gte: yesterday, $lt: today }
      });

      for (const workflowId of workflowsWithEvents) {
        await this.aggregateDailyWorkflowData(workflowId, yesterday);
      }

      logger.info(`Daily aggregation completed for ${workflowsWithEvents.length} workflows`);
    } catch (error) {
      logger.error('Error in daily aggregation:', error);
    } finally {
      this.isRunning = false;
    }
  }

  // Aggregate daily data for a specific workflow
  async aggregateDailyWorkflowData(workflowId, date) {
    try {
      const nextDay = new Date(date);
      nextDay.setDate(nextDay.getDate() + 1);

      // Get workflow info
      const workflow = await Workflow.findById(workflowId);
      if (!workflow) {
        logger.warn(`Workflow ${workflowId} not found for aggregation`);
        return;
      }

      // Aggregate events for this workflow and date
      const aggregation = await WorkflowEvent.aggregate([
        {
          $match: {
            workflowId: workflowId,
            timestamp: { $gte: date, $lt: nextDay }
          }
        },
        {
          $group: {
            _id: {
              workflowId: '$workflowId',
              event: '$event'
            },
            count: { $sum: 1 },
            uniqueVisitors: { $addToSet: '$visitorId' },
            totalExecutionTime: { $sum: '$executionTime' },
            nodeEvents: {
              $push: {
                nodeId: '$nodeId',
                nodeTitle: '$nodeTitle',
                event: '$event',
                success: '$success'
              }
            }
          }
        }
      ]);

      // Process aggregation results
      let triggers = 0;
      let completions = 0;
      let totalExecutionTime = 0;
      const uniqueVisitors = new Set();
      const nodePerformance = new Map();

      for (const agg of aggregation) {
        const eventType = agg._id.event;
        const count = agg.count;

        // Add unique visitors
        agg.uniqueVisitors.forEach(visitor => uniqueVisitors.add(visitor));

        // Sum execution time
        totalExecutionTime += agg.totalExecutionTime || 0;

        // Count event types
        if (eventType === 'Trigger' || eventType === 'workflow_trigger') {
          triggers += count;
        } else if (eventType === 'Action Executed' || eventType === 'action_completed') {
          completions += count;
        }

        // Process node performance
        for (const nodeEvent of agg.nodeEvents) {
          const nodeId = nodeEvent.nodeId;
          if (!nodePerformance.has(nodeId)) {
            nodePerformance.set(nodeId, {
              nodeId,
              nodeTitle: nodeEvent.nodeTitle,
              triggers: 0,
              completions: 0,
              successRate: 0
            });
          }

          const nodeStats = nodePerformance.get(nodeId);
          if (nodeEvent.event === 'Trigger') {
            nodeStats.triggers++;
          } else if (nodeEvent.event === 'Action Executed') {
            nodeStats.completions++;
            if (nodeEvent.success) {
              nodeStats.successRate++;
            }
          }
        }
      }

      // Calculate metrics
      const conversionRate = triggers > 0 ? (completions / triggers) * 100 : 0;
      const avgExecutionTime = completions > 0 ? totalExecutionTime / completions : 0;

      // Update node success rates
      for (const [nodeId, stats] of nodePerformance) {
        if (stats.completions > 0) {
          stats.successRate = (stats.successRate / stats.completions) * 100;
        }
      }

      // Save or update daily aggregation
      await DailyAggregation.findOneAndUpdate(
        {
          workflowId: workflowId,
          date: date
        },
        {
          workflowId: workflowId,
          siteId: workflow.siteId,
          date: date,
          triggers,
          completions,
          conversionRate,
          totalExecutionTime,
          avgExecutionTime,
          uniqueVisitors: uniqueVisitors.size,
          uniqueSessions: uniqueVisitors.size, // Approximation
          nodePerformance: Array.from(nodePerformance.values()),
          updatedAt: new Date()
        },
        { upsert: true, new: true }
      );

      logger.debug(`Daily aggregation saved for workflow ${workflowId} on ${date.toISOString().split('T')[0]}`);
    } catch (error) {
      logger.error(`Error aggregating daily data for workflow ${workflowId}:`, error);
    }
  }

  // Run weekly aggregation
  async runWeeklyAggregation() {
    logger.info('Starting weekly aggregation');
    
    try {
      const lastWeek = new Date();
      lastWeek.setDate(lastWeek.getDate() - 7);
      const weekStart = new Date(lastWeek);
      weekStart.setDate(weekStart.getDate() - weekStart.getDay()); // Start of week (Sunday)
      weekStart.setHours(0, 0, 0, 0);
      
      const weekEnd = new Date(weekStart);
      weekEnd.setDate(weekEnd.getDate() + 6);
      weekEnd.setHours(23, 59, 59, 999);

      // Aggregate from daily data
      const weeklyData = await DailyAggregation.aggregate([
        {
          $match: {
            date: { $gte: weekStart, $lte: weekEnd }
          }
        },
        {
          $group: {
            _id: {
              workflowId: '$workflowId',
              siteId: '$siteId'
            },
            totalTriggers: { $sum: '$triggers' },
            totalCompletions: { $sum: '$completions' },
            totalExecutionTime: { $sum: '$totalExecutionTime' },
            totalUniqueVisitors: { $sum: '$uniqueVisitors' },
            totalUniqueSessions: { $sum: '$uniqueSessions' },
            dailyBreakdown: {
              $push: {
                date: '$date',
                triggers: '$triggers',
                completions: '$completions',
                conversionRate: '$conversionRate'
              }
            }
          }
        }
      ]);

      for (const data of weeklyData) {
        const avgConversionRate = data.totalTriggers > 0 ? 
          (data.totalCompletions / data.totalTriggers) * 100 : 0;
        const avgExecutionTime = data.totalCompletions > 0 ? 
          data.totalExecutionTime / data.totalCompletions : 0;

        await WeeklyAggregation.findOneAndUpdate(
          {
            workflowId: data._id.workflowId,
            weekStart: weekStart
          },
          {
            workflowId: data._id.workflowId,
            siteId: data._id.siteId,
            weekStart,
            weekEnd,
            weekNumber: this.getWeekNumber(weekStart),
            year: weekStart.getFullYear(),
            totalTriggers: data.totalTriggers,
            totalCompletions: data.totalCompletions,
            avgConversionRate,
            totalExecutionTime: data.totalExecutionTime,
            avgExecutionTime,
            totalUniqueVisitors: data.totalUniqueVisitors,
            totalUniqueSessions: data.totalUniqueSessions,
            dailyBreakdown: data.dailyBreakdown,
            updatedAt: new Date()
          },
          { upsert: true, new: true }
        );
      }

      logger.info(`Weekly aggregation completed for ${weeklyData.length} workflows`);
    } catch (error) {
      logger.error('Error in weekly aggregation:', error);
    }
  }

  // Run monthly aggregation
  async runMonthlyAggregation() {
    logger.info('Starting monthly aggregation');
    
    try {
      const lastMonth = new Date();
      lastMonth.setMonth(lastMonth.getMonth() - 1);
      const monthStart = new Date(lastMonth.getFullYear(), lastMonth.getMonth(), 1);
      const monthEnd = new Date(lastMonth.getFullYear(), lastMonth.getMonth() + 1, 0);
      monthEnd.setHours(23, 59, 59, 999);

      // Aggregate from weekly data
      const monthlyData = await WeeklyAggregation.aggregate([
        {
          $match: {
            weekStart: { $gte: monthStart, $lte: monthEnd }
          }
        },
        {
          $group: {
            _id: {
              workflowId: '$workflowId',
              siteId: '$siteId'
            },
            totalTriggers: { $sum: '$totalTriggers' },
            totalCompletions: { $sum: '$totalCompletions' },
            totalExecutionTime: { $sum: '$totalExecutionTime' },
            totalUniqueVisitors: { $sum: '$totalUniqueVisitors' },
            totalUniqueSessions: { $sum: '$totalUniqueSessions' },
            weeklyBreakdown: {
              $push: {
                weekStart: '$weekStart',
                weekEnd: '$weekEnd',
                triggers: '$totalTriggers',
                completions: '$totalCompletions',
                conversionRate: '$avgConversionRate'
              }
            }
          }
        }
      ]);

      for (const data of monthlyData) {
        const avgConversionRate = data.totalTriggers > 0 ? 
          (data.totalCompletions / data.totalTriggers) * 100 : 0;
        const avgExecutionTime = data.totalCompletions > 0 ? 
          data.totalExecutionTime / data.totalCompletions : 0;

        await MonthlyAggregation.findOneAndUpdate(
          {
            workflowId: data._id.workflowId,
            month: monthStart.getMonth() + 1,
            year: monthStart.getFullYear()
          },
          {
            workflowId: data._id.workflowId,
            siteId: data._id.siteId,
            month: monthStart.getMonth() + 1,
            year: monthStart.getFullYear(),
            monthStart,
            monthEnd,
            totalTriggers: data.totalTriggers,
            totalCompletions: data.totalCompletions,
            avgConversionRate,
            totalExecutionTime: data.totalExecutionTime,
            avgExecutionTime,
            totalUniqueVisitors: data.totalUniqueVisitors,
            totalUniqueSessions: data.totalUniqueSessions,
            weeklyBreakdown: data.weeklyBreakdown,
            updatedAt: new Date()
          },
          { upsert: true, new: true }
        );
      }

      logger.info(`Monthly aggregation completed for ${monthlyData.length} workflows`);
    } catch (error) {
      logger.error('Error in monthly aggregation:', error);
    }
  }

  // Clean up old events (older than 24 hours)
  async cleanupOldEvents() {
    try {
      const cutoffTime = new Date();
      cutoffTime.setHours(cutoffTime.getHours() - 24);

      const result = await WorkflowEvent.deleteMany({
        timestamp: { $lt: cutoffTime }
      });

      if (result.deletedCount > 0) {
        logger.info(`Cleaned up ${result.deletedCount} old workflow events`);
      }
    } catch (error) {
      logger.error('Error cleaning up old events:', error);
    }
  }

  // Helper function to get week number
  getWeekNumber(date) {
    const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
    const dayNum = d.getUTCDay() || 7;
    d.setUTCDate(d.getUTCDate() + 4 - dayNum);
    const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
    return Math.ceil((((d - yearStart) / 86400000) + 1) / 7);
  }

  // Manual trigger for testing
  async runManualAggregation(type = 'daily') {
    logger.info(`Running manual ${type} aggregation`);
    
    switch (type) {
      case 'daily':
        await this.runDailyAggregation();
        break;
      case 'weekly':
        await this.runWeeklyAggregation();
        break;
      case 'monthly':
        await this.runMonthlyAggregation();
        break;
      default:
        throw new Error(`Unknown aggregation type: ${type}`);
    }
  }
}

export default new AggregationService();
