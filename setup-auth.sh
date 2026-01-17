#!/bin/bash

# Setup script for gopdfsuit with Google OAuth authentication

set -e

echo "🔧 Setting up gopdfsuit with Google OAuth..."
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Step 1: Setup frontend environment
echo "📝 Step 1: Setting up frontend environment..."
cd frontend

if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "✅ Created frontend/.env from .env.example"
    echo "⚠️  Please edit frontend/.env and add your VITE_GOOGLE_CLIENT_ID"
    echo ""
else
    echo "ℹ️  frontend/.env already exists, skipping..."
fi

# Step 2: Install frontend dependencies
echo "📦 Step 2: Installing frontend dependencies..."
if command -v npm &> /dev/null; then
    npm install
    echo "✅ Frontend dependencies installed"
else
    echo "⚠️  npm not found, skipping frontend dependency installation"
fi

cd ..

# Step 3: Install Go dependencies
echo "📦 Step 3: Installing Go dependencies..."
if command -v go &> /dev/null; then
    go mod tidy
    echo "✅ Go dependencies installed"
else
    echo "❌ Error: Go is not installed"
    exit 1
fi

# Step 4: Print instructions
echo ""
echo "✅ Setup complete!"
echo ""
echo "📋 Next steps:"
echo ""
echo "1. Get Google OAuth Client ID:"
echo "   - Go to https://console.cloud.google.com/apis/credentials"
echo "   - Create OAuth 2.0 Client ID (Web application)"
echo "   - Add authorized origins: http://localhost:5173, http://localhost:8080"
echo "   - Copy the Client ID"
echo ""
echo "2. Configure frontend/.env:"
echo "   VITE_GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com"
echo ""
echo "3. Run locally (no auth required):"
echo "   Terminal 1: go run cmd/gopdfsuit/main.go"
echo "   Terminal 2: cd frontend && npm run dev"
echo "   Open: http://localhost:5173"
echo ""
echo "4. Deploy to Cloud Run:"
echo "   - Update frontend/.env with Cloud Run settings"
echo "   - Run: cd frontend && npm run build"
echo "   - Copy: rm -rf ../docs/* && cp -r dist/* ../docs/"
echo "   - Deploy: gcloud run deploy gopdfsuit --source . --region us-central1"
echo ""
echo "📚 For detailed instructions, see docs/AUTHENTICATION.md"
echo ""
