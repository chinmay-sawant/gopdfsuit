# GoPdfSuit Frontend

This directory contains the React frontend for GoPdfSuit, built with Vite.

## 🚀 Quick Start

### Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

The development server will run on `http://localhost:3000` and proxy API requests to the Go backend on `http://localhost:8080`.

### Production Build

```bash
# Build for production
npm run build
```

The production build will be output to `../docs/` and served by the Go application.

### Preview Production Build

```bash
# Preview production build locally
npm run preview
```

## 📁 Project Structure

```
frontend/
├── src/
│   ├── components/        # Reusable React components
│   │   └── Navbar.jsx    # Navigation component
│   ├── pages/            # Page components for routes
│   │   ├── Home.jsx      # Homepage with README content
│   │   ├── Viewer.jsx    # PDF viewer and template processor
│   │   ├── Editor.jsx    # PDF template editor (placeholder)
│   │   ├── Merge.jsx     # PDF merge tool
│   │   ├── Filler.jsx    # PDF form filler
│   │   ├── HtmlToPdf.jsx # HTML to PDF converter
│   │   └── HtmlToImage.jsx # HTML to Image converter
│   ├── App.jsx           # Main app component with routing
│   ├── main.jsx          # React entry point
│   └── index.css         # Global styles
├── package.json          # Dependencies and scripts
├── vite.config.js        # Vite configuration
└── index.html            # HTML template
```

## 🛠️ Technology Stack

- **React 18** - UI library
- **React Router 6** - Client-side routing
- **Vite 5** - Build tool and dev server
- **Lucide React** - Icons
- **Modern CSS** - Styling with CSS custom properties

## 🎯 Features

- **Responsive Design** - Works on desktop, tablet, and mobile
- **Modern UI** - Clean, gradient-based design with glassmorphism effects
- **Client-side Routing** - Fast navigation between pages
- **API Integration** - Seamless communication with Go backend
- **Live Previews** - Real-time PDF and image generation
- **Drag & Drop** - File uploads with visual feedback

## 🔧 Configuration

### Vite Config

The Vite configuration includes:
- React plugin for JSX support
- Build output to `../docs/`
- Proxy configuration for API and static asset requests
- Production optimizations

### API Proxy

During development, requests to `/api/*` and `/static/*` are proxied to the Go backend at `http://localhost:8080`.

## 🚀 Deployment

The frontend is built as a Single Page Application (SPA) and served by the Go backend. All routes are handled client-side except for API endpoints.

### Build Process

1. `npm run build` generates production assets in `../docs/`
2. Go application serves `index.html` for all non-API routes
3. Static assets are served from `/assets/*`

## 📝 Development Notes

- The build process is integrated with the Go application
- Static files are served directly by Gin from the `docs/` directory
- All frontend routes use React Router for client-side navigation
- The Go backend serves as a fallback for the SPA routing

## 🔄 Migration from Static Files

This React frontend replaces the previous static HTML/CSS/JS files while maintaining the same URL structure:

- `/` → Home page (new)
- `/viewer` → PDF Viewer (migrated)
- `/editor` → Template Editor (placeholder)
- `/merge` → PDF Merge (migrated)
- `/filler` → Form Filler (migrated)
- `/htmltopdf` → HTML to PDF (migrated)
- `/htmltoimage` → HTML to Image (migrated)

All API endpoints remain unchanged and compatible.