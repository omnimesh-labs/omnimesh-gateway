'use client';

export default function HomePage() {
  return (
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
            Gateway Proxy
          </h2>
          <p style={{ color: '#666', marginBottom: '1rem' }}>
            Configure proxy settings and routing
          </p>
          <button style={{ 
            background: '#10b981', 
            color: 'white', 
            padding: '0.5rem 1rem', 
            borderRadius: '6px', 
            border: 'none',
            fontSize: '0.875rem'
          }}>
            Configure Proxy
          </button>
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
            background: '#8b5cf6', 
            color: 'white', 
            padding: '0.5rem 1rem', 
            borderRadius: '6px', 
            border: 'none',
            fontSize: '0.875rem'
          }}>
            View Analytics
          </button>
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
          <li style={{ marginBottom: '0.5rem' }}>Configure gateway proxy settings</li>
          <li style={{ marginBottom: '0.5rem' }}>View real-time analytics and performance metrics</li>
          <li style={{ marginBottom: '0.5rem' }}>Set up authentication and access policies</li>
        </ul>
      </section>
    </div>
  )
}
