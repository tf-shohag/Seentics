import mongoose from 'mongoose';

const websiteSchema = new mongoose.Schema({
  userId: {
    type: mongoose.Schema.Types.ObjectId,
    ref: 'User',
    required: true
  },
  
  name: {
    type: String,
    required: true,
    trim: true
  },
  url: {
    type: String,
    required: true,
    trim: true
  },
  domain: {
    type: String,
    required: true,
    trim: true
  },
  siteId: {
    type: String,
    required: true
  },
  
  // Verification
  isVerified: {
    type: Boolean,
    default: false
  },
  verificationToken: {
    type: String,
    default: null
  },
  
  // Settings
  settings: {
    allowedOrigins: [{
      type: String
    }],
    trackingEnabled: {
      type: Boolean,
      default: true
    },
    dataRetentionDays: {
      type: Number,
      default: 90
    }
  },
  
  // Statistics
  stats: {
    totalPageviews: {
      type: Number,
      default: 0
    },
    uniqueVisitors: {
      type: Number,
      default: 0
    },
    averageSessionDuration: {
      type: Number,
      default: 0
    },
    bounceRate: {
      type: Number,
      default: 0
    }
  },
  
  // Status
  isActive: {
    type: Boolean,
    default: true
  }
}, {
  timestamps: true
});

// Index for faster queries
websiteSchema.index({ userId: 1, isActive: 1 });
websiteSchema.index({ domain: 1 });
websiteSchema.index({ siteId: 1 }, { unique: true });

const Website = mongoose.model('Website', websiteSchema);
export default Website;