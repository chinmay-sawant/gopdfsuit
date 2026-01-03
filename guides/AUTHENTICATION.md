# Google OAuth Authentication Setup for Cloud Run

This document explains how to set up Google OAuth authentication for the gopdfsuit application when deployed on Google Cloud Run.

## Overview

The application now includes Google OAuth authentication that:
- **Only activates on Google Cloud Run** (checks `K_SERVICE` environment variable)
- Allows local development without authentication
- Protects all API endpoints when deployed to Cloud Run
- Uses Google ID tokens for secure, serverless authentication

## Architecture

### Frontend (React)
- **AuthContext**: Manages authentication state and token storage
- **AuthGuard**: Protects routes and shows login page when needed
- **API Config**: Detects Cloud Run environment and conditionally adds auth headers

### Backend (Go)
- **Auth Middleware**: Validates Google ID tokens using `google.golang.org/api/idtoken`
- **Cloud Run Detection**: Checks `K_SERVICE` environment variable
- **Automatic Protection**: All `/api/v1/*` endpoints require auth on Cloud Run

## Setup Instructions

### 1. Google Cloud Console Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create or select your project
3. Enable the **Google+ API** (required for OAuth)
4. Go to **APIs & Services > Credentials**
5. Click **Create Credentials > OAuth 2.0 Client ID**
6. Configure OAuth consent screen if prompted
7. Choose **Web application** as application type
8. Add authorized origins:
   - Local: `http://localhost:5173` (Vite dev server)
   - Local: `http://localhost:8080` (Go server)
   - Cloud Run: `https://your-service-name.run.app`
9. Add authorized redirect URIs:
   - `http://localhost:5173`
   - `https://your-service-name.run.app`
10. Copy the **Client ID** (looks like `xxxx.apps.googleusercontent.com`)

### 2. Frontend Configuration

Create a `.env` file in the `frontend/` directory:

```bash
# Copy from .env.example
cp frontend/.env.example frontend/.env
```

Edit `frontend/.env`:

```env
# Google OAuth Client ID (REQUIRED)
VITE_GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com

# For local development (no auth required)
VITE_API_URL=http://localhost:8080
VITE_IS_CLOUD_RUN=false

# For Cloud Run deployment (auth required)
# VITE_IS_CLOUD_RUN=true
# VITE_ENVIRONMENT=cloudrun
# VITE_CLOUD_RUN_URL=https://your-service-name.run.app
# VITE_API_URL=https://your-service-name.run.app
```

### 3. Backend Configuration

The backend automatically detects Cloud Run via the `K_SERVICE` environment variable. No configuration needed for local development.

For Cloud Run, you should set the `GOOGLE_CLIENT_ID` environment variable to match your frontend Client ID. This ensures the backend validates tokens correctly.

```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
```

Optionally, you can set (Cloud Run auto-provides these):
```bash
GOOGLE_OAUTH_AUDIENCE=https://your-service-name.run.app
CLOUD_RUN_SERVICE_URL=https://your-service-name.run.app
```

### 4. Build and Deploy

#### Local Development (No Auth)
```bash
# Backend
cd /path/to/gopdfsuit
go run cmd/gopdfsuit/main.go

# Frontend (separate terminal)
cd frontend
npm run dev
```

Visit `http://localhost:5173` - no authentication required!

#### Cloud Run Deployment

##### Build Frontend with Cloud Run Config
```bash
cd frontend

# Update .env for production
cat > .env << EOF
VITE_GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
VITE_IS_CLOUD_RUN=true
VITE_ENVIRONMENT=cloudrun
VITE_CLOUD_RUN_URL=https://your-service-name.run.app
VITE_API_URL=https://your-service-name.run.app
EOF

# Build
npm run build

# Copy build to docs/
rm -rf ../docs/*
cp -r dist/* ../docs/
```

##### Deploy to Cloud Run
```bash
cd ..

# Build and deploy
gcloud run deploy gopdfsuit \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

# Note: --allow-unauthenticated is for the service itself
# Our app handles authentication at the application layer
```

## How It Works

### Local Development
1. `K_SERVICE` environment variable is not set
2. Backend middleware checks for Cloud Run and skips authentication
3. Frontend detects `VITE_IS_CLOUD_RUN=false` and doesn't require login
4. All API calls work without tokens

### Cloud Run Deployment
1. Cloud Run sets `K_SERVICE` environment variable automatically
2. Frontend shows Google login screen on app load
3. User signs in with Google account
4. Frontend receives ID token from Google
5. All API requests include `Authorization: Bearer <token>` header
6. Backend validates token using Google's public keys
7. If valid, request proceeds; if invalid, returns 401 Unauthorized

