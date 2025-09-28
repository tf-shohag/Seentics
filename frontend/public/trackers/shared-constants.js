// --- Seentics Shared Constants ---
// Shared constants and utilities between main tracker and workflow tracker

(function() {
  'use strict';

  // --- Shared Storage Keys ---
  const STORAGE_KEYS = {
    VISITOR_ID: 'seentics_visitor_id',
    SESSION_ID: 'seentics_session_id',
    SESSION_LAST_SEEN: 'seentics_session_last_seen',
    RETURNING_VISITOR: 'seentics_returning'
  };

  // --- Shared Time Constants ---
  const TIME_CONSTANTS = {
    SESSION_EXPIRY_MS: 1800000, // 30 minutes
    VISITOR_EXPIRY_MS: 2592000000, // 30 days
    BATCH_DELAY: 100,
    RETRY_DELAY: 1000,
    MAX_RETRY_ATTEMPTS: 3
  };

  // --- Shared Event Types ---
  const EVENT_TYPES = {
    PAGEVIEW: 'pageview',
    SESSION_START: 'session_start',
    SESSION_END: 'session_end',
    CUSTOM: 'custom',
    WORKFLOW: 'wf'
  };

  // --- Shared Workflow Constants ---
  const WORKFLOW_TRIGGERS = {
    PAGE_VIEW: 'Page View',
    ELEMENT_CLICK: 'Element Click',
    FUNNEL: 'Funnel',
    TIME_SPENT: 'Time Spent',
    EXIT_INTENT: 'Exit Intent'
  };

  const WORKFLOW_ACTIONS = {
    SHOW_MODAL: 'Show Modal',
    SHOW_BANNER: 'Show Banner',
    SHOW_NOTIFICATION: 'Show Notification',
    TRACK_EVENT: 'Track Event',
    WEBHOOK: 'Webhook',
    REDIRECT_URL: 'Redirect URL'
  };

  const WORKFLOW_CONDITIONS = {
    URL_PATH: 'URL Path',
    TRAFFIC_SOURCE: 'Traffic Source',
    NEW_VS_RETURNING: 'New vs Returning',
    DEVICE: 'Device Type'
  };

  const FREQUENCY_TYPES = {
    EVERY_TRIGGER: 'every_trigger',
    ONCE_PER_SESSION: 'once_per_session',
    ONCE_EVER: 'once_ever'
  };

  // --- Shared Utilities ---
  const SharedUtils = {
    // Safe localStorage operations
    safeLocalStorage: {
      get: (key) => {
        try {
          return localStorage.getItem(key);
        } catch (error) {
          console.warn('[Seentics] LocalStorage read error:', error);
          return null;
        }
      },
      set: (key, value) => {
        try {
          localStorage.setItem(key, value);
          return true;
        } catch (error) {
          console.warn('[Seentics] LocalStorage write error:', error);
          return false;
        }
      },
      remove: (key) => {
        try {
          localStorage.removeItem(key);
          return true;
        } catch (error) {
          console.warn('[Seentics] LocalStorage remove error:', error);
          return false;
        }
      }
    },

    // Safe sessionStorage operations
    safeSessionStorage: {
      get: (key) => {
        try {
          return sessionStorage.getItem(key);
        } catch (error) {
          console.warn('[Seentics] SessionStorage read error:', error);
          return null;
        }
      },
      set: (key, value) => {
        try {
          sessionStorage.setItem(key, value);
          return true;
        } catch (error) {
          console.warn('[Seentics] SessionStorage write error:', error);
          return false;
        }
      }
    },

    // Throttle utility
    throttle: (fn, delay) => {
      let last = 0;
      return (...args) => {
        const now = Date.now();
        if (now - last >= delay) {
          last = now;
          fn(...args);
        }
      };
    },

    // Generate unique ID
    generateId: () => `${Date.now()}_${Math.random().toString(36).slice(2, 6)}`,

    // Get API host
    getApiHost: () => {
      const config = window.SEENTICS_CONFIG;
      if (config?.apiHost) return config.apiHost;
      return window.location.hostname === 'localhost' ? 
        (config?.devApiHost || 'http://localhost:8080') : 
        'https://api.seentics.com';
    }
  };

  // --- Expose to global scope ---
  window.SEENTICS_SHARED = {
    STORAGE_KEYS,
    TIME_CONSTANTS,
    EVENT_TYPES,
    WORKFLOW_TRIGGERS,
    WORKFLOW_ACTIONS,
    WORKFLOW_CONDITIONS,
    FREQUENCY_TYPES,
    SharedUtils
  };

})();
