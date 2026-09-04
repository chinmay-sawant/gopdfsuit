# Troubleshooting Google Authentication

## Issue: "Invalid ID token: audience provided does not match aud claim in the JWT"

### Symptoms
- You receive a 401 Unauthorized error when making API requests.
- The error response contains:
  ```json
  {
      "details": "idtoken: audience provided does not match aud claim in the JWT",
      "error": "Invalid ID token"
  }
  ```

### Cause
This error occurs when the backend validates the Google ID token against an expected "audience" (aud), but the token's audience doesn't match.

- **Frontend:** When you sign in with Google on the frontend, the ID token issued has your **Google Client ID** as the audience (`aud` claim).
- **Backend:** The backend was previously expecting the **Cloud Run Service URL** as the audience.

### Solution

The backend code has been updated to check for `GOOGLE_CLIENT_ID` as a valid audience. You need to ensure this environment variable is set for your backend service.

#### 1. Update Backend Environment Variables

**For Cloud Run:**
You need to set the `GOOGLE_CLIENT_ID` environment variable in your Cloud Run service configuration.

```bash
gcloud run services update gopdfsuit \
  --update-env-vars GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com \
  --region us-central1
```

Replace `your-client-id.apps.googleusercontent.com` with your actual Google Client ID (the same one used in the frontend).

**For Local Testing (if simulating Cloud Run):**
If you are running locally with `K_SERVICE` set (to simulate Cloud Run), make sure to also export `GOOGLE_CLIENT_ID`.

```bash
export K_SERVICE=gopdfsuit
export GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
go run cmd/gopdfsuit/main.go
```

#### 2. Verify Frontend Configuration

Ensure your frontend is using the correct Client ID in `.env`:
```
VITE_GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
```

### How it works now
The backend middleware now checks for the audience in this order:
1. `GOOGLE_OAUTH_AUDIENCE` (Custom audience, if set)
2. `GOOGLE_CLIENT_ID` (Standard for Google Sign-In)
3. `CLOUD_RUN_SERVICE_URL` (Fallback)

By setting `GOOGLE_CLIENT_ID` on the backend, the validation will match the token issued to your frontend application.
