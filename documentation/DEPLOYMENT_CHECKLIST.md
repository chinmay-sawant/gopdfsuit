# Google OAuth Deployment Checklist

## ✅ Pre-Deployment Checklist

### 1. Google Cloud Console Setup
- [ ✅ ] Created Google Cloud Project
- [ ✅ ] Enabled Google+ API
- [ ✅ ] Created OAuth 2.0 Client ID
- [✅  ] Added authorized JavaScript origins:
  - [ ] `http://localhost:5173` (dev)
  - [ ] `http://localhost:8080` (dev)
  - [ ] `https://your-service-name.run.app` (production)
- [ ✅] Copied Client ID (format: `xxx.apps.googleusercontent.com`)
- [✅ ] OAuth Consent Screen configured

### 2. Local Development Setup
- [ ] Installed dependencies: `npm install` in frontend/
- [ ] Installed Go dependencies: `go mod tidy`
- [ ] Created `frontend/.env` from `.env.example`
- [ ] Set `VITE_GOOGLE_CLIENT_ID` in `.env`
- [ ] Set `VITE_IS_CLOUD_RUN=false` in `.env`
- [ ] Tested local dev (no auth): `go run cmd/gopdfsuit/main.go` + `npm run dev`
- [ ] Verified app loads without login locally
- [ ] Verified API calls work without tokens locally

### 3. Code Review
- [ ] Reviewed [src/contexts/AuthContext.jsx](../frontend/src/contexts/AuthContext.jsx)
- [ ] Reviewed [src/components/AuthGuard.jsx](../frontend/src/components/AuthGuard.jsx)
- [ ] Reviewed [src/utils/apiConfig.js](../frontend/src/utils/apiConfig.js)
- [ ] Reviewed [internal/middleware/auth.go](../internal/middleware/auth.go)
- [ ] Verified all API endpoints use middleware in [handlers.go](../internal/handlers/handlers.go)
- [ ] No syntax errors: Run `npm run build` and `go build`

### 4. Environment Configuration
- [ ] Frontend `.env` ready for Cloud Run:
  ```env
  VITE_GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
  VITE_IS_CLOUD_RUN=true
  VITE_ENVIRONMENT=cloudrun
  VITE_CLOUD_RUN_URL=https://your-service-name.run.app
  VITE_API_URL=https://your-service-name.run.app
  ```
- [ ] Backend will auto-detect Cloud Run via `K_SERVICE`
- [ ] Optional: Set `GOOGLE_OAUTH_AUDIENCE` if needed

## 🚀 Deployment Steps

### Step 1: Build Frontend for Cloud Run
```bash
cd frontend

# Set environment for Cloud Run
export VITE_GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
export VITE_IS_CLOUD_RUN=true
export VITE_ENVIRONMENT=cloudrun
export VITE_CLOUD_RUN_URL=https://your-service-name.run.app
export VITE_API_URL=https://your-service-name.run.app

# Build
npm run build

# Verify build
ls -la dist/

# Copy to docs/
rm -rf ../docs/*
cp -r dist/* ../docs/

# Verify copy
ls -la ../docs/
```
- [ ] Build completed successfully
- [ ] `docs/` directory contains built files
- [ ] `docs/index.html` exists

### Step 2: Commit Changes (Optional)
```bash
cd ..
git add .
git commit -m "Add Google OAuth authentication for Cloud Run"
git push
```
- [ ] Changes committed
- [ ] Pushed to repository

### Step 3: Deploy to Cloud Run
```bash
# Deploy using gcloud
gcloud run deploy gopdfsuit \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --timeout 300 \
  --memory 1Gi

# Or specify project
gcloud run deploy gopdfsuit \
  --source . \
  --platform managed \
  --region us-central1 \
  --project your-project-id \
  --allow-unauthenticated
```
- [ ] Deployment started
- [ ] Build completed
- [ ] Service deployed
- [ ] Received service URL: `https://your-service-name.run.app`

### Step 4: Update OAuth Authorized Origins (if new URL)
- [ ] Go back to Google Cloud Console
- [ ] Navigate to OAuth Client ID settings
- [ ] Add new Cloud Run URL to authorized origins
- [ ] Save changes
- [ ] Wait ~5 minutes for propagation

### Step 5: Test Deployment
- [ ] Visit Cloud Run URL
- [ ] Sees Google login screen
- [ ] Can sign in with Google account
- [ ] After login, app loads correctly
- [ ] Can access all features
- [ ] API calls work (check browser Network tab)
- [ ] User profile displays correctly
- [ ] Can sign out

## 🧪 Post-Deployment Testing

### Manual Testing
- [ ] **Login Flow**
  - [ ] Visit app URL
  - [ ] Login screen appears
  - [ ] Google One-Tap dialog shows
  - [ ] Can sign in successfully
  - [ ] User info displays (name, email, picture)
  
- [ ] **API Access**
  - [ ] Load template data works
  - [ ] Generate PDF works
  - [ ] Download PDF works
  - [ ] All protected endpoints accessible
  
