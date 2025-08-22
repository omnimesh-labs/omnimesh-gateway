'use client';

import React, { useState } from 'react';
import { type LogEntry } from '@/lib/api';

interface LogTableProps {
  logs: LogEntry[];
  loading: boolean;
  onRefresh: () => void;
}

export function LogTable({ logs, loading, onRefresh }: LogTableProps) {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  const toggleRow = (logId: string) => {
    setExpandedRows(prev => {
      const next = new Set(prev);
      if (next.has(logId)) {
        next.delete(logId);
      } else {
        next.add(logId);
      }
      return next;
    });
  };

  const getLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
      case 'error':
        return '#ef4444';
      case 'warn':
        return '#f59e0b';
      case 'info':
        return '#3b82f6';
      case 'debug':
        return '#6b7280';
      default:
        return '#6b7280';
    }
  };

  const getLevelBadge = (level: string) => (
    <span style={{
      background: getLevelColor(level),
      color: 'white',
      padding: '0.25rem 0.5rem',
      borderRadius: '9999px',
      fontSize: '0.75rem',
      fontWeight: '500',
      textTransform: 'uppercase'
    }}>
      {level}
    </span>
  );

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  if (loading && (!logs || logs.length === 0)) {
    return (
      <div style={{
        background: 'white',
        borderRadius: '8px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        padding: '2rem',
        textAlign: 'center'
      }}>
        <div style={{ color: '#6b7280' }}>Loading logs...</div>
      </div>
    );
  }

  if (!logs || logs.length === 0) {
    return (
      <div style={{
        background: 'white',
        borderRadius: '8px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        padding: '2rem',
        textAlign: 'center'
      }}>
        <div style={{ fontSize: '1.125rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
          No logs found
        </div>
        <div style={{ color: '#6b7280' }}>
          Try adjusting your search filters or check back later.
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
                Timestamp
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
                Level
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
                User/Method
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
                Duration
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
                Details
              </th>
            </tr>
          </thead>
          <tbody>
            {logs && logs.map((log, index) => (
              <React.Fragment key={log.id}>
                <tr 
                  style={{ 
                    borderTop: index > 0 ? '1px solid #f3f4f6' : 'none',
                    background: log.error_flag ? '#fef2f2' : 'white'
                  }}
                >
                  <td style={{ padding: '1rem' }}>
                    <div style={{ fontSize: '0.875rem', color: '#374151', fontFamily: 'monospace' }}>
                      {formatDate(log.started_at)}
                    </div>
                  </td>
                  <td style={{ padding: '1rem' }}>
                    {getLevelBadge(log.level)}
                  </td>
                  <td style={{ padding: '1rem' }}>
                    <div style={{ fontSize: '0.875rem', color: '#374151' }}>
                      {log.user_id || 'system'}
                    </div>
                    {log.rpc_method && (
                      <div style={{ fontSize: '0.75rem', color: '#6b7280', fontFamily: 'monospace' }}>
                        {log.rpc_method}
                      </div>
                    )}
                  </td>
                  <td style={{ padding: '1rem' }}>
                    {log.status_code ? (
                      <span style={{
                        background: log.status_code >= 400 ? '#fef2f2' : '#f0fdf4',
                        color: log.status_code >= 400 ? '#dc2626' : '#16a34a',
                        padding: '0.25rem 0.5rem',
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        fontFamily: 'monospace'
                      }}>
                        {log.status_code}
                      </span>
                    ) : (
                      <span style={{ color: '#6b7280', fontSize: '0.875rem' }}>-</span>
                    )}
                  </td>
                  <td style={{ padding: '1rem' }}>
                    {log.duration_ms ? (
                      <span style={{ fontSize: '0.875rem', color: '#374151', fontFamily: 'monospace' }}>
                        {log.duration_ms}ms
                      </span>
                    ) : (
                      <span style={{ color: '#6b7280', fontSize: '0.875rem' }}>-</span>
                    )}
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'right' }}>
                    <button
                      onClick={() => toggleRow(log.id)}
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
                      {expandedRows.has(log.id) ? 'Hide' : 'Show'}
                    </button>
                  </td>
                </tr>
                {expandedRows.has(log.id) && (
                  <tr style={{ background: '#f9fafb' }}>
                    <td colSpan={6} style={{ padding: '1rem' }}>
                      <div style={{ fontSize: '0.875rem' }}>
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
                          <div>
                            <strong style={{ color: '#374151' }}>Server ID:</strong>
                            <div style={{ fontFamily: 'monospace', fontSize: '0.8125rem', color: '#6b7280' }}>
                              {log.server_id || 'N/A'}
                            </div>
                          </div>
                          <div>
                            <strong style={{ color: '#374151' }}>Session ID:</strong>
                            <div style={{ fontFamily: 'monospace', fontSize: '0.8125rem', color: '#6b7280' }}>
                              {log.session_id || 'N/A'}
                            </div>
                          </div>
                          <div>
                            <strong style={{ color: '#374151' }}>Remote IP:</strong>
                            <div style={{ fontFamily: 'monospace', fontSize: '0.8125rem', color: '#6b7280' }}>
                              {log.remote_ip || 'N/A'}
                            </div>
                          </div>
                          <div>
                            <strong style={{ color: '#374151' }}>Storage:</strong>
                            <div style={{ fontSize: '0.8125rem', color: '#6b7280' }}>
                              {log.storage_provider}
                            </div>
                          </div>
                        </div>
                        {log.object_uri && (
                          <div style={{ marginTop: '0.75rem' }}>
                            <strong style={{ color: '#374151' }}>Log Location:</strong>
                            <div style={{ 
                              fontFamily: 'monospace', 
                              fontSize: '0.8125rem', 
                              color: '#6b7280',
                              background: 'white',
                              padding: '0.5rem',
                              borderRadius: '4px',
                              marginTop: '0.25rem',
                              border: '1px solid #e5e7eb'
                            }}>
                              {log.object_uri}
                              {log.byte_offset && ` (offset: ${log.byte_offset})`}
                            </div>
                          </div>
                        )}
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}