## Security Features

✅ **Serverless Authentication**: No session storage, tokens validated on each request  
✅ **Google's Security**: Leverages Google's OAuth infrastructure  
✅ **Token Validation**: Backend verifies tokens using Google's public keys  
✅ **Environment Detection**: Automatically enables/disables based on deployment  
✅ **Token Storage**: ID tokens stored in browser localStorage  
✅ **User Info**: Email, name, and picture available in backend context  

## API Endpoints

All endpoints under `/api/v1/*` are protected when running on Cloud Run:

- `POST /api/v1/generate/template-pdf` - Generate PDF from template
- `POST /api/v1/fill` - Fill PDF forms
- `POST /api/v1/merge` - Merge PDFs
- `GET /api/v1/template-data` - Get template data
- `GET /api/v1/fonts` - List available fonts
- `POST /api/v1/htmltopdf` - Convert HTML to PDF
- `POST /api/v1/htmltoimage` - Convert HTML to Image

## Accessing User Information (Backend)

In your handlers, you can access authenticated user info:

```go
import "github.com/chinmay-sawant/gopdfsuit/internal/middleware"

func YourHandler(c *gin.Context) {
    // Get user email
    email, exists := middleware.GetUserEmail(c)
    if exists {
        log.Printf("Request from: %s", email)
    }

    // Get all user info
    userInfo := middleware.GetUserInfo(c)
    // userInfo contains: email, name, picture, sub (Google user ID)
    
    // Your handler logic...
}
```

## Troubleshooting

### "Authorization header required"
- Check that `VITE_GOOGLE_CLIENT_ID` is set in frontend `.env`
- Verify you're logged in (check localStorage for `google_id_token`)
- Try logging out and back in

### "Invalid ID token"
- Token may have expired (refresh the page to get new token)
- Check that `GOOGLE_OAUTH_AUDIENCE` matches your Cloud Run URL
- Verify OAuth client is configured for your Cloud Run domain

### Authentication not required locally
- This is expected! Check that `K_SERVICE` is not set
- Frontend should have `VITE_IS_CLOUD_RUN=false`

### Can't access API on Cloud Run
- Ensure frontend was built with Cloud Run environment variables
- Check browser console for authentication errors
- Verify Cloud Run service is deployed and accessible

## Environment Variables Reference

### Frontend (.env)
| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_GOOGLE_CLIENT_ID` | Yes | OAuth client ID from Google Console |
| `VITE_IS_CLOUD_RUN` | No | Set to `true` for Cloud Run builds |
| `VITE_ENVIRONMENT` | No | Set to `cloudrun` for Cloud Run |
| `VITE_API_URL` | No | Backend API base URL |
| `VITE_CLOUD_RUN_URL` | No | Cloud Run service URL |

### Backend (Environment)
| Variable | Required | Description |
|----------|----------|-------------|
| `K_SERVICE` | Auto | Set by Cloud Run, triggers auth |
| `K_REVISION` | Auto | Set by Cloud Run (backup check) |
| `GOOGLE_CLIENT_ID` | Yes | Expected audience (Client ID) for tokens |
| `GOOGLE_OAUTH_AUDIENCE` | Optional | Custom expected audience |
| `CLOUD_RUN_SERVICE_URL` | Optional | Service URL for validation (fallback) |

## Files Modified

### Frontend
- `src/contexts/AuthContext.jsx` - New: Authentication state management
- `src/components/AuthGuard.jsx` - New: Login UI and route protection
- `src/utils/apiConfig.js` - New: Environment detection and API helpers
- `src/pages/Editor.jsx` - Updated: Uses authenticated requests
- `src/App.jsx` - Updated: Conditional auth based on environment
- `src/main.jsx` - Updated: Wrapped with OAuth provider
- `.env.example` - New: Environment configuration template

### Backend
- `internal/middleware/auth.go` - New: OAuth token validation
- `internal/handlers/handlers.go` - Updated: Applied auth middleware
- `go.mod` - Updated: Added `google.golang.org/api/idtoken`

## Testing

### Test Locally (No Auth)
```bash
# Start backend
go run cmd/gopdfsuit/main.go

# Start frontend
cd frontend && npm run dev

# Access http://localhost:5173
# Should work without login!
```

### Test Cloud Run (With Auth)
1. Deploy to Cloud Run
2. Visit your Cloud Run URL
3. Should see Google login screen
4. Sign in with Google account
5. After login, full access to editor

## Next Steps

- Consider adding role-based access control (RBAC)
- Implement token refresh for long sessions
- Add audit logging for authenticated requests
- Set up authorized domains list for production
