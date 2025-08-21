'use client';

import { useState } from 'react';
import { type MCPServer } from '@/lib/api';

interface ServerTableProps {
  servers: MCPServer[];
  onUnregister: (serverId: string) => Promise<void>;
  loading: boolean;
}

export function ServerTable({ servers, onUnregister, loading }: ServerTableProps) {
  const [unregisteringIds, setUnregisteringIds] = useState<Set<string>>(new Set());

  const handleUnregister = async (serverId: string, serverName: string) => {
    if (!confirm(`Are you sure you want to unregister "${serverName}"?`)) {
      return;
    }

    setUnregisteringIds(prev => new Set(prev).add(serverId));
    try {
      await onUnregister(serverId);
    } finally {
      setUnregisteringIds(prev => {
        const next = new Set(prev);
        next.delete(serverId);
        return next;
      });
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return '#10b981';
      case 'inactive':
        return '#6b7280';
      case 'unhealthy':
        return '#ef4444';
      case 'maintenance':
        return '#f59e0b';
      default:
        return '#6b7280';
    }
  };

  const getStatusBadge = (status: string) => (
    <span style={{
      background: getStatusColor(status),
      color: 'white',
      padding: '0.25rem 0.5rem',
      borderRadius: '9999px',
      fontSize: '0.75rem',
      fontWeight: '500',
      textTransform: 'capitalize'
    }}>
      {status}
    </span>
  );

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (loading && servers.length === 0) {
    return (
      <div style={{
        background: 'white',
        borderRadius: '8px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        padding: '2rem',
        textAlign: 'center'
      }}>
        <div style={{ color: '#6b7280' }}>Loading servers...</div>
      </div>
    );
  }

  if (servers.length === 0) {
    return (
      <div style={{
        background: 'white',
        borderRadius: '8px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        padding: '2rem',
        textAlign: 'center'
      }}>
        <div style={{ fontSize: '1.125rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
          No servers registered
        </div>
        <div style={{ color: '#6b7280' }}>
          Get started by registering your first MCP server or browse available servers to add.
        </div>
      </div>
    );
  }

  return (
    <div style={{
      background: 'white',
      borderRadius: '8px',
      boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
      overflow: 'hidden'
    }}>
      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead style={{ background: '#f9fafb' }}>
            <tr>
              <th style={{ 
                textAlign: 'left', 
                padding: '0.75rem 1rem', 
                fontSize: '0.75rem', 
                fontWeight: '500', 
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Name
              </th>
              <th style={{ 
                textAlign: 'left', 
                padding: '0.75rem 1rem', 
                fontSize: '0.75rem', 
                fontWeight: '500', 
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Protocol
              </th>
              <th style={{ 
                textAlign: 'left', 
                padding: '0.75rem 1rem', 
                fontSize: '0.75rem', 
                fontWeight: '500', 
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Status
              </th>
              <th style={{ 
                textAlign: 'left', 
                padding: '0.75rem 1rem', 
                fontSize: '0.75rem', 
                fontWeight: '500', 
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Version
              </th>
              <th style={{ 
                textAlign: 'left', 
                padding: '0.75rem 1rem', 
                fontSize: '0.75rem', 
                fontWeight: '500', 
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Created
              </th>
              <th style={{ 
                textAlign: 'right', 
                padding: '0.75rem 1rem', 
                fontSize: '0.75rem', 
                fontWeight: '500', 
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {servers.map((server, index) => (
              <tr 
                key={server.id}
                style={{ 
                  borderTop: index > 0 ? '1px solid #f3f4f6' : 'none',
                  opacity: unregisteringIds.has(server.id) ? 0.5 : 1
                }}
              >
                <td style={{ padding: '1rem' }}>
                  <div>
                    <div style={{ fontWeight: '500', color: '#111827', marginBottom: '0.25rem' }}>
                      {server.name}
                    </div>
                    {server.description && (
                      <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                        {server.description}
                      </div>
                    )}
                    {server.command && (
                      <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginTop: '0.25rem' }}>
                        {server.command} {server.args?.join(' ')}
                      </div>
                    )}
                  </div>
                </td>
                <td style={{ padding: '1rem' }}>
                  <span style={{
                    background: '#f3f4f6',
                    color: '#374151',
                    padding: '0.25rem 0.5rem',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: '500',
                    textTransform: 'uppercase'
                  }}>
                    {server.protocol}
                  </span>
                </td>
                <td style={{ padding: '1rem' }}>
                  {getStatusBadge(server.status)}
                </td>
                <td style={{ padding: '1rem', color: '#6b7280', fontSize: '0.875rem' }}>
                  {server.version || '-'}
                </td>
                <td style={{ padding: '1rem', color: '#6b7280', fontSize: '0.875rem' }}>
                  {formatDate(server.created_at)}
                </td>
                <td style={{ padding: '1rem', textAlign: 'right' }}>
                  <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                    <button
                      style={{
                        background: '#f3f4f6',
                        color: '#374151',
                        padding: '0.375rem 0.75rem',
                        borderRadius: '4px',
                        border: 'none',
                        fontSize: '0.75rem',
                        cursor: 'pointer',
                        transition: 'background-color 0.2s'
                      }}
                      onMouseOver={(e) => e.currentTarget.style.background = '#e5e7eb'}
                      onMouseOut={(e) => e.currentTarget.style.background = '#f3f4f6'}
                    >
                      View
                    </button>
                    <button
                      onClick={() => handleUnregister(server.id, server.name)}
                      disabled={unregisteringIds.has(server.id)}
                      style={{
                        background: '#fef2f2',
                        color: '#dc2626',
                        padding: '0.375rem 0.75rem',
                        borderRadius: '4px',
                        border: 'none',
                        fontSize: '0.75rem',
                        cursor: unregisteringIds.has(server.id) ? 'not-allowed' : 'pointer',
                        transition: 'background-color 0.2s',
                        opacity: unregisteringIds.has(server.id) ? 0.5 : 1
                      }}
                      onMouseOver={(e) => {
                        if (!unregisteringIds.has(server.id)) {
                          e.currentTarget.style.background = '#fee2e2';
                        }
                      }}
                      onMouseOut={(e) => {
                        if (!unregisteringIds.has(server.id)) {
                          e.currentTarget.style.background = '#fef2f2';
                        }
                      }}
                    >
                      {unregisteringIds.has(server.id) ? 'Removing...' : 'Remove'}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
