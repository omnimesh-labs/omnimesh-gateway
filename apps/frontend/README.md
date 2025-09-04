# MCP Gateway Frontend

## Overview

The MCP Gateway Frontend is a Next.js dashboard for managing the MCP Gateway - a production-ready API gateway for Model Context Protocol (MCP) servers.

## Features

- **Dashboard**: Administrative interface for managing MCP servers, namespaces, and configurations
- **Authentication**: JWT-based authentication with role-based access control
- **Server Management**: Register, monitor, and manage MCP servers
- **Namespace Management**: Multi-tenant namespace isolation and management
- **Virtual Servers**: Create and manage virtual MCP servers from REST/GraphQL/gRPC services
- **Monitoring**: Real-time health monitoring and logging
- **Rate Limiting**: Configure and monitor rate limiting policies

## Technology Stack

- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript
- **UI Library**: Material-UI (MUI) v7
- **Styling**: Tailwind CSS
- **State Management**: React Query for server state
- **Authentication**: NextAuth.js with JWT
- **Package Manager**: Bun

## Development

```bash
# Install dependencies
bun install

# Start development server
bun run dev

# Build for production
bun run build

# Start production server
bun run start

# Lint code
bun run lint

# Fix linting issues
bun run lint:fix
```

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
├── components/            # Reusable React components
├── lib/                   # Utilities and API clients
├── @auth/                # Authentication components and logic
├── @fuse/               # UI framework components
├── @i18n/              # Internationalization
├── configs/             # Application configuration
├── contexts/            # React contexts
├── hooks/               # Custom React hooks
├── styles/              # Global styles and themes
└── utils/               # Utility functions
```

## License

This project is part of the MCP Gateway system.
