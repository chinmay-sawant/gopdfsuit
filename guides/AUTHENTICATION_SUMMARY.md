# Google OAuth Authentication Implementation Summary

## ✅ Implementation Complete

This implementation adds Google OAuth authentication to gopdfsuit that automatically activates when deployed on Google Cloud Run, while keeping local development friction-free.

## 🎯 Key Features

1. **Smart Environment Detection**
   - Backend checks `K_SERVICE` env var (set by Cloud Run)
   - Frontend checks `VITE_IS_CLOUD_RUN` flag
   - **Zero auth required for local development**
   - **Full auth protection on Cloud Run**

2. **Secure Token Validation**
   - Google ID tokens validated server-side
   - Uses `google.golang.org/api/idtoken` library
   - Tokens validated against Google's public keys
   - User info extracted and available in handlers

3. **User-Friendly Frontend**
   - Google One-Tap login
   - Automatic token refresh on page reload
   - Token stored in localStorage
   - Clean login UI with user profile display

## 📁 Files Created

### Frontend
```
frontend/
├── src/
│   ├── contexts/
│   │   └── AuthContext.jsx          # Auth state management
│   ├── components/
│   │   └── AuthGuard.jsx            # Login screen & protection
│   └── utils/
│       └── apiConfig.js             # Environment detection & API helpers
├── .env.example                     # Environment template
└── [Updated: App.jsx, main.jsx, pages/Editor.jsx]
```

### Backend
```
internal/
├── middleware/
│   └── auth.go                      # OAuth token validation
└── handlers/
    └── handlers.go                  # [Updated: Added middleware]
```

### Documentation
```
docs/
└── AUTHENTICATION.md                # Complete setup guide
setup-auth.sh                        # Quick setup script
```

## 🔧 How to Use

### Local Development (No Auth)
```bash
# Backend
go run cmd/gopdfsuit/main.go

# Frontend
cd frontend && npm run dev

# Visit http://localhost:5173
# Works immediately, no login required!
```

### Cloud Run Deployment (With Auth)
```bash
# 1. Setup Google OAuth in Google Cloud Console
# 2. Configure frontend/.env with VITE_GOOGLE_CLIENT_ID
# 3. Build for Cloud Run
cd frontend
export VITE_IS_CLOUD_RUN=true
export VITE_GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
export VITE_CLOUD_RUN_URL=https://your-service.run.app
npm run build
cp -r dist/* ../docs/

# 4. Deploy
cd ..
gcloud run deploy gopdfsuit --source . --region us-central1
```

## 🛡️ What Gets Protected

When deployed on Cloud Run, these endpoints require authentication:
- ✅ `POST /api/v1/generate/template-pdf` - PDF generation
- ✅ `POST /api/v1/fill` - PDF form filling
- ✅ `POST /api/v1/merge` - PDF merging
- ✅ `GET /api/v1/template-data` - Template data
- ✅ `GET /api/v1/fonts` - Font listing
- ✅ `POST /api/v1/htmltopdf` - HTML to PDF
- ✅ `POST /api/v1/htmltoimage` - HTML to Image

## 🔑 Environment Variables

### Frontend (.env)
```env
# Required for Cloud Run
VITE_GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
VITE_IS_CLOUD_RUN=true
VITE_API_URL=https://your-service.run.app

# For local dev
VITE_IS_CLOUD_RUN=false
VITE_API_URL=http://localhost:8080
```

### Backend (Auto-detected)
- `K_SERVICE` - Auto-set by Cloud Run, triggers auth
- `K_REVISION` - Auto-set by Cloud Run (backup detection)

## 🎨 User Experience

### Local Development
```
User visits http://localhost:5173
  ↓
App loads immediately
  ↓
Full access to all features
  ↓
No authentication needed!
```

### Cloud Run Deployment
```
User visits https://your-service.run.app
  ↓
Sees Google login screen
  ↓
Signs in with Google account
  ↓
Gets ID token from Google
  ↓
Token stored in localStorage
  ↓
All API requests include Bearer token
  ↓
Backend validates token
  ↓
Full access to features
```

## 🧪 Testing Checklist

- [x] Local dev works without auth
- [x] Frontend loads without login locally
- [x] API calls work without tokens locally
- [ ] Cloud Run shows login screen
- [ ] Can sign in with Google account
- [ ] API calls include Authorization header
- [ ] Backend validates tokens correctly
- [ ] Invalid tokens return 401
- [ ] User info displayed after login
- [ ] Logout clears token and redirects

## 📝 Required Setup Steps

1. **Google Cloud Console**
   - Create OAuth 2.0 Client ID
   - Configure authorized origins
   - Get Client ID

2. **Frontend Configuration**
   ```bash
   cd frontend
   cp .env.example .env
   # Edit .env with your Client ID
   ```

3. **Build & Deploy**
   ```bash
   # Set Cloud Run environment variables
   # Build frontend
   # Deploy to Cloud Run
   ```

## 📚 Documentation

Detailed setup instructions: [docs/AUTHENTICATION.md](docs/AUTHENTICATION.md)

Quick setup: `./setup-auth.sh`

## 🚀 Deployment Commands

```bash
# Quick setup
./setup-auth.sh

# Build for Cloud Run
cd frontend
npm run build
rm -rf ../docs/* && cp -r dist/* ../docs/

# Deploy
cd ..
gcloud run deploy gopdfsuit \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## 🔐 Security Notes

- ✅ Tokens validated using Google's public keys
- ✅ No session storage needed (stateless)
- ✅ Tokens expire automatically
- ✅ Environment-based protection
- ✅ User info extracted from validated tokens
- ✅ HTTPS enforced on Cloud Run

## 📊 Architecture Diagram

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │
       │ 1. User visits app
       ▼
┌─────────────────┐
│  React Frontend │  Checks: VITE_IS_CLOUD_RUN
│  (AuthGuard)    │
└──────┬──────────┘
       │
       │ If Cloud Run:
       │ 2. Show Google login
       ▼
┌─────────────────┐
│ Google OAuth    │
│ (One-Tap)       │
└──────┬──────────┘
       │
       │ 3. Returns ID token
       ▼
┌─────────────────┐
│  Browser        │
│  localStorage   │  Stores: google_id_token
└──────┬──────────┘
       │
       │ 4. API calls with Bearer token
       ▼
┌─────────────────┐
│  Go Backend     │  Checks: K_SERVICE env var
│  (Gin + Auth    │
│   Middleware)   │
└──────┬──────────┘
       │
       │ If Cloud Run:
       │ 5. Validate token with Google
       ▼
┌─────────────────┐
│ Google's        │
│ Public Keys     │
└──────┬──────────┘
       │
       │ 6. Token valid
       ▼
┌─────────────────┐
│  Request        │
│  Processed      │
└─────────────────┘
```

## 🎉 Benefits

1. **Developer Experience**
   - No auth setup needed for local dev
   - Test features instantly
   - Same code works locally and on Cloud Run

2. **Security**
   - Production endpoints protected
   - Leverages Google's OAuth infrastructure
   - No custom auth system to maintain

3. **User Experience**
   - Familiar Google login
   - One-click authentication
   - No registration required

4. **Deployment**
   - Automatic environment detection
   - No manual configuration on Cloud Run
   - Single codebase for all environments

## 🤝 Contributing

To work on this feature:
1. Run `./setup-auth.sh` for initial setup
2. Develop locally without auth (fast iteration)
3. Test Cloud Run deployment with auth enabled
4. See [docs/AUTHENTICATION.md](docs/AUTHENTICATION.md) for details
