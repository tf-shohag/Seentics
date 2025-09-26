# 🌐 Gateway CORS Configuration for Public Tracking

## 🎯 **OBJECTIVE**
Enable any website to send tracking data to Seentics analytics by allowing cross-origin requests for public tracking endpoints while maintaining security for protected routes.

## 🔧 **IMPLEMENTATION**

### **Modified CORS Middleware**
The gateway CORS middleware now differentiates between:
1. **Public tracking endpoints** - Allow any domain
2. **Protected dashboard endpoints** - Restricted to configured domains

### **Code Changes**

#### **1. Enhanced CORS Logic**
```go
// Check if this is a public tracking endpoint
isPublicTrackingEndpoint := isPublicTrackingRoute(path)

if isPublicTrackingEndpoint {
    // Allow any origin for public tracking endpoints
    if origin != "" {
        w.Header().Set("Access-Control-Allow-Origin", origin)
    } else {
        w.Header().Set("Access-Control-Allow-Origin", "*")
    }
    w.Header().Set("Access-Control-Allow-Credentials", "false") // No credentials for public endpoints
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,X-API-Key,X-Site-ID,X-Domain")
    w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Range")
    w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours for public endpoints
} else {
    // Existing protected endpoint logic...
}
```

#### **2. Public Tracking Route Detection**
```go
func isPublicTrackingRoute(path string) bool {
    // Remove query parameters for path matching
    cleanPath := strings.Split(path, "?")[0]

    // Public tracking endpoints that should allow any domain
    publicTrackingPrefixes := []string{
        "/api/v1/track",
        "/api/v1/analytics/event",
        "/api/v1/analytics/event/batch",
        "/api/v1/analytics/track",
        "/api/v1/workflows/analytics/track",
        "/api/v1/workflows/analytics/track/batch",
        "/api/v1/workflows/site/",
        "/api/v1/workflows/active",           // Active workflows for tracker (public)
        "/api/v1/workflows/execution/action", // Workflow execution (public with validation)
        "/api/v1/funnels/track",              // Funnel event tracking (public)
        "/api/v1/funnels/active",             // Active funnels for tracker (public)
    }

    for _, prefix := range publicTrackingPrefixes {
        if strings.HasPrefix(cleanPath, prefix) {
            return true
        }
    }

    return false
}
```

## 📋 **PUBLIC TRACKING ENDPOINTS**

### **Analytics Tracking**
- `POST /api/v1/analytics/event` - Single event tracking
- `POST /api/v1/analytics/event/batch` - Batch event tracking
- `POST /api/v1/analytics/track` - General analytics tracking
- `GET /api/v1/track` - Simple tracking endpoint

### **Workflow Tracking**
- `POST /api/v1/workflows/analytics/track` - Single workflow event
- `POST /api/v1/workflows/analytics/track/batch` - Batch workflow events
- `GET /api/v1/workflows/site/{siteId}/active` - Get active workflows
- `GET /api/v1/workflows/active` - Get active workflows (query param)
- `POST /api/v1/workflows/execution/action` - Execute workflow actions

### **Funnel Tracking**
- `POST /api/v1/funnels/track` - Funnel event tracking
- `GET /api/v1/funnels/active` - Get active funnels for tracking

## 🔒 **SECURITY CONSIDERATIONS**

### **Public Endpoints Security**
- ✅ **No credentials allowed** (`Access-Control-Allow-Credentials: false`)
- ✅ **Limited HTTP methods** (GET, POST, OPTIONS only)
- ✅ **Site validation** still enforced by backend services
- ✅ **API key validation** still required for tracking
- ✅ **Rate limiting** still applies

### **Protected Endpoints Security**
- ✅ **Domain whitelist maintained** for dashboard/admin routes
- ✅ **Credentials allowed** for authenticated requests
- ✅ **Full HTTP methods** (GET, POST, PUT, DELETE, etc.)
- ✅ **JWT authentication** still required

