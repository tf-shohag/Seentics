(function (d) {
  'use strict';

  // --- Seentics Funnel Tracker v1.0.3 (Optimized) ---

  const DBG = location.hostname === 'localhost';
  const script = d.currentScript;
  let siteId = script?.getAttribute('data-site-id') || d.querySelector('script[data-site-id]')?.getAttribute('data-site-id') || window.seentics?.siteId || window.SEENTICS_SITE_ID;

  if (!siteId) {
    if (DBG) console.warn('Seentics Funnel: No site ID');
    return;
  }

  // Shared constants
  const S = window.SEENTICS_SHARED || {};
  const K = S.STORAGE_KEYS || { VISITOR_ID: 'seentics_visitor_id', SESSION_ID: 'seentics_session_id' };
  const T = S.TIME_CONSTANTS || { VISITOR_EXPIRY_MS: 2592000000, SESSION_EXPIRY_MS: 1800000, BATCH_DELAY: 1000 };

  const apiHost = S.SharedUtils?.getApiHost() || (window.SEENTICS_CONFIG?.apiHost || (location.hostname === 'localhost' ? 'http://localhost:8080' : 'https://api.seentics.com'));
  const API = `${apiHost}/api/v1/funnels/track`;
  const trackEnabled = (script?.getAttribute('data-track-funnels') || '').toLowerCase() !== 'false';
  const STATE_KEY = 'seentics_funnel_state';

  // Get or create ID
  const getId = (key, exp) => {
    const storage = S.SharedUtils?.safeLocalStorage;
    if (storage) {
      try {
        let id = storage.get(key);
        if (!id) {
          id = S.SharedUtils.generateId();
          storage.set(key, id);
        }
        return id;
      } catch (e) {
        return S.SharedUtils.generateId();
      }
    }
    try {
      let id = localStorage.getItem(key);
      if (!id) {
        id = Date.now().toString(36) + Math.random().toString(36).substring(2);
        localStorage.setItem(key, id);
      }
      return id;
    } catch (e) {
      return Date.now().toString(36) + Math.random().toString(36).substring(2);
    }
  };

  // Debounce utility
  const debounce = S.SharedUtils?.throttle || ((fn, delay) => {
    let timer;
    return (...args) => {
      clearTimeout(timer);
      timer = setTimeout(() => fn.apply(null, args), delay);
    };
  });

  // State
  let vid = getId(K.VISITOR_ID, T.VISITOR_EXPIRY_MS);
  let sid = getId(K.SESSION_ID, T.SESSION_EXPIRY_MS);
  let funnels = new Map();
  let queue = [];
  let timer = null;
  let validated = false;
  let url = location.pathname;
  let clickHandler = null;

  // --- Core Functions ---

  async function init() {
    if (!trackEnabled || !siteId) return;
    await loadFunnels();
    startMonitoring();
  }

  async function loadFunnels() {
    try {
      const apiUrl = `${apiHost}/api/v1/funnels/active?website_id=${siteId}`;
      if (DBG) console.log('🔍 Fetching funnels:', apiUrl);

      const res = await fetch(apiUrl, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'omit',
        keepalive: true
      });

      if (res.ok) {
        const data = await res.json();
        const list = data === null ? [] : (Array.isArray(data) ? data : (data?.funnels || data?.data || []));

        if (DBG) console.log('🔍 API Response:', { status: res.status, data });
        if (DBG) console.log('🔍 Extracted funnels:', list);

        funnels.clear();
        initFunnels(list);
        validated = true;

        setTimeout(() => {
          if (DBG) console.log('🔍 Initial page check:', url);
          checkPage();
        }, 100);

        if (queue.length > 0) sendEvents();
      } else {
        const err = await res.text();
        console.warn('🔍 Fetch failed:', { status: res.status, statusText: res.statusText, url: apiUrl, error: err });
      }
    } catch (e) {
      console.warn('🔍 Load error:', e);
      funnels.clear();
      initFunnels([]);
    }
  }

  function initFunnels(list) {
    if (DBG) console.log('🔍 Initializing funnels:', list);
    if (!Array.isArray(list)) {
      console.warn('🔍 Funnels not array:', list);
      return;
    }

    list.forEach(f => {
      if (DBG) console.log('🔍 Processing funnel:', f.name, 'Active:', f.is_active);
      if (f.is_active) {
        funnels.set(f.id, {
          ...f,
          currentStep: 0,
          completedSteps: [],
          startedAt: null,
          lastActivity: null,
          converted: false
        });
        if (DBG) console.log('🔍 Added funnel:', f.id, 'with', f.steps?.length || 0, 'steps');
      }
    });

    if (DBG) console.log('🔍 Active funnels:', funnels.size);
  }

  const saveState = debounce(() => {
    try {
      const state = {};
      funnels.forEach((s, id) => {
        state[id] = {
          currentStep: s.currentStep,
          completedSteps: s.completedSteps,
          startedAt: s.startedAt,
          lastActivity: s.lastActivity?.toISOString?.() || s.lastActivity,
          converted: s.converted
        };
      });
      const storage = S.SharedUtils?.safeLocalStorage || {
        set: (k, v) => { try { localStorage.setItem(k, v); return true; } catch { return false; } }
      };
      storage.set(STATE_KEY, JSON.stringify(state));
    } catch (e) {
      console.warn('State save failed:', e);
    }
  }, 200);

  function startMonitoring() {
    checkPage();
    monitorClicks();
    monitorCustom();
  }

  function checkPage() {
    if (DBG) console.log('🔍 Checking page:', url);
    if (DBG) console.log('🔍 Funnels count:', funnels.size);

    let matched = false;

    funnels.forEach((state, fid) => {
      if (!state.steps) {
        if (DBG) console.log('🔍 No steps:', fid);
        return;
      }

      const isActive = state.currentStep > 0;
      const canStart = state.currentStep === 0;

      if (DBG) console.log(`🔍 Funnel ${fid} - step: ${state.currentStep}, active: ${isActive}`);

      let found = false;

      state.steps.forEach((step, idx) => {
        const num = idx + 1;
        if (step.type === 'page' && step.condition?.page) {
          const match = matchPage(url, step.condition.page);
          if (match) {
            if (num === 1 || isActive) {
              if (DBG) console.log(`🎯 Page match: funnel ${fid}, step ${num}: ${step.name}`);
              progress(fid, num, step);
              found = true;
              matched = true;
            } else {
              if (DBG) console.log(`🔍 Ignoring step ${num} - not in funnel`);
            }
          }
        }
      });

      if (!found && isActive) {
        if (state.startedAt && (new Date() - new Date(state.startedAt)) > 5000) {
          if (DBG) console.log(`🔍 Dropoff: ${fid}`);
          queueDropoff(fid, state);
        } else {
          if (DBG) console.log(`🔍 Left too quickly: ${fid}`);
        }
        state.currentStep = 0;
        state.completedSteps = [];
        state.startedAt = null;
        state.converted = false;
        funnels.set(fid, state);
        saveState();
      }
    });

    if (DBG && !matched) console.log('🔍 No matches, no API calls');
  }

  function queueDropoff(fid, state) {
    const evt = createEvent(fid, state, {
      step_name: state.steps[state.currentStep - 1]?.name || 'Unknown',
      step_type: state.steps[state.currentStep - 1]?.type || 'page',
      dropoff_reason: "navigated_to_unexpected_page",
      event_type: "dropoff"
    });
    queue.push(evt);
    sendEvents();
  }

  function monitorClicks() {
    const throttle = S.SharedUtils?.throttle || debounce;
    clickHandler = throttle((e) => {
      const el = e.target;
      funnels.forEach((state, fid) => {
        if (!state.steps) return;
        state.steps.forEach((step, idx) => {
          if (step.type === 'event' && step.condition?.event && matchEvent(el, step.condition.event)) {
            progress(fid, idx + 1, step);
          }
        });
      });
    }, 100);
    d.addEventListener('click', clickHandler, { passive: true });
  }

  function monitorCustom() {
    d.addEventListener('seentics:custom-event', (e) => {
      const { eventName, data } = e.detail;
      checkCustom(eventName, data);
    });
  }

  function checkCustom(name, data = {}) {
    funnels.forEach((state, fid) => {
      if (!state.steps) return;
      state.steps.forEach((step, idx) => {
        if (step.type === 'custom' && step.condition?.custom === name) {
          progress(fid, idx + 1, step, data);
        }
      });
    });
  }

  function progress(fid, num, step, extra = {}) {
    const state = funnels.get(fid);
    if (!state) return;

    if (state._updating) {
      if (DBG) console.log(`🔍 Update in progress: ${fid}`);
      return;
    }
    state._updating = true;

    try {
      const next = state.currentStep + 1;

      if (num === 1) {
        if (DBG) console.log(`🔍 Starting funnel ${fid}`);
        state.currentStep = 1;
        state.completedSteps = [0];
        state.startedAt = new Date().toISOString();
        state.converted = false;
      } else if (num === next) {
        if (DBG) console.log(`🔍 Sequential progress: ${fid}, step ${num}`);
        state.currentStep = num;
        state.completedSteps.push(num - 1);
        if (num === state.steps.length) {
          state.converted = true;
          console.log(`🔍 Funnel ${fid} completed!`);
        }
      } else {
        if (DBG) console.log(`🔍 Invalid progression: ${fid}, current=${state.currentStep}, attempted=${num}`);
        return;
      }

      state.lastActivity = new Date();
      funnels.set(fid, state);
      saveState();
      queueEvent(fid, state, num, extra);
    } finally {
      state._updating = false;
    }
  }

  function queueEvent(fid, state, step, extra) {
    const evt = createEvent(fid, state, {
      step_name: state.steps[step - 1]?.name || `Step ${step}`,
      step_type: state.steps[step - 1]?.type || 'page',
      ...extra
    });
    queue.push(evt);
    clearTimeout(timer);
    timer = setTimeout(sendEvents, T.BATCH_DELAY);
  }

  function createEvent(fid, state, props) {
    return {
      funnel_id: fid,
      website_id: siteId,
      visitor_id: vid,
      session_id: sid,
      current_step: state.currentStep,
      completed_steps: state.completedSteps,
      started_at: state.startedAt,
      last_activity: state.lastActivity,
      converted: state.converted,
      event_type: props.event_type,
      ...props
    };
  }

  async function sendEvents() {
    if (!validated || queue.length === 0) return;
    const events = [...queue];
    queue = [];

    if (DBG) console.log(`🔍 Sending ${events.length} events`);

    try {
      const send = async (evt) => {
        if (DBG) console.log('🔍 Sending:', evt);
        if (navigator.sendBeacon) {
          const blob = new Blob([JSON.stringify(evt)], { type: 'application/json' });
          if (!navigator.sendBeacon(API, blob)) console.warn('🔍 sendBeacon failed');
        } else {
          await fetch(API, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(evt),
            credentials: 'omit',
            keepalive: true
          });
        }
        try {
          d.dispatchEvent(new CustomEvent('seentics:funnel-event', { detail: evt }));
        } catch (e) {
          console.warn('🔍 Emit failed:', e);
        }
      };

      await Promise.all(events.map(send));
      if (DBG) console.log('🔍 Events sent');
    } catch (e) {
      console.error('🔍 Send error:', e);
      queue.unshift(...events);
    }
  }

  // Public API
  function trackStep(fid, num, data = {}) {
    if (!trackEnabled || !siteId) return;
    const state = funnels.get(fid);
    if (!state) {
      console.warn(`🔍 Funnel ${fid} not found`);
      return;
    }
    if (state.steps?.[num - 1]) progress(fid, num, state.steps[num - 1], data);
  }

  function trackConversion(fid, val = 0, extra = {}) {
    if (!trackEnabled || !siteId) return;
    const state = funnels.get(fid);
    if (!state) {
      console.warn(`🔍 Funnel ${fid} not found`);
      return;
    }
    state.converted = true;
    state.lastActivity = new Date();
    funnels.set(fid, state);
    saveState();

    const evt = createEvent(fid, state, {
      step_name: state.steps[state.currentStep - 1]?.name || `Step ${state.currentStep}`,
      step_type: state.steps[state.currentStep - 1]?.type || 'page',
      conversion_value: val,
      event_type: "conversion",
      ...extra
    });
    queue.push(evt);
    sendEvents();
  }

  // Helpers
  function matchPage(path, pattern) {
    if (path === pattern) return true;
    if (pattern.includes('*')) {
      const regex = new RegExp(pattern.replace(/\*/g, '.*'));
      return regex.test(path);
    }
    if (pattern.endsWith('*')) return path.startsWith(pattern.slice(0, -1));
    return false;
  }

  function matchEvent(el, sel) {
    try {
      return !!(el.matches?.(sel) || el.closest?.(sel));
    } catch {
      return false;
    }
  }

  // Route change
  function onRoute() {
    const newUrl = location.pathname;
    if (newUrl === url) return;
    if (DBG) console.log(`🔍 Route: ${url} -> ${newUrl}`);
    url = newUrl;
    if (funnels.size === 0) {
      if (DBG) console.log('🔍 No funnels, skip');
      return;
    }
    setTimeout(() => {
      if (DBG) console.log('🔍 Checking after route');
      checkPage();
    }, 200);
  }

  // SPA navigation
  const origPush = history.pushState;
  const origReplace = history.replaceState;
  history.pushState = function (...a) { origPush.apply(history, a); onRoute(); };
  history.replaceState = function (...a) { origReplace.apply(history, a); onRoute(); };
  window.addEventListener('popstate', onRoute);
  window.addEventListener('beforeunload', () => { if (queue.length > 0) sendEvents(); });

  // Init
  function start() {
    if (!siteId) {
      console.warn('🔍 No siteId');
      return;
    }

    if (DBG) console.log('🔍 Init:', siteId);
    if (DBG) console.log('🔍 URL:', url);
    if (DBG) console.log('🔍 API:', apiHost);

    init();

    window.seentics = window.seentics || {};
    window.seentics.funnelTracker = {
      trackFunnelStep: trackStep,
      trackFunnelConversion: trackConversion,
      getFunnelState: (fid) => funnels.get(fid),
      getActiveFunnels: () => Array.from(funnels.keys()),
      monitorPageChanges: checkPage,
      triggerFunnelEvent: (fid, type, idx = 0, extra = {}) => {
        const evt = {
          funnel_id: fid,
          current_step: idx,
          completed_steps: [idx],
          started_at: new Date().toISOString(),
          last_activity: new Date().toISOString(),
          converted: type === 'conversion',
          properties: { event_type: type, ...extra }
        };
        console.log('🎯 Manual trigger:', evt);
        d.dispatchEvent(new CustomEvent('seentics:funnel-event', { detail: evt }));
        return evt;
      },
      cleanup: () => {
        if (timer) { clearTimeout(timer); timer = null; }
        if (clickHandler?.cancel) clickHandler.cancel();
        clickHandler = null;
        if (queue.length > 0) sendEvents();
        funnels.clear();
        queue = [];
        window.removeEventListener('popstate', onRoute);
        if (origPush) history.pushState = origPush;
        if (origReplace) history.replaceState = origReplace;
      }
    };
  }

  if (d.readyState === 'loading') {
    d.addEventListener('DOMContentLoaded', start);
  } else {
    start();
  }

  window.addEventListener('beforeunload', () => {
    if (window.seentics?.funnelTracker?.cleanup) window.seentics.funnelTracker.cleanup();
  });

})(document);