- [ ] **Token Validation**
  - [ ] Check browser console for errors
  - [ ] Verify Authorization header in Network tab
  - [ ] Token format: `Bearer eyJ...`
  
- [ ] **Session Persistence**
  - [ ] Refresh page stays logged in
  - [ ] Close and reopen browser stays logged in
  - [ ] Sign out clears session
  - [ ] After sign out, must log in again

### Browser DevTools Checks
```javascript
// Check in browser console:

// 1. Verify token stored
localStorage.getItem('google_id_token')
// Should return: "eyJhbG..."

// 2. Verify user info
JSON.parse(localStorage.getItem('google_user'))
// Should return: {email: "...", name: "...", picture: "..."}

// 3. Check API call headers (Network tab)
// Look for: Authorization: Bearer eyJhbG...
```

- [ ] Token present in localStorage
- [ ] User info present in localStorage
- [ ] API calls include Authorization header
- [ ] No console errors

### Backend Logs Check
```bash
# View Cloud Run logs
gcloud run logs read gopdfsuit --region us-central1 --limit 50

# Or in Cloud Console
# Go to: Cloud Run > gopdfsuit > Logs
```

Look for:
- [ ] No authentication errors
- [ ] Requests being processed
- [ ] Token validation succeeding
- [ ] User info being extracted

### Error Scenarios
- [ ] **Invalid Token**: Clear localStorage, try API call → Should redirect to login
- [ ] **Expired Token**: Wait 1 hour → Should redirect to login
- [ ] **No Token**: Open in incognito → Should show login screen
- [ ] **Wrong Audience**: Check logs for validation errors

## 📊 Monitoring

### Setup Alerts (Optional)
- [ ] Set up alert for 401 errors (authentication failures)
- [ ] Monitor token validation errors
- [ ] Track authentication success/failure rates

### Logging
- [ ] Enable detailed logging if needed
- [ ] Monitor user authentication events
- [ ] Track API usage per user

## 🔧 Troubleshooting

### Issue: Login screen doesn't appear
**Check:**
- [ ] `VITE_IS_CLOUD_RUN=true` in build environment
- [ ] Frontend built with correct environment variables
- [ ] Browser console for errors
- [ ] Network tab for failed requests

**Fix:**
```bash
# Rebuild with correct environment
cd frontend
export VITE_IS_CLOUD_RUN=true
export VITE_GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
npm run build
rm -rf ../docs/* && cp -r dist/* ../docs/
cd .. && gcloud run deploy gopdfsuit --source . --region us-central1
```

### Issue: "Authorization header required"
**Check:**
- [ ] Signed in successfully
- [ ] Token in localStorage: `localStorage.getItem('google_id_token')`
- [ ] Network tab shows Authorization header
- [ ] Frontend `getAuthHeaders()` working

**Fix:**
- Clear localStorage and sign in again
- Check browser console for errors
- Verify `useAuth()` hook is called in component

### Issue: "Invalid ID token"
**Check:**
- [ ] Token expired (tokens last ~1 hour)
- [ ] OAuth client configured for Cloud Run URL
- [ ] `GOOGLE_OAUTH_AUDIENCE` matches service URL (if set)
- [ ] Backend logs for validation errors

**Fix:**
- Sign out and sign in again (gets new token)
- Verify OAuth client authorized origins
- Check backend environment variables

### Issue: Works locally but not on Cloud Run
**Check:**
- [ ] Frontend built with `VITE_IS_CLOUD_RUN=true`
- [ ] Correct Cloud Run URL in environment variables
- [ ] OAuth client includes Cloud Run URL
- [ ] `K_SERVICE` environment variable present (Cloud Run auto-sets)

**Fix:**
- Rebuild frontend with production config
- Update OAuth client authorized origins
- Redeploy to Cloud Run

## 📝 Documentation References

- Full Guide: [docs/AUTHENTICATION.md](AUTHENTICATION.md)
- Summary: [AUTHENTICATION_SUMMARY.md](../AUTHENTICATION_SUMMARY.md)
- Quick Reference: [QUICK_AUTH_REFERENCE.md](../QUICK_AUTH_REFERENCE.md)
- Flow Diagrams: [docs/AUTH_FLOW_DIAGRAMS.md](AUTH_FLOW_DIAGRAMS.md)

## ✅ Final Verification

- [ ] App accessible at Cloud Run URL
- [ ] Google login required to access
- [ ] All features work after authentication
- [ ] Users can sign out successfully
- [ ] No errors in logs
- [ ] Performance is acceptable
- [ ] Security best practices followed

## 🎉 Deployment Complete!

Your gopdfsuit application is now deployed with Google OAuth authentication!

**Next Steps:**
- Share the URL with users
- Monitor logs for issues
- Set up regular security audits
- Consider adding additional features:
  - Role-based access control
  - Usage quotas per user
  - Audit logging
  - Rate limiting

---

**Need Help?**
- Check troubleshooting section above
- Review [docs/AUTHENTICATION.md](AUTHENTICATION.md)
- Check Cloud Run logs for errors
- Verify OAuth configuration in Google Console