## 🌍 **CORS HEADERS COMPARISON**

### **Public Tracking Endpoints**
```http
Access-Control-Allow-Origin: https://example.com (or *)
Access-Control-Allow-Credentials: false
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,X-API-Key,X-Site-ID,X-Domain
Access-Control-Max-Age: 86400
```

### **Protected Dashboard Endpoints**
```http
Access-Control-Allow-Origin: https://seentics.com (whitelisted only)
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
Access-Control-Allow-Headers: DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization,X-API-Key,X-Site-ID,X-Domain
Access-Control-Max-Age: 1728000
```

## 🚀 **USAGE EXAMPLES**

### **JavaScript Tracking from Any Website**
```javascript
// This will now work from any domain
fetch('https://api.seentics.com/api/v1/analytics/event', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key',
    'X-Site-ID': 'your-site-id'
  },
  body: JSON.stringify({
    event_type: 'pageview',
    page: '/home',
    visitor_id: 'visitor123'
  })
});
```

### **Workflow Tracking from Any Website**
```javascript
// Get active workflows
fetch('https://api.seentics.com/api/v1/workflows/active?siteId=your-site-id', {
  method: 'GET',
  headers: {
    'X-API-Key': 'your-api-key'
  }
})
.then(response => response.json())
.then(workflows => {
  // Process workflows
});
```

### **Batch Event Tracking**
```javascript
// Send multiple events at once
fetch('https://api.seentics.com/api/v1/analytics/event/batch', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key'
  },
  body: JSON.stringify({
    siteId: 'your-site-id',
    events: [
      { event_type: 'pageview', page: '/home' },
      { event_type: 'click', element: 'button' }
    ]
  })
});
```

## 🔍 **VALIDATION & TESTING**

### **Test CORS Headers**
```bash
# Test public tracking endpoint
curl -H "Origin: https://example.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     https://api.seentics.com/api/v1/analytics/event

# Expected response headers:
# Access-Control-Allow-Origin: https://example.com
# Access-Control-Allow-Methods: GET, POST, OPTIONS
# Access-Control-Allow-Credentials: false
```

### **Test Protected Endpoint**
```bash
# Test protected dashboard endpoint
curl -H "Origin: https://unauthorized-domain.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS \
     https://api.seentics.com/api/v1/user/profile

# Expected: No CORS headers (blocked)
```

## 📊 **MONITORING**

### **Metrics to Track**
1. **CORS preflight requests** for tracking endpoints
2. **Cross-origin tracking requests** from various domains
3. **Failed CORS requests** (should be minimal for tracking)
4. **Performance impact** of CORS processing

### **Logs to Monitor**
- CORS header setting for public vs protected routes
- Origin validation for protected endpoints
- Public tracking endpoint usage patterns

## ⚠️ **IMPORTANT NOTES**

### **Security Boundaries**
- **Public tracking** ≠ **No validation**
- API keys and site IDs still required
- Backend services still validate requests
- Rate limiting still applies

### **Cache Considerations**
- Public endpoints have shorter cache time (24 hours)
- Protected endpoints have longer cache time (20 days)
- Browsers will cache CORS preflight responses

### **Deployment**
- Changes take effect immediately after gateway restart
- No environment variable changes needed
- Backward compatible with existing integrations

## 🎯 **BENEFITS**

1. **Universal Tracking**: Any website can integrate Seentics tracking
2. **Simplified Integration**: No CORS configuration needed by users
3. **Better Performance**: Reduced preflight request failures
4. **Maintained Security**: Protected routes remain secure
5. **Scalability**: Supports tracking from unlimited domains

## 🔄 **BACKWARD COMPATIBILITY**

- ✅ **Existing integrations continue to work**
- ✅ **No breaking changes to API**
- ✅ **Protected routes maintain same security**
- ✅ **Dashboard functionality unchanged**

This implementation enables Seentics to be used as a true third-party analytics service while maintaining security for administrative functions.
