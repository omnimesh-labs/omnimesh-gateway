'use client';

import { ProtectedRoute } from '@/components/ProtectedRoute';

export default function HomePage() {
  return (
    <ProtectedRoute>
      <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
      <header style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <h1 style={{ fontSize: '2.5rem', fontWeight: 'bold', color: '#333', marginBottom: '1rem' }}>
          Dashboard
        </h1>
        <p style={{ fontSize: '1.25rem', color: '#666' }}>
          Model Context Protocol Gateway
        </p>
      </header>

      <div style={{ 
        display: 'grid', 
        gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', 
        gap: '1.5rem',
        marginBottom: '3rem'
      }}>
        <div style={{ 
          background: 'white', 
          borderRadius: '8px', 
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)', 
          padding: '1.5rem' 
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333', marginBottom: '0.75rem' }}>
            Server Management
          </h2>
          <p style={{ color: '#666', marginBottom: '1rem' }}>
            Manage and monitor your MCP servers
          </p>
          <a
            href="/servers"
            style={{ 
              background: '#3b82f6', 
              color: 'white', 
              padding: '0.5rem 1rem', 
              borderRadius: '6px', 
              border: 'none',
              fontSize: '0.875rem',
              textDecoration: 'none',
              display: 'inline-block',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => e.currentTarget.style.background = '#2563eb'}
            onMouseOut={(e) => e.currentTarget.style.background = '#3b82f6'}
          >
            View Servers
          </a>
        </div>

        <div style={{ 
          background: 'white', 
          borderRadius: '8px', 
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)', 
          padding: '1.5rem' 
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333', marginBottom: '0.75rem' }}>
            Content Management
          </h2>
          <p style={{ color: '#666', marginBottom: '1rem' }}>
            Manage tools, prompts, and resources
          </p>
          <a
            href="/content"
            style={{ 
              background: '#06b6d4', 
              color: 'white', 
              padding: '0.5rem 1rem', 
              borderRadius: '6px', 
              border: 'none',
              fontSize: '0.875rem',
              textDecoration: 'none',
              display: 'inline-block',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => e.currentTarget.style.background = '#0891b2'}
            onMouseOut={(e) => e.currentTarget.style.background = '#06b6d4'}
          >
            Manage Content
          </a>
        </div>

        <div style={{ 
          background: 'white', 
          borderRadius: '8px', 
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)', 
          padding: '1.5rem' 
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333', marginBottom: '0.75rem' }}>
            Configuration
          </h2>
          <p style={{ color: '#666', marginBottom: '1rem' }}>
            Export and import your MCP Gateway configuration
          </p>
          <a
            href="/configuration"
            style={{ 
              background: '#dc2626', 
              color: 'white', 
              padding: '0.5rem 1rem', 
              borderRadius: '6px', 
              border: 'none',
              fontSize: '0.875rem',
              textDecoration: 'none',
              display: 'inline-block',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => e.currentTarget.style.background = '#b91c1c'}
            onMouseOut={(e) => e.currentTarget.style.background = '#dc2626'}
          >
            Manage Configuration
          </a>
        </div>

        <div style={{ 
          background: 'white', 
          borderRadius: '8px', 
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)', 
          padding: '1.5rem' 
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333', marginBottom: '0.75rem' }}>
            Logging & Audit
          </h2>
          <p style={{ color: '#666', marginBottom: '1rem' }}>
            Monitor system logs and audit trail
          </p>
          <a
            href="/logs"
            style={{ 
              background: '#8b5cf6', 
              color: 'white', 
              padding: '0.5rem 1rem', 
              borderRadius: '6px', 
              border: 'none',
              fontSize: '0.875rem',
              textDecoration: 'none',
              display: 'inline-block',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => e.currentTarget.style.background = '#7c3aed'}
            onMouseOut={(e) => e.currentTarget.style.background = '#8b5cf6'}
          >
            View Logs
          </a>
        </div>

        <div style={{ 
          background: 'white', 
          borderRadius: '8px', 
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)', 
          padding: '1.5rem' 
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333', marginBottom: '0.75rem' }}>
            Analytics
          </h2>
          <p style={{ color: '#666', marginBottom: '1rem' }}>
            View usage statistics and performance metrics
          </p>
          <button style={{ 
            background: '#f59e0b', 
            color: 'white', 
            padding: '0.5rem 1rem', 
            borderRadius: '6px', 
            border: 'none',
            fontSize: '0.875rem'
          }}>
            View Analytics
          </button>
        </div>

        <div style={{ 
          background: 'white', 
          borderRadius: '8px', 
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)', 
          padding: '1.5rem' 
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333', marginBottom: '0.75rem' }}>
            Policy Management
          </h2>
          <p style={{ color: '#666', marginBottom: '1rem' }}>
            Configure authentication, rate limiting, and access policies
          </p>
          <a
            href="/policies"
            style={{ 
              background: '#10b981', 
              color: 'white', 
              padding: '0.5rem 1rem', 
              borderRadius: '6px', 
              border: 'none',
              fontSize: '0.875rem',
              textDecoration: 'none',
              display: 'inline-block',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => e.currentTarget.style.background = '#059669'}
            onMouseOut={(e) => e.currentTarget.style.background = '#10b981'}
          >
            Manage Policies
          </a>
        </div>
      </div>

      <section style={{ 
        background: 'white', 
        borderRadius: '8px', 
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)', 
        padding: '2rem' 
      }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: '600', color: '#333', marginBottom: '1rem' }}>
          Quick Start
        </h2>
        <p style={{ color: '#666', marginBottom: '1rem' }}>
          Welcome to the MCP Gateway dashboard. This interface allows you to:
        </p>
        <ul style={{ color: '#666', paddingLeft: '1.5rem' }}>
          <li style={{ marginBottom: '0.5rem' }}>Manage and monitor MCP server instances</li>
          <li style={{ marginBottom: '0.5rem' }}>Create and manage tools, prompts, and resources</li>
          <li style={{ marginBottom: '0.5rem' }}>Export and import gateway configurations across environments</li>
          <li style={{ marginBottom: '0.5rem' }}>Configure authentication, rate limiting, and access policies</li>
          <li style={{ marginBottom: '0.5rem' }}>Monitor system logs and audit administrative actions</li>
          <li style={{ marginBottom: '0.5rem' }}>View real-time analytics and performance metrics</li>
        </ul>
      </section>
      </div>
    </ProtectedRoute>
  )
}
