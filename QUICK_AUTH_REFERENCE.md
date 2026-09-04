# Google OAuth Quick Reference

## 🚀 Quick Start

### Get Google Client ID
1. Visit: https://console.cloud.google.com/apis/credentials
2. Create OAuth 2.0 Client ID
3. Add origins: `http://localhost:5173`, `http://localhost:8080`, `https://your-service.run.app`
4. Copy Client ID

### Configure
```bash
# frontend/.env
VITE_GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
VITE_IS_CLOUD_RUN=false  # true for Cloud Run
VITE_API_URL=http://localhost:8080
```

### Run Locally (No Auth)
```bash
go run cmd/gopdfsuit/main.go          # Terminal 1
cd frontend && npm run dev             # Terminal 2
```

### Deploy to Cloud Run (With Auth)
```bash
cd frontend
export VITE_IS_CLOUD_RUN=true
export VITE_GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
export VITE_CLOUD_RUN_URL=https://your-service.run.app
npm run build
rm -rf ../docs/* && cp -r dist/* ../docs/
cd .. && gcloud run deploy gopdfsuit --source . --region us-central1
```

## 🔑 How It Works

| Environment | Auth Required | Detection |
|-------------|--------------|-----------|
| Local Dev | ❌ No | `K_SERVICE` not set |
| Local Dev + `REQUIRE_AUTH=1` | ✅ Yes | Forced, for testing authenticated paths |
| Cloud Run | ✅ Yes | `K_SERVICE` set automatically |

## 📝 Environment Variables

### Frontend
- `VITE_GOOGLE_CLIENT_ID` - OAuth client ID (**required**)
- `VITE_IS_CLOUD_RUN` - `true` for Cloud Run, `false` for local
- `VITE_API_URL` - Backend URL

### Backend (Auto)
- `K_SERVICE` - Set by Cloud Run (triggers auth)
- `K_REVISION` - Set by Cloud Run (backup check)

### Backend (Optional)
- `REQUIRE_AUTH=1` - Force auth even without `K_SERVICE` (staging/local testing)

## 🎯 Protected Endpoints

All `/api/v1/*` endpoints require auth on Cloud Run:
- `/api/v1/generate/template-pdf`
- `/api/v1/fill`
- `/api/v1/merge`
- `/api/v1/template-data`
- `/api/v1/fonts`
- `/api/v1/htmltopdf`
- `/api/v1/htmltoimage`

## 🛠️ Troubleshooting

**"Authorization header required"**
→ Check `VITE_GOOGLE_CLIENT_ID` is set  
→ Sign in again

**"Invalid ID token"**
→ Token expired, refresh page  
→ Check OAuth client configured for Cloud Run URL

**No login screen locally**
→ Expected! `VITE_IS_CLOUD_RUN=false` disables auth

## 📚 Full Docs

- **Setup Guide**: [docs/AUTHENTICATION.md](docs/AUTHENTICATION.md)
- **Summary**: [AUTHENTICATION_SUMMARY.md](AUTHENTICATION_SUMMARY.md)
- **Setup Script**: `./setup-auth.sh`
