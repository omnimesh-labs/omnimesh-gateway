# MCP Gateway Frontend - Claude Development Guide

## Project Overview

The **MCP Gateway Frontend** is a Next.js 14 admin dashboard for managing the MCP Gateway backend. It provides a clean, responsive interface for server management, monitoring, and configuration.

### Core Purpose
- **Admin Interface**: Web-based dashboard for MCP Gateway administration
- **Server Management**: Register, monitor, and configure MCP servers
- **Real-time Monitoring**: Health checks and server status visualization
- **User Experience**: Clean, modern interface with inline CSS styling

### Key Features
- 🖥️ **Server Management Dashboard** - Full CRUD operations for MCP servers
- 🔍 **MCP Server Discovery** - Browse and register community MCP servers
- ⚡ **Real-time Health Monitoring** - Live server status and health checks
- 📊 **Clean Admin Interface** - Intuitive tabs, modals, and data tables
- 🌐 **API Integration** - Comprehensive REST API client for backend communication

## Architecture

### Technology Stack
- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript
- **Styling**: CSS-in-JS (inline styles) for simplicity and performance
- **Package Manager**: Bun (faster alternative to npm)
- **State Management**: React hooks (useState, useEffect)
- **HTTP Client**: Native Fetch API with comprehensive error handling

### Project Structure
```
apps/frontend/
├── src/
│   ├── app/                  # Next.js App Router pages
│   │   ├── layout.tsx        # Root layout with navigation
│   │   ├── page.tsx          # Dashboard homepage
│   │   ├── globals.css       # Global CSS styles
│   │   └── servers/          # Server management section
│   │       └── page.tsx      # Server management page
│   ├── components/           # Reusable React components
│   │   ├── Navigation.tsx    # Main navigation component
│   │   ├── HealthCheck.tsx   # Backend health monitoring
│   │   ├── Toast.tsx         # Toast notification system
│   │   └── servers/          # Server-specific components
│   │       ├── ServerTable.tsx         # Server data table
│   │       ├── AvailableServersList.tsx # MCP discovery component
│   │       └── RegisterServerModal.tsx  # Server registration modal
│   └── lib/                  # Utility libraries
│       └── api.ts            # API service layer
├── package.json              # Dependencies and scripts
├── next.config.js           # Next.js configuration
├── tsconfig.json            # TypeScript configuration
└── README.md                # Basic project documentation
```

## Component Architecture

### Core Components

#### Navigation (`src/components/Navigation.tsx`)
- **Purpose**: Main application navigation with active state tracking
- **Features**: Responsive navigation bar with highlighted current page
- **Styling**: Inline CSS with hover effects and transitions

#### Server Management Components

##### ServerTable (`src/components/servers/ServerTable.tsx`)
- **Purpose**: Display registered MCP servers in a clean data table
- **Features**:
  - Server status badges with color coding (active/inactive/unhealthy/maintenance)
  - Confirmation dialogs for destructive actions
  - Loading states and optimistic updates
  - Responsive table design with overflow handling
  - Action buttons (View, Remove) with proper loading states

##### AvailableServersList (`src/components/servers/AvailableServersList.tsx`)
- **Purpose**: Browse and register community MCP servers
- **Features**:
  - Search and pagination for MCP packages
  - Server details with GitHub integration
  - One-click server registration
  - Package metadata display (stars, downloads)

##### RegisterServerModal (`src/components/servers/RegisterServerModal.tsx`)
- **Purpose**: Modal dialog for manual server registration
- **Features**:
  - Form validation and error handling
  - Support for different protocols (STDIO, HTTP, WebSocket)
  - Environment variable and argument configuration
  - Modal overlay with escape key handling

#### Utility Components

##### HealthCheck (`src/components/HealthCheck.tsx`)
- **Purpose**: Monitor backend connectivity and health status
- **Features**: Real-time health status indicator with retry logic

##### Toast (`src/components/Toast.tsx`)
- **Purpose**: Non-intrusive notification system
- **Features**: Success/error toast messages with auto-dismiss

## API Integration

### API Service Layer (`src/lib/api.ts`)

Comprehensive REST API client with the following modules:

#### Server Management API
```typescript
serverApi.listServers()           // Get all registered servers
serverApi.getServer(id)           // Get specific server details  
serverApi.registerServer(data)    // Register new server
serverApi.updateServer(id, data)  // Update existing server
serverApi.unregisterServer(id)    // Remove server registration
serverApi.getServerStats(id)      // Get server statistics
```

#### MCP Discovery API
```typescript
discoveryApi.searchPackages(query)    // Search community packages
discoveryApi.listPackages()           // List all available packages
discoveryApi.getPackageDetails(name)  // Get package details
```

