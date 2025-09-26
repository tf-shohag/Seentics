# 🌐 Tracker API URL Configuration Update

## 🎯 **OBJECTIVE**
Update all tracking scripts to use the correct production API URL (`https://www.api.seentics.com`) instead of dynamic hostname-based URLs.

## 🔧 **CHANGES IMPLEMENTED**

### **1. Analytics Tracker (tracker.js)**
```javascript
// BEFORE
const apiHost = win.SEENTICS_CONFIG?.apiHost ||
  (loc.hostname === 'localhost' ?
    (win.SEENTICS_CONFIG?.devApiHost || 'http://localhost:8080') :
    `https://${loc.hostname}`);  // ❌ Dynamic hostname

// AFTER  
const apiHost = win.SEENTICS_CONFIG?.apiHost ||
  (loc.hostname === 'localhost' ?
    (win.SEENTICS_CONFIG?.devApiHost || 'http://localhost:8080') :
    'https://www.api.seentics.com');  // ✅ Fixed production URL
```

### **2. Workflow Tracker (workflow-tracker.js)**
```javascript
// BEFORE
_getApiHost() {
  const config = window.SEENTICS_CONFIG;
  if (config?.apiHost) return config.apiHost;
  return window.location.hostname === 'localhost' ? 
    'http://localhost:8080' : 
    `https://${window.location.hostname}`;  // ❌ Dynamic hostname
}

// AFTER
_getApiHost() {
  const config = window.SEENTICS_CONFIG;
  if (config?.apiHost) return config.apiHost;
  return window.location.hostname === 'localhost' ? 
    'http://localhost:8080' : 
    'https://www.api.seentics.com';  // ✅ Fixed production URL
}
```

### **3. Funnel Tracker (funnel-tracker.js)**
```javascript
// BEFORE
const apiHost = window.SEENTICS_CONFIG?.apiHost ||
  (window.location.hostname === 'localhost' ? 
    'http://localhost:8080' : 
    `https://${window.location.hostname}`);  // ❌ Dynamic hostname

// AFTER
const apiHost = window.SEENTICS_CONFIG?.apiHost ||
  (window.location.hostname === 'localhost' ? 
    'http://localhost:8080' : 
    'https://www.api.seentics.com');  // ✅ Fixed production URL
```

## 📋 **API ENDPOINTS AFFECTED**

### **Analytics Tracking**
- **Endpoint**: `https://www.api.seentics.com/api/v1/analytics/event/batch`
- **Used by**: Main analytics tracker
- **Purpose**: Batch event tracking

### **Workflow Tracking**
- **Endpoint**: `https://www.api.seentics.com/api/v1/workflows/analytics/track/batch`
- **Used by**: Workflow tracker analytics
- **Purpose**: Workflow event tracking

- **Endpoint**: `https://www.api.seentics.com/api/v1/workflows/site/{siteId}/active`
- **Used by**: Workflow tracker initialization
- **Purpose**: Fetch active workflows

### **Funnel Tracking**
- **Endpoint**: `https://www.api.seentics.com/api/v1/funnels/track`
- **Used by**: Funnel tracker
- **Purpose**: Funnel event tracking

- **Endpoint**: `https://www.api.seentics.com/api/v1/funnels/active?website_id={siteId}`
- **Used by**: Funnel tracker initialization
- **Purpose**: Fetch active funnels

## 🔄 **ENVIRONMENT HANDLING**

### **Development (localhost)**
```javascript
// All trackers use localhost for development
'http://localhost:8080'
```

### **Production (any other domain)**
```javascript
// All trackers now use fixed production URL
'https://www.api.seentics.com'
```

### **Custom Configuration Override**
```javascript
// Users can still override via window.SEENTICS_CONFIG
window.SEENTICS_CONFIG = {
  apiHost: 'https://custom-api.example.com'
};
```

## 🚀 **USAGE SCENARIOS**

