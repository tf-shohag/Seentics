import mongoose from 'mongoose';

// Subscription plans enum
export const SUBSCRIPTION_PLANS = {
  FREE: 'free',
  STANDARD: 'standard',
  PRO: 'pro'
};

// Plan limits configuration
export const PLAN_LIMITS = {
  [SUBSCRIPTION_PLANS.FREE]: {
    websites: 1,
    monthlyEvents: 1000,
    workflows: 1,
    funnels: 1,
    features: ['basic_analytics', 'email_support', 'privacy_compliant']
  },
  [SUBSCRIPTION_PLANS.STANDARD]: {
    websites: 5,
    monthlyEvents: 100000,
    workflows: 10,
    funnels: 10,
    features: ['advanced_analytics', 'priority_support', 'custom_domains', 'api_access', 'advanced_integrations', 'ab_testing']
  },
  [SUBSCRIPTION_PLANS.PRO]: {
    websites: 10,
    monthlyEvents: 500000,
    workflows: 30,
    funnels: 30,
    features: ['enterprise_analytics', '24_7_support', 'white_label', 'advanced_api', 'custom_integrations', 'advanced_ab_testing', 'team_collaboration', 'custom_reports']
  }
};

const subscriptionSchema = new mongoose.Schema({
  userId: {
    type: mongoose.Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    unique: true
  },
  
  // Current subscription details
  plan: {
    type: String,
    enum: Object.values(SUBSCRIPTION_PLANS),
    default: SUBSCRIPTION_PLANS.FREE,
    required: true
  },
  
  status: {
    type: String,
    enum: ['active', 'cancelled', 'expired', 'past_due', 'unpaid'],
    default: 'active',
    required: true
  },
  
  // Lemon Squeezy integration fields
  lemonSqueezyCustomerId: {
    type: String,
    default: null
  },
  
  lemonSqueezySubscriptionId: {
    type: String,
    default: null
  },
  
  lemonSqueezyVariantId: {
    type: String,
    default: null
  },
  
  // Billing details
  currentPeriodStart: {
    type: Date,
    default: Date.now
  },
  
  currentPeriodEnd: {
    type: Date,
    default: function() {
      // Free plan never expires, others default to 1 month
      if (this.plan === SUBSCRIPTION_PLANS.FREE) {
        return null;
      }
      const date = new Date();
      date.setMonth(date.getMonth() + 1);
      return date;
    }
  },
  
  // Usage tracking
  currentUsage: {
    websites: {
      type: Number,
      default: 0
    },
    monthlyEvents: {
      type: Number,
      default: 0
    },
    workflows: {
      type: Number,
      default: 0
    },
    funnels: {
      type: Number,
      default: 0
    }
  },
  
  // Reset monthly usage tracking
  lastUsageReset: {
    type: Date,
    default: Date.now
  },
  
  // Billing history
  invoices: [{
    lemonSqueezyInvoiceId: String,
    amount: Number,
    currency: String,
    status: String,
    paidAt: Date,
    createdAt: {
      type: Date,
      default: Date.now
    }
  }],
  
  // Cancellation details
  cancelAtPeriodEnd: {
    type: Boolean,
    default: false
  },
  
  cancelledAt: {
    type: Date,
    default: null
  },
  
  cancellationReason: {
    type: String,
    default: null
  }
}, {
  timestamps: true,
  toJSON: {
    transform: function(doc, ret) {
      ret.id = ret._id;
      delete ret.__v;
      return ret;
    }
  }
});

// Instance methods
subscriptionSchema.methods.getPlanLimits = function() {
  return PLAN_LIMITS[this.plan];
};

subscriptionSchema.methods.canCreateWebsite = function() {
  const limits = this.getPlanLimits();
  return this.currentUsage.websites < limits.websites;
};

subscriptionSchema.methods.canCreateWorkflow = function() {
  const limits = this.getPlanLimits();
  return this.currentUsage.workflows < limits.workflows;
};

subscriptionSchema.methods.canCreateFunnel = function() {
  const limits = this.getPlanLimits();
  return this.currentUsage.funnels < limits.funnels;
};

subscriptionSchema.methods.canTrackEvents = function(eventCount = 1) {
  const limits = this.getPlanLimits();
  return (this.currentUsage.monthlyEvents + eventCount) <= limits.monthlyEvents;
};

subscriptionSchema.methods.hasFeature = function(feature) {
  const limits = this.getPlanLimits();
  return limits.features.includes(feature);
};

subscriptionSchema.methods.isActive = function() {
  return this.status === 'active' && (!this.currentPeriodEnd || this.currentPeriodEnd > new Date());
};

subscriptionSchema.methods.incrementUsage = function(type, count = 1) {
  if (this.currentUsage[type] !== undefined) {
    this.currentUsage[type] += count;
  }
};

subscriptionSchema.methods.resetMonthlyUsage = function() {
  this.currentUsage.monthlyEvents = 0;
  this.lastUsageReset = new Date();
};

// Static methods
subscriptionSchema.statics.createFreeSubscription = function(userId) {
  return new this({
    userId,
    plan: SUBSCRIPTION_PLANS.FREE,
    status: 'active',
    currentPeriodStart: new Date(),
    currentPeriodEnd: null // Free plan never expires
  });
};

subscriptionSchema.statics.findByUserId = function(userId) {
  return this.findOne({ userId });
};

// Middleware to reset monthly usage if needed
subscriptionSchema.pre('save', function(next) {
  const now = new Date();
  const lastReset = new Date(this.lastUsageReset);
  
  // Reset monthly usage if it's a new month
  if (now.getMonth() !== lastReset.getMonth() || now.getFullYear() !== lastReset.getFullYear()) {
    this.resetMonthlyUsage();
  }
  
  next();
});

const Subscription = mongoose.model('Subscription', subscriptionSchema);
export default Subscription;
