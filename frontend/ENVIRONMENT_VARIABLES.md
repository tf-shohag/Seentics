# Frontend Environment Variables for Vercel Deployment

This file contains all the environment variables needed for the Seentics frontend when deploying to Vercel.

## Required Environment Variables

### API Configuration
```bash
# Backend API URL - Point to your deployed backend
NEXT_PUBLIC_API_URL=https://api.seentics.com

# Frontend URL - Your Vercel deployment URL
NEXT_PUBLIC_FRONTEND_URL=https://your-app-name.vercel.app

# API Base URL (legacy - used in admin page)
NEXT_PUBLIC_API_BASE_URL=https://api.seentics.com/api/v1
```

### OAuth Configuration
```bash
# Google OAuth Client ID
NEXT_PUBLIC_GOOGLE_CLIENT_ID=your-google-client-id

# GitHub OAuth Client ID  
NEXT_PUBLIC_GITHUB_CLIENT_ID=your-github-client-id

# OAuth Redirect URI (should match your Vercel URL)
NEXT_PUBLIC_REDIRECT_URI=https://your-app-name.vercel.app
```

### Admin Configuration
```bash
# Admin access code for /admin page
NEXT_PUBLIC_ADMIN_CODE=your-secure-admin-code
```

### Site Configuration
```bash
# Default site ID (optional)
NEXT_PUBLIC_DEFAULT_SITE_ID=

# Support email
NEXT_PUBLIC_SUPPORT_EMAIL=support@seentics.com

# Resend from email
NEXT_PUBLIC_RESEND_FROM_EMAIL=noreply@seentics.com
```

### Feature Flags
```bash
# Enable/disable email support (default: true)
NEXT_PUBLIC_ENABLE_EMAIL_SUPPORT=true

# Enable/disable OAuth (default: true)
NEXT_PUBLIC_ENABLE_OAUTH=true

# Enable/disable analytics (default: true)
NEXT_PUBLIC_ENABLE_ANALYTICS=true

# Enable/disable funnel tracking (default: true)
NEXT_PUBLIC_ENABLE_FUNNEL_TRACKING=true

# Enable/disable privacy features (default: true)
NEXT_PUBLIC_ENABLE_PRIVACY_FEATURES=true
```

### Build Information (Optional)
```bash
# App version
NEXT_PUBLIC_VERSION=1.0.0

# Build timestamp (auto-generated if not provided)
NEXT_PUBLIC_BUILD_TIME=2024-01-01T00:00:00.000Z
```

## Vercel Deployment Instructions

1. **In your Vercel dashboard:**
   - Go to your project settings
   - Navigate to "Environment Variables"
   - Add each variable from above with your actual values

2. **Important URLs to update:**
   - Replace `https://your-app-name.vercel.app` with your actual Vercel URL
   - Replace `https://api.seentics.com` with your actual backend API URL

3. **OAuth Setup:**
   - Update your Google OAuth app's authorized redirect URIs to include your Vercel URL
   - Update your GitHub OAuth app's callback URL to include your Vercel URL

4. **Backend CORS Configuration:**
   - Ensure your backend services are configured to accept requests from your Vercel URL
   - Update the CORS_ORIGIN environment variables in your Docker Compose files

## Security Notes

- Never commit actual environment variable values to version control
- Use strong, unique values for NEXT_PUBLIC_ADMIN_CODE
- Ensure OAuth client secrets are kept secure and never exposed in frontend code
- All NEXT_PUBLIC_ variables are exposed to the client-side code