### **Scenario 1: Website Integration**
```html
<!-- From any website (e.g., https://example.com) -->
<script src="https://cdn.seentics.com/trackers/tracker.js" data-site-id="site123"></script>
<!-- Will send data to: https://www.api.seentics.com -->
```

### **Scenario 2: Local Development**
```html
<!-- From localhost:3000 -->
<script src="http://localhost:8080/trackers/tracker.js" data-site-id="site123"></script>
<!-- Will send data to: http://localhost:8080 -->
```

### **Scenario 3: Custom API Host**
```html
<script>
window.SEENTICS_CONFIG = {
  apiHost: 'https://api.custom-domain.com'
};
</script>
<script src="https://cdn.seentics.com/trackers/tracker.js" data-site-id="site123"></script>
<!-- Will send data to: https://api.custom-domain.com -->
```

## 🔍 **BEFORE vs AFTER BEHAVIOR**

### **❌ BEFORE (Problematic)**
```
Website: https://example.com
Tracker API calls: https://example.com/api/v1/analytics/event/batch
Result: 404 Error (example.com doesn't have Seentics API)

Website: https://mystore.com  
Tracker API calls: https://mystore.com/api/v1/workflows/active
Result: 404 Error (mystore.com doesn't have Seentics API)
```

### **✅ AFTER (Fixed)**
```
Website: https://example.com
Tracker API calls: https://www.api.seentics.com/api/v1/analytics/event/batch
Result: ✅ Success (correct API endpoint)

Website: https://mystore.com
Tracker API calls: https://www.api.seentics.com/api/v1/workflows/active  
Result: ✅ Success (correct API endpoint)
```

## 🛡️ **CORS COMPATIBILITY**

The recent CORS updates in the gateway now support these API calls:

```javascript
// This will work from any domain now
fetch('https://www.api.seentics.com/api/v1/analytics/event', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key',
    'X-Site-ID': 'your-site-id'
  },
  body: JSON.stringify({
    event_type: 'pageview',
    page: '/home'
  })
});
```

## 📊 **TESTING VERIFICATION**

### **Test Cases**
1. **Local Development**: Verify localhost uses `http://localhost:8080`
2. **Production Integration**: Verify any domain uses `https://www.api.seentics.com`
3. **Custom Configuration**: Verify override works with `window.SEENTICS_CONFIG`
4. **CORS Headers**: Verify tracking requests succeed from any domain

### **Expected Results**
```bash
# From any website
curl -H "Origin: https://example.com" \
     -X POST \
     https://www.api.seentics.com/api/v1/analytics/event

# Should return:
# Access-Control-Allow-Origin: https://example.com
# Status: 200 OK (with valid API key)
```

## 🎯 **BENEFITS**

1. **Correct API Routing**: All tracking data goes to the right endpoint
2. **Universal Compatibility**: Works from any website domain
3. **Simplified Integration**: No domain-specific configuration needed
4. **Better Reliability**: No more 404 errors from incorrect API URLs
5. **CORS Compatibility**: Works with the new CORS configuration

## 📁 **FILES MODIFIED**

- ✅ `frontend/public/trackers/tracker.js` - Analytics tracker API URL
- ✅ `frontend/public/trackers/workflow-tracker.js` - Workflow tracker API URL  
- ✅ `frontend/public/trackers/funnel-tracker.js` - Funnel tracker API URL

## 🚨 **DEPLOYMENT NOTES**

### **Immediate Effect**
- Changes take effect as soon as trackers are redeployed
- Existing integrations will automatically use correct API URL
- No breaking changes for current users

### **Backward Compatibility**
- ✅ Custom `window.SEENTICS_CONFIG.apiHost` still works
- ✅ Development environment (localhost) unchanged
- ✅ All existing functionality preserved

## 🎉 **SUMMARY**

All tracking scripts now correctly point to `https://www.api.seentics.com` in production, ensuring reliable data collection from any website while maintaining development flexibility and custom configuration options.