#### Session Management API
```typescript
sessionApi.createSession(serverId)  // Create MCP session
sessionApi.listSessions()           // List active sessions
sessionApi.closeSession(sessionId)  // Close session
```

### API Configuration
- **Base URL**: `http://localhost:8080/api`
- **CORS**: Configured for cross-origin requests
- **Error Handling**: Comprehensive error mapping and user-friendly messages
- **Request Format**: JSON with proper Content-Type headers

## Page Structure

### Dashboard (`src/app/page.tsx`)
- **Purpose**: Main landing page with navigation cards
- **Features**:
  - Server Management card with direct navigation
  - Analytics section (placeholder)
  - Quick start guide for new users

### Server Management (`src/app/servers/page.tsx`)
- **Purpose**: Complete server management interface
- **Features**:
  - Tabbed interface (Registered Servers / Available Servers)
  - Real-time server listing with refresh functionality
  - Integrated health monitoring
  - Toast notifications for user feedback
  - Modal-based server registration

## Styling Approach

### CSS-in-JS Strategy
- **Inline Styles**: All styling done via React style props for simplicity
- **Design System**: Consistent color palette and spacing scale
- **Responsive Design**: Grid layouts and flexible components
- **Interactive States**: Hover effects and transitions for better UX

### Color Palette
```css
/* Status Colors */
--success: #10b981    /* Active/healthy states */
--warning: #f59e0b    /* Maintenance/warning */
--danger: #ef4444     /* Error/unhealthy states */
--info: #3b82f6       /* Primary actions/links */
--neutral: #6b7280    /* Inactive/secondary text */

/* UI Colors */
--background: #ffffff
--border: #e5e7eb
--surface: #f9fafb
--text-primary: #111827
--text-secondary: #6b7280
```

## Development Workflow

### Build Commands
```bash
# Development
bun install          # Install dependencies
bun dev             # Start development server (localhost:3000)

# Production
bun build           # Build for production
bun start           # Start production server

# Code Quality
bun lint            # Run ESLint
bun type-check      # TypeScript type checking
```

### Development Guidelines
1. **Component Structure**: Keep components small and focused
2. **State Management**: Use React hooks for local state, lift state up when needed
3. **Error Handling**: Always handle loading and error states
4. **TypeScript**: Leverage strict typing for API responses and component props
5. **Accessibility**: Use semantic HTML and proper ARIA attributes

## Backend Integration

### API Endpoints Used
- `GET /api/gateway/servers` - List registered servers
- `POST /api/gateway/servers` - Register new server
- `DELETE /api/gateway/servers/{id}` - Remove server
- `GET /api/mcp/search` - Search community packages
- `GET /api/mcp/packages` - List available packages
- `GET /health` - Backend health check

### Error Handling
- **Network Errors**: Graceful degradation with retry options
- **CORS Issues**: Clear error messages for configuration problems
- **API Errors**: User-friendly error display with technical details
- **Loading States**: Comprehensive loading indicators throughout the UI

## Feature Roadmap

### ✅ Implemented
- Server management dashboard with CRUD operations
- MCP server discovery and registration
- Real-time health monitoring
- Toast notification system
- Responsive table design with proper loading states
- Modal-based server registration

### 🔄 In Progress / Planned
- Authentication and user management UI
- Advanced server configuration (rate limiting, policies)
- Real-time analytics dashboard
- Virtual server management interface
- Logging and audit trail viewer
- WebSocket integration for real-time updates

## Deployment

### Production Considerations
- **Build Optimization**: Next.js automatic code splitting and optimization
- **Environment Variables**: Configure API base URL for different environments
- **Static Assets**: Optimized image loading and caching
- **SEO**: Proper meta tags and structured data

### Environment Configuration
```bash
# .env.local
NEXT_PUBLIC_API_BASE_URL=https://your-backend-domain.com/api
```

## Testing Strategy

### Component Testing (Recommended)
- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test API integration and user workflows
- **E2E Tests**: Full user journey testing with Playwright/Cypress

### Manual Testing Checklist
- [ ] Server registration and removal
- [ ] Health check functionality
- [ ] Toast notifications
- [ ] Responsive design on different screen sizes
- [ ] Error state handling
- [ ] Loading state indicators

## Security Considerations

### Frontend Security
- **XSS Prevention**: Proper sanitization of user input
- **CSRF Protection**: Use proper headers and validation
- **Content Security Policy**: Implement CSP headers in production
- **Secure API Communication**: HTTPS only in production

### Data Handling
- **Sensitive Data**: Never store secrets or API keys in frontend code
- **User Input**: Validate all form inputs client-side and server-side
- **Error Messages**: Don't expose sensitive system information in error messages

The frontend provides a clean, efficient admin interface for the MCP Gateway with room for future enhancements and integrations.
