(function () {
  'use strict';

  // --- Seentics Workflow Tracker v1.0.3 (Optimized) ---

  const S = window.SEENTICS_SHARED || {};
  const T = S.WORKFLOW_TRIGGERS || { PAGE_VIEW: 'Page View', ELEMENT_CLICK: 'Element Click', FUNNEL: 'Funnel', TIME_SPENT: 'Time Spent', EXIT_INTENT: 'Exit Intent' };
  const A = S.WORKFLOW_ACTIONS || { SHOW_MODAL: 'Show Modal', SHOW_BANNER: 'Show Banner', SHOW_NOTIFICATION: 'Show Notification', TRACK_EVENT: 'Track Event', WEBHOOK: 'Webhook', REDIRECT_URL: 'Redirect URL' };
  const C = S.WORKFLOW_CONDITIONS || { URL_PATH: 'URL Path', TRAFFIC_SOURCE: 'Traffic Source', NEW_VS_RETURNING: 'New vs Returning', DEVICE: 'Device Type' };
  const F = S.FREQUENCY_TYPES || { EVERY_TRIGGER: 'every_trigger', ONCE_PER_SESSION: 'once_per_session', ONCE_EVER: 'once_ever' };

  // Storage utilities
  const mkStorage = (store) => ({
    get: k => { try { return store.getItem(k) } catch (e) { return null } },
    set: (k, v) => { try { store.setItem(k, v); return true } catch (e) { return false } },
    remove: k => { try { store.removeItem(k); return true } catch (e) { return false } }
  });

  const storage = S.SharedUtils?.safeLocalStorage || mkStorage(localStorage);
  const sessionStorage = S.SharedUtils?.safeSessionStorage || mkStorage(window.sessionStorage);
  const genId = S.SharedUtils?.generateId || (() => `${Date.now()}_${Math.random().toString(36).slice(2, 6)}`);
  const throttle = S.SharedUtils?.throttle || ((fn, d) => { let l = 0; return (...a) => { const n = Date.now(); if (n - l >= d) { l = n; fn(...a) } } });

  const getDevice = () => {
    if (window.seentics?.getDeviceInfo) return window.seentics.getDeviceInfo().device || 'Desktop';
    const ua = navigator.userAgent;
    return /iPad|Android(?=.*Mobile)|PlayBook|Silk/i.test(ua) ? 'Tablet' : /Mobi|Android/i.test(ua) ? 'Mobile' : 'Desktop';
  };

  // Analytics Manager
  class Analytics {
    constructor(host, siteId) {
      this.host = host;
      this.siteId = siteId;
      this.batch = [];
      this.timer = null;
    }

    add(payload) {
      this.batch.push(payload);
      if (!this.timer) this.timer = setTimeout(() => this.send(), 2000);
    }

    async send() {
      if (!this.batch.length) return;
      const events = [...this.batch];
      this.batch = [];
      this.timer = null;
      try {
        await fetch(`${this.host}/api/v1/workflows/analytics/track/batch`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ siteId: this.siteId, events }),
          credentials: 'omit',
          keepalive: true
        });
      } catch (e) { }
    }

    destroy() { this.send(); }
  }

  // Main Tracker
  const WT = {
    siteId: null,
    visitorId: null,
    isReturning: false,
    workflows: [],
    analytics: null,
    timers: new Set(),
    listeners: new Map(),

    init(siteId) {
      if (!siteId) return console.error('[Seentics] Invalid siteId');

      this.siteId = siteId;
      this.visitorId = this._getVid();
      this.isReturning = this._checkRet();
      this.analytics = new Analytics(this._getHost(), siteId);

      if (siteId === 'preview' && window.__SEENTICS_PREVIEW_WORKFLOW) {
        this.workflows = [window.__SEENTICS_PREVIEW_WORKFLOW];
        this._setup();
      } else {
        this._fetch();
      }
    },

    _getVid() {
      let id = storage.get('seentics_visitor_id');
      if (!id) { id = genId(); storage.set('seentics_visitor_id', id); }
      return id;
    },

    _checkRet() {
      const ret = storage.get('seentics_returning');
      if (ret) return true;
      storage.set('seentics_returning', 'true');
      return false;
    },

    _getHost() {
      const cfg = window.SEENTICS_CONFIG;
      return cfg?.apiHost || (location.hostname === 'localhost' ? 'http://localhost:8080' : 'https://api.seentics.com');
    },

    async _fetch() {
      try {
        const res = await fetch(`${this._getHost()}/api/v1/workflows/site/${this.siteId}/active`, { credentials: 'omit' });
        const data = await res.json();
        this.workflows = data?.workflows?.filter(w => w.status === 'Active') || [];
        this._setup();
      } catch (e) {
        console.error('[Seentics] Failed to fetch workflows:', e);
      }
    },

    _setup() {
      // Page view
      this._check(T.PAGE_VIEW);
      this._setupNav();

      // Time spent
      this.workflows.forEach(w => {
        this._nodes(w, T.TIME_SPENT).forEach(n => {
          const s = n.data?.settings?.seconds || 0;
          if (s > 0) {
            const t = setTimeout(() => this._exec(w, n), s * 1000);
            this.timers.add(t);
          }
        });
      });

      // Exit intent
      const exit = throttle(e => { if (e.clientY <= 5) this._check(T.EXIT_INTENT); }, 50);
      document.addEventListener('mousemove', exit);
      this.listeners.set('exit', () => document.removeEventListener('mousemove', exit));

      // Clicks
      const click = throttle(e => {
        this.workflows.forEach(w => {
          this._nodes(w, T.ELEMENT_CLICK).forEach(n => {
            const sel = n.data?.settings?.selector;
            if (sel && e.target.matches?.(sel)) this._exec(w, n);
          });
        });
      }, 100);
      document.addEventListener('click', click);
      this.listeners.set('click', () => document.removeEventListener('click', click));

      // Funnel
      document.addEventListener('seentics:funnel-event', e => this._check(T.FUNNEL, e.detail));
    },

    _setupNav() {
      let url = location.pathname;
      const onChange = () => {
        const newUrl = location.pathname;
        if (newUrl !== url) { url = newUrl; this._check(T.PAGE_VIEW); }
      };

      const origPush = history.pushState;
      const origReplace = history.replaceState;
      history.pushState = function (...a) { origPush.apply(this, a); onChange(); };
      history.replaceState = function (...a) { origReplace.apply(this, a); onChange(); };
      window.addEventListener('popstate', onChange);
    },

    _nodes(w, type) {
      return w.nodes?.filter(n => n.data?.type === 'Trigger' && n.data?.title === type) || [];
    },

    _check(type, data = {}) {
      this.workflows.forEach(w => {
        this._nodes(w, type).forEach(n => {
          if (this._shouldTrigger(n, type, data)) this._exec(w, n);
        });
      });
    },

    _shouldTrigger(n, type, data) {
      const s = n.data?.settings || {};
      return type === T.FUNNEL ? s.funnelId === data.funnel_id && s.eventType === data.event_type : true;
    },

    async _exec(w, node) {
      const rid = genId();
      if (!this._willExec(w)) return;

      this._track(w, node, 'workflow_trigger', { runId: rid, triggerType: node.data?.title, nodeType: 'trigger' });

      let n = node, executed = false;
      while (n) {
        if (n.data?.type === 'Condition') {
          const result = this._evalCond(n);
          this._track(w, n, 'condition_evaluated', { runId: rid, conditionType: n.data?.title, nodeType: 'condition', result: result ? 'passed' : 'failed' });
          if (!result) {
            this._track(w, n, 'workflow_stopped', { runId: rid, reason: 'condition_failed', nodeType: 'condition' });
            break;
          }
        } else if (n.data?.type === 'Action') {
          if (this._checkFreq(n, w)) {
            executed = true;
            this._track(w, n, 'action_started', { runId: rid, actionType: n.data?.title, nodeType: 'action', frequency: n.data?.settings?.frequency || 'every_trigger' });
            try {
              await this._execAction(n, w);
              this._record(n, w);
              this._track(w, n, 'action_completed', { runId: rid, actionType: n.data?.title, nodeType: 'action', status: 'success' });
            } catch (e) {
              this._track(w, n, 'action_failed', { runId: rid, actionType: n.data?.title, nodeType: 'action', status: 'error', error: e.message });
            }
          }
        }
        n = this._next(w, n);
      }

      if (executed) this._track(w, null, 'workflow_completed', { runId: rid, totalNodes: w.nodes?.length || 0 });
    },

    _willExec(w) {
      try {
        const actions = w.nodes?.filter(n => n.data?.type === 'Action') || [];
        return actions.length > 0 && actions.some(a => this._checkFreq(a, w));
      } catch (e) {
        return true;
      }
    },

    _checkFreq(n, w) {
      try {
        const freq = n.data?.settings?.frequency || F.ONCE_PER_SESSION;
        if (freq === F.EVERY_TRIGGER) return true;

        const key = `seentics_action_${w.id}_${n.id}`;
        if (freq === F.ONCE_PER_SESSION) {
          const sk = `session_${key}`;
          return sessionStorage.get(sk) !== 'true';
        }
        if (freq === F.ONCE_EVER) {
          const pk = `permanent_${key}`;
          return storage.get(pk) !== 'true';
        }
        return true;
      } catch (e) {
        return true;
      }
    },

    _record(n, w) {
      try {
        const freq = n.data?.settings?.frequency || F.EVERY_TRIGGER;
        if (freq === F.EVERY_TRIGGER) return;

        const key = `seentics_action_${w.id}_${n.id}`;
        const ts = new Date().toISOString();

        if (freq === F.ONCE_PER_SESSION) {
          const sk = `session_${key}`;
          sessionStorage.set(sk, 'true');
          sessionStorage.set(`${sk}_timestamp`, ts);
        }
        if (freq === F.ONCE_EVER) {
          const pk = `permanent_${key}`;
          storage.set(pk, 'true');
          storage.set(`${pk}_timestamp`, ts);
        }
      } catch (e) { }
    },

    _evalCond(n) {
      const type = n.data?.title;
      const s = n.data?.settings || {};

      switch (type) {
        case C.URL_PATH: return location.pathname.includes(s.url || '');
        case C.TRAFFIC_SOURCE: return document.referrer.includes(s.referrerUrl || '');
        case C.NEW_VS_RETURNING: return s.visitorType === (this.isReturning ? 'returning' : 'new');
        case C.DEVICE: return this._evalDevice(s);
        default: return true;
      }
    },

    _evalDevice(s) {
      const dev = getDevice();
      if (s.deviceType && s.deviceType !== 'Any' && s.deviceType !== dev) return false;
      if (s.minScreenWidth && screen.width < s.minScreenWidth) return false;
      if (s.maxScreenWidth && screen.width > s.maxScreenWidth) return false;
      if (s.touchSupport) {
        const touch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
        if (s.touchSupport === 'touch' && !touch) return false;
        if (s.touchSupport === 'no-touch' && touch) return false;
      }
      return true;
    },

    async _execAction(n, w) {
      const type = n.data?.title;
      const s = n.data?.settings || {};

      if (n.data?.isServerAction) {
        this._execServer(n, w.id);
        return;
      }

      switch (type) {
        case A.SHOW_MODAL: this._modal(s); break;
        case A.SHOW_BANNER: this._banner(s); break;
        case A.SHOW_NOTIFICATION: this._notify(s); break;
        case A.REDIRECT_URL: if (s.redirectUrl) location.href = s.redirectUrl; break;
        case A.TRACK_EVENT: if (s.eventName && window.seentics?.track) window.seentics.track(s.eventName); break;
        case A.WEBHOOK: this._webhook(s); break;
      }
    },

    _execServer(n, wid) {
      fetch(`${this._getHost()}/api/v1/workflows/execution/action`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ workflowId: wid, nodeId: n.id, siteId: this.siteId, visitorId: this.visitorId }),
        credentials: 'omit',
        keepalive: true
      }).catch(() => { });
    },

    _next(w, n) {
      const edge = w.edges?.find(e => e.source === n.id);
      return edge ? w.nodes?.find(x => x.id === edge.target) : null;
    },

    _modal(s) {
      if (!s.modalTitle && !s.modalContent && !s.customHtml) return;
      if (s.displayMode === 'custom' && s.customHtml) return this._customModal(s);

      const overlay = document.createElement('div');
      overlay.className = 'seentics-overlay';
      const modal = document.createElement('div');
      modal.className = 'seentics-modal';

      if (s.modalTitle) {
        const t = document.createElement('h2');
        t.textContent = s.modalTitle;
        modal.appendChild(t);
      }
      if (s.modalContent) {
        const c = document.createElement('p');
        c.textContent = s.modalContent;
        modal.appendChild(c);
      }

      const close = document.createElement('button');
      close.innerHTML = '×';
      close.className = 'seentics-close-button';
      close.onclick = () => document.body.removeChild(overlay);
      modal.appendChild(close);

      overlay.appendChild(modal);
      document.body.appendChild(overlay);
      overlay.onclick = e => { if (e.target === overlay) document.body.removeChild(overlay); };
    },

    _customModal(s) {
      if (document.querySelector('.seentics-custom-modal-container')) return;

      const cont = document.createElement('div');
      cont.className = 'seentics-custom-modal-container';
      cont.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;z-index:999999;';

      if (s.customCss) {
        const style = document.createElement('style');
        style.id = 'seentics-custom-modal-styles';
        style.textContent = s.customCss;
        document.getElementById('seentics-custom-modal-styles')?.remove();
        document.head.appendChild(style);
      }

      if (s.customHtml) {
        let html = s.customHtml;
        const modalMatch = html.match(/<div[^>]*class="[^"]*exit-overlay[^"]*"[^>]*>[\s\S]*?<\/div>/i);
        if (modalMatch) {
          html = modalMatch[0];
        } else {
          const bodyMatch = html.match(/<body[^>]*>([\s\S]*)<\/body>/i);
          if (bodyMatch) html = bodyMatch[1].replace(/<h1[^>]*>.*?<\/h1>/gi, '').replace(/<p>(?!.*modal).*?<\/p>/gi, '').replace(/Scroll around.*?modal will show up\./gi, '').replace(/My Page Content/gi, '');
        }
        cont.innerHTML = html;
      }

      document.body.appendChild(cont);

      window.seenticsCloseModal = () => {
        cont.parentNode?.removeChild(cont);
        document.querySelectorAll('.seentics-custom-modal-container').forEach(c => c.parentNode?.removeChild(c));
        document.getElementById('seentics-custom-modal-styles')?.remove();
        document.body.style.overflow = '';
        document.documentElement.style.overflow = '';
      };

      setTimeout(() => {
        const overlay = cont.querySelector('#exitOverlay, .exit-overlay');
        if (overlay) overlay.addEventListener('click', e => { if (e.target === overlay) window.seenticsCloseModal(); });
      }, 100);

      if (s.customJs) {
        try {
          const script = document.createElement('script');
          script.textContent = `(function(){const modalContainer=document.querySelector('.seentics-custom-modal-container');const exitOverlay=modalContainer?modalContainer.querySelector('#exitOverlay'):null;const exitClose=modalContainer?modalContainer.querySelector('#exitClose'):null;if(exitClose){exitClose.addEventListener('click',()=>{if(window.seenticsCloseModal){window.seenticsCloseModal();}});}${s.customJs}})();`;
          document.body.appendChild(script);
          setTimeout(() => script.parentNode?.removeChild(script), 100);
        } catch (e) { }
      }
    },

    _banner(s) {
      if (!s.bannerContent && !s.customHtml) return;
      if (s.displayMode === 'custom' && s.customHtml) return this._customBanner(s);

      const banner = document.createElement('div');
      banner.className = `seentics-banner ${s.bannerPosition === 'bottom' ? 'seentics-banner-bottom' : 'seentics-banner-top'}`;
      const content = document.createElement('div');
      content.className = 'seentics-banner-content';
      content.innerHTML = `<p>${s.bannerContent}</p>`;
      banner.appendChild(content);

      const close = document.createElement('button');
      close.innerHTML = '×';
      close.className = 'seentics-close-button';
      close.onclick = () => document.body.removeChild(banner);
      banner.appendChild(close);

      document.body.appendChild(banner);
    },

    _customBanner(s) {
      if (document.querySelector('.seentics-custom-banner-container')) return;

      const cont = document.createElement('div');
      cont.className = 'seentics-custom-banner-container';
      const pos = s.bannerPosition === 'bottom' ? 'bottom:0;' : 'top:0;';
      cont.style.cssText = `position:fixed;left:0;width:100%;${pos}z-index:999998;`;

      if (s.customCss) {
        const style = document.createElement('style');
        style.id = 'seentics-custom-banner-styles';
        style.textContent = s.customCss;
        document.getElementById('seentics-custom-banner-styles')?.remove();
        document.head.appendChild(style);
      }

      if (s.customHtml) {
        let html = s.customHtml;
        const bodyMatch = html.match(/<body[^>]*>([\s\S]*)<\/body>/i);
        if (bodyMatch) html = bodyMatch[1];
        cont.innerHTML = html;
      }

      document.body.appendChild(cont);

      window.seenticsCloseBanner = () => {
        cont.parentNode?.removeChild(cont);
        document.getElementById('seentics-custom-banner-styles')?.remove();
      };

      if (s.customJs) {
        try {
          const script = document.createElement('script');
          script.textContent = `(function(){const bannerContainer=document.querySelector('.seentics-custom-banner-container');${s.customJs}})();`;
          document.body.appendChild(script);
          setTimeout(() => script.parentNode?.removeChild(script), 100);
        } catch (e) { }
      }
    },

    _notify(s) {
      if (!s.notificationMessage) return;

      const notif = document.createElement('div');
      const pos = s.notificationPosition || 'top-right';
      const type = s.notificationType || 'info';
      notif.className = `seentics-notification ${pos} ${type}`;

      const icons = { success: '✓', error: '✕', warning: '⚠', info: 'ℹ' };

      if (s.showIcon !== false) {
        const icon = document.createElement('span');
        icon.className = 'notification-icon';
        icon.textContent = icons[type] || icons.info;
        notif.appendChild(icon);
      }

      const msg = document.createElement('span');
      msg.className = 'notification-message';
      msg.textContent = s.notificationMessage;
      notif.appendChild(msg);

      if (s.showCloseButton !== false) {
        const close = document.createElement('button');
        close.innerHTML = '×';
        close.className = 'notification-close';
        close.onclick = () => this._removeNotif(notif);
        notif.appendChild(close);
      }

      document.body.appendChild(notif);

      const dur = s.notificationDuration || 5000;
      if (dur > 0) setTimeout(() => { if (notif.parentNode) this._removeNotif(notif); }, dur);

      if (s.clickToDismiss !== false) {
        notif.classList.add('clickable');
        notif.onclick = () => this._removeNotif(notif);
      }
    },

    _removeNotif(n) {
      if (!n.parentNode) return;
      n.classList.add('slide-out');
      setTimeout(() => n.parentNode?.removeChild(n), 300);
    },

    _webhook(s) {
      if (!s.webhookUrl) return;
      fetch(s.webhookUrl, {
        method: s.webhookMethod || 'POST',
        headers: { 'Content-Type': 'application/json', ...(s.webhookHeaders || {}) },
        body: JSON.stringify({
          siteId: this.siteId,
          visitorId: this.visitorId,
          sessionId: this.sessionId,
          timestamp: new Date().toISOString(),
          url: location.href,
          userAgent: navigator.userAgent,
          ...(s.webhookData || {})
        }),
        keepalive: true
      }).catch(() => { });
    },

    _track(w, n, evt, opts = {}) {
      const p = {
        w: this.siteId,
        v: this.visitorId,
        s: this.sessionId,
        t: 'wf',
        wf: w.id,
        n: n?.id || null,
        e: evt,
        ts: Date.now()
      };
      if (n?.data?.title && evt !== 'workflow_trigger') p.nt = n.data.title;
      if (opts.result) p.r = opts.result;
      if (opts.status && opts.status !== 'success') p.st = opts.status;
      if (opts.error) p.err = opts.error.substring(0, 100);
      this.analytics.add(p);
    },

    destroy() {
      this.timers.forEach(t => clearTimeout(t));
      this.timers.clear();
      this.listeners.forEach(c => c());
      this.listeners.clear();
      if (this.analytics) { this.analytics.destroy(); this.analytics = null; }
      this.workflows = [];
      this.siteId = null;
      this.visitorId = null;
    }
  };

  // Auto-init
  const sid = window.SEENTICS_SITE_ID || document.currentScript?.getAttribute('data-site-id');
  if (sid) {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', () => WT.init(sid));
    } else {
      WT.init(sid);
    }
  }

  // Expose
  window.seentics = window.seentics || {};
  window.seentics.workflowTracker = WT;

  window.triggerFunnelEvent = (fid, evt, step = 0, data = {}) => {
    document.dispatchEvent(new CustomEvent('seentics:funnel-event', {
      detail: { funnel_id: fid, event_type: evt, step_index: step, ...data }
    }));
  };

  window.addEventListener('beforeunload', () => WT.destroy());

})();