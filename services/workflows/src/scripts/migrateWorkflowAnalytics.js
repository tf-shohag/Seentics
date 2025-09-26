#!/usr/bin/env node

/**
 * Migration script to move legacy analytics fields to new analytics structure
 * Run this script after deploying the new model changes
 */

import mongoose from 'mongoose';
import { config } from '../config/config.js';
import { logger } from '../utils/logger.js';
import Workflow from '../models/Workflow.js';

async function connectToDatabase() {
  try {
    await mongoose.connect(config.mongoUri);
    logger.info('Connected to MongoDB for migration');
  } catch (error) {
    logger.error('Failed to connect to MongoDB:', error);
    process.exit(1);
  }
}

async function migrateWorkflowAnalytics() {
  logger.info('Starting workflow analytics migration...');
  
  try {
    // Find all workflows that have legacy fields
    const workflows = await Workflow.find({
      $or: [
        { totalTriggers: { $exists: true } },
        { totalCompletions: { $exists: true } },
        { completionRate: { $exists: true } }
      ]
    });

    logger.info(`Found ${workflows.length} workflows to migrate`);

    let migratedCount = 0;
    let skippedCount = 0;

    for (const workflow of workflows) {
      try {
        const updateDoc = {};
        let needsUpdate = false;

        // Migrate legacy fields to analytics structure
        if (workflow.totalTriggers !== undefined || workflow.totalCompletions !== undefined) {
          updateDoc['analytics.totalTriggers'] = workflow.totalTriggers || 0;
          updateDoc['analytics.totalCompletions'] = workflow.totalCompletions || 0;
          updateDoc['analytics.totalRuns'] = workflow.totalTriggers || 0;
          updateDoc['analytics.successfulRuns'] = workflow.totalCompletions || 0;
          updateDoc['analytics.failedRuns'] = 0;
          updateDoc['analytics.averageCompletionTime'] = 0;
          
          // Initialize nodeStats as empty Map if not exists
          if (!workflow.analytics?.nodeStats) {
            updateDoc['analytics.nodeStats'] = new Map();
          }
          
          needsUpdate = true;
        }

        // Remove legacy fields
        const unsetDoc = {};
        if (workflow.totalTriggers !== undefined) {
          unsetDoc.totalTriggers = '';
        }
        if (workflow.totalCompletions !== undefined) {
          unsetDoc.totalCompletions = '';
        }
        if (workflow.completionRate !== undefined) {
          unsetDoc.completionRate = '';
        }

        if (needsUpdate) {
          const updateOperation = { $set: updateDoc };
          if (Object.keys(unsetDoc).length > 0) {
            updateOperation.$unset = unsetDoc;
          }

          await Workflow.findByIdAndUpdate(workflow._id, updateOperation);
          migratedCount++;
          
          if (migratedCount % 100 === 0) {
            logger.info(`Migrated ${migratedCount} workflows...`);
          }
        } else {
          skippedCount++;
        }
      } catch (error) {
        logger.error(`Error migrating workflow ${workflow._id}:`, error);
      }
    }

    logger.info(`Migration completed: ${migratedCount} migrated, ${skippedCount} skipped`);
    
    // Verify migration
    const remainingLegacy = await Workflow.countDocuments({
      $or: [
        { totalTriggers: { $exists: true } },
        { totalCompletions: { $exists: true } },
        { completionRate: { $exists: true } }
      ]
    });

    if (remainingLegacy > 0) {
      logger.warn(`Warning: ${remainingLegacy} workflows still have legacy fields`);
    } else {
      logger.info('✅ All workflows successfully migrated to new analytics structure');
    }

  } catch (error) {
    logger.error('Migration failed:', error);
    throw error;
  }
}

async function cleanupUIProperties() {
  logger.info('Cleaning up UI properties from nodes and edges...');
  
  try {
    // Remove UI properties from nodes
    const nodeCleanupResult = await Workflow.updateMany(
      {},
      {
        $unset: {
          'nodes.$[].selected': '',
          'nodes.$[].dragging': '',
          'nodes.$[].positionAbsolute': '',
          'nodes.$[].measured': '',
          'nodes.$[].resizing': '',
          'nodes.$[].style': '',
          'edges.$[].selected': '',
          'edges.$[].markerEnd': '',
          'edges.$[].markerStart': '',
          'edges.$[].style': ''
        }
      }
    );

    logger.info(`Cleaned UI properties from ${nodeCleanupResult.modifiedCount} workflows`);
  } catch (error) {
    logger.error('Error cleaning UI properties:', error);
  }
}

async function runMigration() {
  try {
    await connectToDatabase();
    await migrateWorkflowAnalytics();
    await cleanupUIProperties();
    
    logger.info('🎉 Migration completed successfully!');
    process.exit(0);
  } catch (error) {
    logger.error('Migration failed:', error);
    process.exit(1);
  }
}

// Handle script execution
if (import.meta.url === `file://${process.argv[1]}`) {
  runMigration();
}

export { migrateWorkflowAnalytics, cleanupUIProperties };
