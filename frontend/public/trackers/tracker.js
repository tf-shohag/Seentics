(function () {
  // --- Seentics Analytics Tracker v1.0.2 (Performance & Reliability Fixed) ---

  // --- Main Tracker ---
  if (document.currentScript) {
    (function (doc, win, nav, loc) {
      // Configuration
      const scriptTag = doc.currentScript;
      const siteId = scriptTag.getAttribute('data-site-id');
      const apiHost = win.SEENTICS_CONFIG?.apiHost ||
        (loc.hostname === 'localhost' ?
          (win.SEENTICS_CONFIG?.devApiHost || 'http://localhost:8080') :
          'https://api.seentics.com');
      const API_ENDPOINT = `${apiHost}/api/v1/analytics/event/batch`;
      const DEBUG = !!(win.SEENTICS_CONFIG?.debugMode) && loc.hostname === 'localhost';

      // Constants
      const VISITOR_ID_KEY = 'seentics_visitor_id';
      const SESSION_ID_KEY = 'seentics_session_id';
      const SESSION_LAST_SEEN_KEY = 'seentics_session_last_seen';
      const SESSION_EXPIRY_MS = 1800000; // 30 minutes
      const VISITOR_EXPIRY_MS = 2592000000; // 30 days
      const BATCH_DELAY = 100;
      const MAX_RETRY_ATTEMPTS = 3;
      const RETRY_DELAY = 1000;

      // State variables
      let visitorId = null;
      let sessionId = null;
      let pageStartTime = performance.now();
      let pageviewSent = false;
      let currentUrl = loc.pathname;
      let cachedUTMParams = null;
      let lastUrlForUTM = '';
      let activityTimeout = null;
      let isDestroyed = false;

      // Event batching
      const eventQueue = [];
      let flushTimeout = null;
      const pendingRequests = new Map();

      // Cleanup tracking
      const cleanupTasks = [];

      // --- Core Functions ---

      function safeLocalStorage(operation) {
        try {
          return operation();
        } catch (error) {
          if (DEBUG) console.warn('Seentics: LocalStorage error:', error);
          return null;
        }
      }

      function getOrCreateId(key, expiryMs) {
        return safeLocalStorage(() => {
          const raw = localStorage.getItem(key);
          if (raw) {
            let obj;
            try {
              obj = JSON.parse(raw);
            } catch {
              if (raw.length > 10) return raw;
            }

            if (obj?.value && (!obj.expiresAt || Date.now() < obj.expiresAt)) {
              return obj.value;
            }
          }

          const value = Date.now().toString(36) + Math.random().toString(36).slice(2);
          const expiresAt = Date.now() + expiryMs;
          const data = JSON.stringify({ value, expiresAt });

          localStorage.setItem(key, data);
          if (key === SESSION_ID_KEY) {
            localStorage.setItem(SESSION_LAST_SEEN_KEY, Date.now().toString());
          }

          return value;
        }) || Date.now().toString(36) + Math.random().toString(36).slice(2);
      }

      function refreshSessionIfNeeded() {
        safeLocalStorage(() => {
          const now = Date.now();
          const lastSeenStr = localStorage.getItem(SESSION_LAST_SEEN_KEY);
          const lastSeen = lastSeenStr ? parseInt(lastSeenStr, 10) : 0;

          if (lastSeen && now - lastSeen < SESSION_EXPIRY_MS) {
            localStorage.setItem(SESSION_LAST_SEEN_KEY, now.toString());
            return;
          }

          sessionId = getOrCreateId(SESSION_ID_KEY, SESSION_EXPIRY_MS);
        });
      }

      function getCurrentDomain() {
        return loc.hostname.split(':')[0];
      }

      // --- Request Deduplication ---
      async function deduplicatedFetch(url, options, retryCount = 0) {
        if (isDestroyed) return;

        const key = `${url}:${JSON.stringify(options)}`;
        if (pendingRequests.has(key)) {
          return pendingRequests.get(key);
        }

        const promise = (async () => {
          try {
            const response = await fetch(url, options);
            if (!response.ok && retryCount < MAX_RETRY_ATTEMPTS) {
              await new Promise(resolve => setTimeout(resolve, RETRY_DELAY * (retryCount + 1)));
              return deduplicatedFetch(url, options, retryCount + 1);
            }
            return response;
          } catch (error) {
            if (retryCount < MAX_RETRY_ATTEMPTS) {
              await new Promise(resolve => setTimeout(resolve, RETRY_DELAY * (retryCount + 1)));
              return deduplicatedFetch(url, options, retryCount + 1);
            }
            throw error;
          }
        })();

        pendingRequests.set(key, promise);
        promise.finally(() => pendingRequests.delete(key));
        return promise;
      }

      // --- Event Batching ---

      function queueEvent(event) {
        if (isDestroyed) return;
        eventQueue.push(event);

        if (!flushTimeout) {
          flushTimeout = setTimeout(flushEventQueue, BATCH_DELAY);
        }
      }

      async function flushEventQueue() {
        if (isDestroyed || eventQueue.length === 0) return;

        const events = eventQueue.splice(0);
        flushTimeout = null;

        try {
          const batchData = {
            siteId,
            domain: getCurrentDomain(),
            events
          };

          if (nav.sendBeacon) {
            const blob = new Blob([JSON.stringify(batchData)], { type: 'application/json' });
            nav.sendBeacon(API_ENDPOINT, blob);
          } else {
            await deduplicatedFetch(API_ENDPOINT, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(batchData),
              credentials: 'omit', // Explicitly omit credentials for public tracking
              keepalive: true
            });
          }
        } catch (error) {
          if (DEBUG) console.warn('Seentics: Failed to send events', error);
          // Re-queue events on failure
          eventQueue.unshift(...events);
        }
      }

      // --- Custom Event Tracking ---

      // Cache for session-level data to avoid repetition
      let lastReferrer = null;
      let deviceInfo = null;

      function getDeviceInfo() {
        if (deviceInfo) return deviceInfo;
        const ua = nav.userAgent;
        deviceInfo = {
          browser: getBrowser(ua),
          device: getDevice(ua),
          os: getOS(ua)
        };
        return deviceInfo;
      }

      function getBrowser(ua) {
        if (ua.includes('Chrome')) return 'Chrome';
        if (ua.includes('Firefox')) return 'Firefox';
        if (ua.includes('Safari')) return 'Safari';
        if (ua.includes('Edge')) return 'Edge';
        return 'Other';
      }

      function getDevice(ua) {
        if (/iPad|Android(?=.*Mobile)|PlayBook|Silk/i.test(ua)) return 'Tablet';
        if (/Mobi|Android/i.test(ua)) return 'Mobile';
        return 'Desktop';
      }

      function getOS(ua) {
        if (ua.includes('Windows')) return 'Windows';
        if (ua.includes('Mac')) return 'macOS';
        if (ua.includes('Linux')) return 'Linux';
        if (ua.includes('Android')) return 'Android';
        if (ua.includes('iOS')) return 'iOS';
        return 'Other';
      }

      function trackCustomEvent(eventName, properties = {}) {
        if (!siteId || !eventName || typeof eventName !== 'string') {
          if (DEBUG) console.warn('Seentics: Invalid event parameters');
          return;
        }

        refreshSessionIfNeeded();

        const event = {
          website_id: siteId,
          visitor_id: visitorId,
          session_id: sessionId,
          event_type: eventName,
          page: loc.pathname,
          properties,
          timestamp: new Date().toISOString()
        };

        // Only include referrer if it's new or first event in session
        const currentReferrer = doc.referrer || null;
        if (currentReferrer !== lastReferrer) {
          event.referrer = currentReferrer;
          lastReferrer = currentReferrer;
        }


        const deviceData = getDeviceInfo();
        event.browser = deviceData.browser;
        event.device = deviceData.device;
        event.os = deviceData.os;

        queueEvent(event);

        // Emit for funnel tracker
        try {
          doc.dispatchEvent(new CustomEvent('seentics:custom-event', {
            detail: { eventName, data: properties }
          }));
        } catch { }
      }

      // --- Pageview Tracking ---

      // Optimized UTM parameter extraction with better caching
      let urlParamsCache = null;

      function extractUTMParameters() {
        const currentUrl = loc.search;
        if (currentUrl === lastUrlForUTM && cachedUTMParams) {
          return cachedUTMParams;
        }

        // Reuse URLSearchParams object if URL hasn't changed
        if (!urlParamsCache || currentUrl !== lastUrlForUTM) {
          urlParamsCache = new URLSearchParams(currentUrl);
        }

        cachedUTMParams = {
          utm_source: urlParamsCache.get('utm_source'),
          utm_medium: urlParamsCache.get('utm_medium'),
          utm_campaign: urlParamsCache.get('utm_campaign'),
          utm_term: urlParamsCache.get('utm_term'),
          utm_content: urlParamsCache.get('utm_content')
        };
        lastUrlForUTM = currentUrl;

        return cachedUTMParams;
      }


      function sendPageview() {
        if (pageviewSent || !siteId || isDestroyed) return;

        const timeOnPage = Math.round((performance.now() - pageStartTime) / 1000);

        refreshSessionIfNeeded();

        const event = {
          website_id: siteId,
          visitor_id: visitorId,
          session_id: sessionId,
          event_type: 'pageview',
          page: loc.pathname,
          time_on_page: timeOnPage,
          timestamp: new Date().toISOString()
        };

        // Only include referrer if it's new or first pageview in session
        const currentReferrer = doc.referrer || null;
        if (currentReferrer !== lastReferrer) {
          event.referrer = currentReferrer;
          lastReferrer = currentReferrer;
        }


        const deviceData = getDeviceInfo();
        event.browser = deviceData.browser;
        event.device = deviceData.device;
        event.os = deviceData.os;

        // Include UTM parameters if available
        const utmParams = extractUTMParameters();
        // Only include UTM params that have values
        Object.keys(utmParams).forEach(key => {
          if (utmParams[key]) {
            event[key] = utmParams[key];
          }
        });

        pageviewSent = true;
        queueEvent(event);
      }

      // --- Event Listeners ---

      function onActivity() {
        if (activityTimeout || isDestroyed) return;
        activityTimeout = setTimeout(() => {
          if (!isDestroyed) {
            refreshSessionIfNeeded();
          }
          activityTimeout = null;
        }, 1000);
      }

      function onRouteChange() {
        if (isDestroyed) return;
        const newUrl = loc.pathname;
        if (newUrl === currentUrl) return;

        currentUrl = newUrl;
        cachedUTMParams = null;
        pageStartTime = performance.now();
        pageviewSent = false;

        requestIdleCallback(() => sendPageview());
      }

      // Cleanup function
      function cleanup() {
        isDestroyed = true;

        // Clear timers
        if (activityTimeout) {
          clearTimeout(activityTimeout);
          activityTimeout = null;
        }
        if (flushTimeout) {
          clearTimeout(flushTimeout);
          flushTimeout = null;
        }

        // Execute cleanup tasks
        cleanupTasks.forEach(task => {
          try {
            task();
          } catch (error) {
            if (DEBUG) console.warn('Seentics: Cleanup error:', error);
          }
        });
        cleanupTasks.length = 0;

        // Final flush
        if (eventQueue.length > 0) {
          flushEventQueue();
        }

        // Clear pending requests
        pendingRequests.clear();
      }

      // Setup all event listeners
      function setupEventListeners() {
        const visibilityHandler = () => {
          if (isDestroyed) return;
          if (doc.hidden && !pageviewSent) {
            sendPageview();
          }
          refreshSessionIfNeeded();
        };
        doc.addEventListener('visibilitychange', visibilityHandler, { passive: true });
        cleanupTasks.push(() => doc.removeEventListener('visibilitychange', visibilityHandler));

        const beforeUnloadHandler = () => {
          if (!pageviewSent) sendPageview();
          if (eventQueue.length > 0) flushEventQueue();
          cleanup();
        };
        win.addEventListener('beforeunload', beforeUnloadHandler);
        cleanupTasks.push(() => win.removeEventListener('beforeunload', beforeUnloadHandler));

        // Activity listeners with throttling for high-frequency events
        let throttleTimeout = null;
        const throttledActivity = () => {
          if (throttleTimeout) return;
          throttleTimeout = setTimeout(() => {
            onActivity();
            throttleTimeout = null;
          }, 100); // Throttle to max 10 calls per second
        };

        // Low frequency events - no throttling needed
        ['click', 'keydown'].forEach(evt => {
          doc.addEventListener(evt, onActivity, { passive: true });
          cleanupTasks.push(() => doc.removeEventListener(evt, onActivity));
        });

        // High frequency events - throttled
        ['scroll', 'mousemove', 'touchstart'].forEach(evt => {
          doc.addEventListener(evt, throttledActivity, { passive: true });
          cleanupTasks.push(() => doc.removeEventListener(evt, throttledActivity));
        });

        // SPA navigation detection
        const originalPushState = history.pushState;
        const originalReplaceState = history.replaceState;

        history.pushState = function (...args) {
          originalPushState.apply(this, args);
          onRouteChange();
        };

        history.replaceState = function (...args) {
          originalReplaceState.apply(this, args);
          onRouteChange();
        };

        const popstateHandler = onRouteChange;
        win.addEventListener('popstate', popstateHandler);
        cleanupTasks.push(() => {
          win.removeEventListener('popstate', popstateHandler);
          history.pushState = originalPushState;
          history.replaceState = originalReplaceState;
        });
      }

      // --- Resource Loading ---

      function loadResource(src, type = 'script') {
        return new Promise((resolve) => {
          const existing = doc.querySelector(`${type}[${type === 'script' ? 'src' : 'href'}="${src}"]`);
          if (existing) {
            resolve();
            return;
          }

          const element = doc.createElement(type);
          if (type === 'script') {
            element.src = src;
            element.async = true;
          } else {
            element.rel = 'stylesheet';
            element.href = src;
            element.type = 'text/css';
          }

          element.onload = () => resolve();
          element.onerror = () => resolve();

          doc.head.appendChild(element);
        });
      }

      async function loadAdditionalTrackers() {
        try {
          await Promise.all([
            loadResource('/trackers/shared-constants.js'), // Load shared constants first
            loadResource('/trackers/styles/tracker-styles.css', 'link'),
            loadResource('/trackers/workflow-tracker.js'),
            loadResource('/trackers/funnel-tracker.js')
          ]);

          if (win.seentics?.workflowTracker) {
            await win.seentics.workflowTracker.init(siteId);
          }
        } catch (error) {
          if (DEBUG) console.warn('Seentics: Failed to load additional trackers', error);
        }
      }

      // --- Initialization ---
      function init() {
        if (!siteId) {
          if (DEBUG) console.warn('Seentics: No site ID provided');
          return;
        }

        // Initialize IDs
        visitorId = getOrCreateId(VISITOR_ID_KEY, VISITOR_EXPIRY_MS);
        sessionId = getOrCreateId(SESSION_ID_KEY, SESSION_EXPIRY_MS);

        requestIdleCallback(() => sendPageview());
        setupEventListeners();
        requestIdleCallback(() => loadAdditionalTrackers());

        // Public API
        win.seentics = {
          siteId,
          apiHost,
          track: trackCustomEvent,
          sendPageview,
          cleanup,
          getDeviceInfo, // Expose device info for other trackers
          ...(DEBUG && {
            getVisitorId: () => visitorId,
            getSessionId: () => sessionId,
            flushEvents: flushEventQueue
          })
        };
      }

      // Initialize
      if (doc.readyState === 'loading') {
        doc.addEventListener('DOMContentLoaded', init);
      } else {
        requestIdleCallback(init);
      }

    })(document, window, navigator, location);
  }
})();