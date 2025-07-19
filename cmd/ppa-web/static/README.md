# PPA Web Frontend

A modern web frontend built with Preact + Zustand for visualizing PPA packet analysis.

## Features

- 📁 Upload and manage PCAP files
- 🔍 Real-time packet analysis with progress tracking
- 📊 Interactive packet timeline and message type visualization
- 🔎 Search through analysis documents and results
- 📝 Markdown document viewer with syntax highlighting
- 📱 Responsive design with Bootstrap

## Technology Stack

- **Frontend Framework:** Preact 10.x (React-compatible but lighter)
- **State Management:** Zustand (simple and performant)
- **Styling:** Bootstrap 5.3 + custom CSS
- **Build Tool:** Vite (fast development and building)
- **Language:** TypeScript

## Development Setup

### Prerequisites

- [Bun](https://bun.sh/) runtime (recommended) or Node.js 18+
- Go backend server running on port 8080

### Quick Start

1. **Install dependencies:**
   ```bash
   cd cmd/ppa-web/static
   bun install
   ```

2. **Start development server:**
   ```bash
   bun run dev
   ```
   This starts the Vite dev server on http://localhost:3000 with:
   - Hot module replacement (HMR)
   - TypeScript compilation
   - API proxy to Go backend on :8080

3. **Start the Go backend:**
   ```bash
   # From project root
   go run ./cmd/ppa-web
   ```

4. **Access the application:**
   - Development with HMR: http://localhost:3000
   - Go-served static files: http://localhost:8080

### Build for Production

```bash
cd cmd/ppa-web/static
bun run build
```

This creates optimized bundles in `../static-dist/` that can be served by the Go backend.

## Architecture

### Frontend Structure

```
cmd/ppa-web/static/
├── js/
│   ├── components/          # Preact components
│   │   ├── Layout.tsx       # Main layout wrapper
│   │   ├── Dashboard.tsx    # Main dashboard view
│   │   ├── AnalysisView.tsx # Analysis results viewer
│   │   ├── PacketsView.tsx  # Detailed packet viewer
│   │   ├── DocumentsView.tsx# Document browser
│   │   └── ...
│   ├── store.ts            # Zustand state management
│   ├── types.ts            # TypeScript type definitions
│   └── main.tsx            # Application entry point
├── css/
│   └── app.css             # Custom styles
├── package.json            # Dependencies and scripts
├── tsconfig.json           # TypeScript configuration
├── vite.config.ts          # Vite build configuration
└── index.html              # Production HTML template
```

### State Management

Uses Zustand for state management with these main stores:

- **PCAP Files:** Upload status, analysis progress
- **Analysis Results:** Packet data, timelines, visualizations
- **Documents:** Search results, markdown content
- **UI State:** Active views, selected items, loading states

### API Integration

The frontend communicates with the Go backend via REST APIs:

- `POST /api/pcap/upload` - Upload PCAP files
- `POST /api/pcap/{id}/analyze` - Start analysis
- `GET /api/pcap/{id}/status` - Check analysis progress
- `GET /api/analysis/{id}` - Get analysis results
- `GET /api/documents/search` - Search documents

### Component Highlights

#### Dashboard
- File upload with drag & drop
- Statistics overview
- Recent files and analyses

#### Analysis View
- Packet timeline visualization
- Message type distribution charts
- Interactive packet table with filtering

#### Packet View
- Detailed packet information
- Hex dump with search functionality
- Metadata viewer

#### Documents View
- Markdown rendering with syntax highlighting
- Full-text search with result highlighting
- Document browsing and filtering

## Development Guidelines

### Code Style

- Use TypeScript for all components
- Follow React/Preact patterns (hooks, functional components)
- Use Bootstrap classes for styling, custom CSS for specific needs
- Keep components small and focused

### State Management

- Use Zustand selectors for optimal re-renders
- Keep async operations in store actions
- Handle loading and error states consistently

### API Calls

- All API calls go through Zustand store actions
- Use proper error handling and loading states
- Include CORS headers for development

### Testing

```bash
# Run type checking
bun run build

# Format code
bun run format  # (if configured)
```

## Deployment

For production deployment:

1. Build the frontend: `bun run build`
2. The Go server will serve static files from `cmd/ppa-web/static-dist/`
3. All routing is handled client-side (SPA)

## Troubleshooting

### Common Issues

1. **"Module not found" errors:**
   - Ensure all dependencies are installed: `bun install`
   - Check import paths are correct

2. **API calls failing:**
   - Verify Go backend is running on port 8080
   - Check CORS headers in router configuration

3. **Build failures:**
   - Clear node_modules and reinstall: `rm -rf node_modules && bun install`
   - Check TypeScript errors: `bun run build`

4. **Hot reload not working:**
   - Restart the dev server: `bun run dev`
   - Check file permissions and paths

### Development vs Production

- **Development:** Uses Vite dev server with HMR on :3000
- **Production:** Go serves built static files from static-dist/
- **Hybrid:** Go serves dev.html which loads modules for development

## Contributing

When adding new features:

1. Define TypeScript types in `types.ts`
2. Add state management in `store.ts`
3. Create reusable components in `components/`
4. Update API integration as needed
5. Test with both development and production builds

## Browser Support

- Modern browsers with ES2020+ support
- Chrome 88+, Firefox 85+, Safari 14+, Edge 88+
