'use client';

import { useState } from 'react';

interface EndpointURLsProps {
  urls: {
    sse: string;
    http: string;
    websocket: string;
    openapi: string;
    documentation: string;
  };
  endpointName: string;
  useQueryParamAuth?: boolean;
}

export function EndpointURLs({ urls, endpointName, useQueryParamAuth }: EndpointURLsProps) {
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null);

  const copyToClipboard = (url: string, label: string) => {
    navigator.clipboard.writeText(url);
    setCopiedUrl(label);
    setTimeout(() => setCopiedUrl(null), 2000);
  };

  const urlItems = [
    {
      label: 'SSE (Server-Sent Events)',
      url: urls.sse,
      description: 'Real-time event streaming endpoint',
      icon: '📡'
    },
    {
      label: 'HTTP/MCP',
      url: urls.http,
      description: 'Streamable HTTP endpoint for MCP protocol',
      icon: '🔗'
    },
    {
      label: 'WebSocket',
      url: urls.websocket,
      description: 'Bidirectional WebSocket connection',
      icon: '🔌'
    },
    {
      label: 'OpenAPI Spec',
      url: urls.openapi,
      description: 'OpenAPI 3.0 specification document',
      icon: '📄'
    },
    {
      label: 'API Documentation',
      url: urls.documentation,
      description: 'Interactive Swagger UI documentation',
      icon: '📚'
    }
  ];

  return (
    <div>
      <h4 style={{
        fontSize: '0.875rem',
        fontWeight: '600',
        color: '#111827',
        marginBottom: '0.75rem'
      }}>
        Endpoint URLs
      </h4>

      <div style={{ display: 'grid', gap: '0.75rem' }}>
        {urlItems.map((item) => (
          <div
            key={item.label}
            style={{
              padding: '0.75rem',
              backgroundColor: 'white',
              border: '1px solid #e5e7eb',
              borderRadius: '0.375rem'
            }}
          >
            <div style={{ display: 'flex', alignItems: 'start', justifyContent: 'space-between' }}>
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
                  <span style={{ fontSize: '1rem' }}>{item.icon}</span>
                  <span style={{ fontWeight: '500', fontSize: '0.875rem', color: '#111827' }}>
                    {item.label}
                  </span>
                </div>
                <p style={{ fontSize: '0.75rem', color: '#6b7280', marginBottom: '0.5rem' }}>
                  {item.description}
                </p>
                <div style={{
                  padding: '0.5rem',
                  backgroundColor: '#f9fafb',
                  borderRadius: '0.25rem',
                  fontFamily: 'monospace',
                  fontSize: '0.75rem',
                  color: '#374151',
                  wordBreak: 'break-all'
                }}>
                  {item.url}
                </div>
              </div>
              <button
                onClick={() => copyToClipboard(item.url, item.label)}
                style={{
                  marginLeft: '0.75rem',
                  padding: '0.375rem',
                  backgroundColor: copiedUrl === item.label ? '#10b981' : 'white',
                  color: copiedUrl === item.label ? 'white' : '#6b7280',
                  border: '1px solid',
                  borderColor: copiedUrl === item.label ? '#10b981' : '#d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.75rem',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                  whiteSpace: 'nowrap'
                }}
                onMouseEnter={(e) => {
                  if (copiedUrl !== item.label) {
                    e.currentTarget.style.backgroundColor = '#f9fafb';
                    e.currentTarget.style.borderColor = '#9ca3af';
                  }
                }}
                onMouseLeave={(e) => {
                  if (copiedUrl !== item.label) {
                    e.currentTarget.style.backgroundColor = 'white';
                    e.currentTarget.style.borderColor = '#d1d5db';
                  }
                }}
              >
                {copiedUrl === item.label ? '✓ Copied' : 'Copy'}
              </button>
            </div>
          </div>
        ))}
      </div>

      {useQueryParamAuth && (
        <div style={{
          marginTop: '1rem',
          padding: '0.75rem',
          backgroundColor: '#fef3c7',
          border: '1px solid #fde68a',
          borderRadius: '0.375rem'
        }}>
          <p style={{ fontSize: '0.75rem', color: '#92400e' }}>
            <strong>Query Parameter Authentication Enabled:</strong> You can append <code style={{
              padding: '0.125rem 0.25rem',
              backgroundColor: 'white',
              borderRadius: '0.125rem',
              fontFamily: 'monospace'
            }}>?api_key=YOUR_API_KEY</code> to any URL for authentication.
          </p>
        </div>
      )}
    </div>
  );
}
