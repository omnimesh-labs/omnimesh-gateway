# MCP Gateway Frontend

Admin dashboard for the MCP Gateway project - a Next.js 14 web interface for managing Model Context Protocol servers.

## Overview

This frontend provides a comprehensive admin interface for the MCP Gateway backend, featuring:

- **Server Management**: Register, monitor, and configure MCP servers
- **MCP Discovery**: Browse and register community MCP servers  
- **Real-time Monitoring**: Live health checks and server status
- **Clean Interface**: Modern, responsive design with intuitive navigation

## Getting Started

### Prerequisites
- Node.js 18+ (Bun recommended)
- MCP Gateway backend running on `localhost:8080`

### Installation

```bash
bun install
```

### Development

```bash
bun dev
```

Open [http://localhost:3000](http://localhost:3000) to view the dashboard.

## Tech Stack

- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript  
- **Styling**: CSS-in-JS (inline styles)
- **Package Manager**: Bun
- **State Management**: React hooks
- **HTTP Client**: Fetch API with comprehensive error handling

## Project Structure

```
src/
├── app/                          # Next.js App Router
│   ├── layout.tsx               # Root layout with navigation
│   ├── page.tsx                 # Dashboard homepage
│   ├── globals.css              # Global styles
│   └── servers/                 # Server management
│       └── page.tsx            
├── components/                   # React components
│   ├── Navigation.tsx           # Main navigation
│   ├── HealthCheck.tsx          # Backend health monitoring
│   ├── Toast.tsx               # Notification system
│   └── servers/                # Server management components
│       ├── ServerTable.tsx     # Server data table
│       ├── AvailableServersList.tsx  # MCP discovery
│       └── RegisterServerModal.tsx   # Server registration
└── lib/
    └── api.ts                   # API service layer
```

## Available Scripts

- `bun dev` - Start development server (localhost:3000)
- `bun build` - Build for production
- `bun start` - Start production server  
- `bun lint` - Run ESLint
- `bun type-check` - TypeScript type checking

## Features

### Server Management
- **Dashboard Overview**: Server statistics and quick navigation
- **Server Registry**: View all registered MCP servers with status
- **Server Registration**: Add new servers manually or from community
- **Server Monitoring**: Real-time health checks and status updates

### MCP Discovery  
- **Community Packages**: Browse available MCP servers
- **Package Search**: Find servers by name or description
- **One-Click Install**: Register community servers instantly
- **Package Details**: GitHub stars, downloads, and metadata

### User Experience
- **Responsive Design**: Works on desktop, tablet, and mobile
- **Real-time Updates**: Live status indicators and notifications
- **Toast Notifications**: Success/error feedback for all actions
- **Loading States**: Comprehensive loading indicators throughout

## API Integration

The frontend communicates with the MCP Gateway backend via REST API:

```typescript
// Server Management
serverApi.listServers()           // Get registered servers
serverApi.registerServer(data)    // Register new server
serverApi.unregisterServer(id)    // Remove server

// MCP Discovery  
discoveryApi.searchPackages(query)  // Search community packages
discoveryApi.listPackages()         // Browse all packages

// Health Monitoring
// GET /health - Backend health check
```

## Configuration

### Environment Variables
```bash
# .env.local
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api
```

### Backend Requirements
- MCP Gateway backend running on port 8080
- CORS configured for frontend domain
- All API endpoints available and responding

## Development

### Code Style
- **TypeScript**: Strict typing for all components and API calls
- **Inline CSS**: Component-scoped styling for simplicity
- **React Hooks**: useState/useEffect for state management
- **Error Handling**: Comprehensive error states and user feedback

### Adding New Features
1. Create components in appropriate subdirectories
2. Add API calls to `lib/api.ts`
3. Implement proper loading and error states
4. Add TypeScript types for all data structures
5. Follow existing patterns for styling and state management

For detailed development information, see [CLAUDE.md](./CLAUDE.md).
