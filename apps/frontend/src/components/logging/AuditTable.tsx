'use client';

import React, { useState } from 'react';
import { type AuditLogEntry } from '@/lib/api';

interface AuditTableProps {
  auditLogs: AuditLogEntry[];
  loading: boolean;
  onRefresh: () => void;
}

export function AuditTable({ auditLogs, loading, onRefresh }: AuditTableProps) {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  const toggleRow = (auditId: string) => {
    setExpandedRows(prev => {
      const next = new Set(prev);
      if (next.has(auditId)) {
        next.delete(auditId);
      } else {
        next.add(auditId);
      }
      return next;
    });
  };

  const getActionColor = (action: string) => {
    switch (action.toLowerCase()) {
      case 'create':
        return '#10b981';
      case 'update':
        return '#3b82f6';
      case 'delete':
      case 'unregister':
        return '#ef4444';
      case 'register':
        return '#8b5cf6';
      default:
        return '#6b7280';
    }
  };

  const getActionBadge = (action: string) => (
    <span style={{
      background: getActionColor(action),
      color: 'white',
      padding: '0.25rem 0.5rem',
      borderRadius: '9999px',
      fontSize: '0.75rem',
      fontWeight: '500',
      textTransform: 'capitalize'
    }}>
      {action}
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

  const formatJsonValue = (obj: Record<string, any> | undefined) => {
    if (!obj) return null;
    return (
      <pre style={{
        background: '#f9fafb',
        border: '1px solid #e5e7eb',
        borderRadius: '4px',
        padding: '0.5rem',
        fontSize: '0.8125rem',
        fontFamily: 'monospace',
        overflow: 'auto',
        maxHeight: '200px',
        margin: '0.25rem 0'
      }}>
        {JSON.stringify(obj, null, 2)}
      </pre>
    );
  };

  if (loading && (!auditLogs || auditLogs.length === 0)) {
    return (
      <div style={{
        background: 'white',
        borderRadius: '8px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        padding: '2rem',
        textAlign: 'center'
      }}>
        <div style={{ color: '#6b7280' }}>Loading audit logs...</div>
      </div>
    );
  }

  if (!auditLogs || auditLogs.length === 0) {
    return (
      <div style={{
        background: 'white',
        borderRadius: '8px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        padding: '2rem',
        textAlign: 'center'
      }}>
        <div style={{ fontSize: '1.125rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
          No audit logs found
        </div>
        <div style={{ color: '#6b7280' }}>
          Audit logs will appear here as actions are performed in the system.
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
                Action
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
                Resource
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
                Actor
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
                IP Address
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
            {auditLogs && auditLogs.map((audit, index) => (
              <React.Fragment key={audit.id}>
                <tr
                  style={{
                    borderTop: index > 0 ? '1px solid #f3f4f6' : 'none'
                  }}
                >
                  <td style={{ padding: '1rem' }}>
                    <div style={{ fontSize: '0.875rem', color: '#374151', fontFamily: 'monospace' }}>
                      {formatDate(audit.created_at)}
                    </div>
                  </td>
                  <td style={{ padding: '1rem' }}>
                    {getActionBadge(audit.action)}
                  </td>
                  <td style={{ padding: '1rem' }}>
                    <div style={{ fontSize: '0.875rem', color: '#374151' }}>
                      {audit.resource_type}
                    </div>
                    {audit.resource_id && (
                      <div style={{ fontSize: '0.75rem', color: '#6b7280', fontFamily: 'monospace', marginTop: '0.25rem' }}>
                        {audit.resource_id.substring(0, 8)}...
                      </div>
                    )}
                  </td>
                  <td style={{ padding: '1rem' }}>
                    <div style={{ fontSize: '0.875rem', color: '#374151' }}>
                      {audit.actor_id}
                    </div>
                  </td>
                  <td style={{ padding: '1rem' }}>
                    <div style={{ fontSize: '0.875rem', color: '#6b7280', fontFamily: 'monospace' }}>
                      {audit.actor_ip || 'N/A'}
                    </div>
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'right' }}>
                    <button
                      onClick={() => toggleRow(audit.id)}
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
                      {expandedRows.has(audit.id) ? 'Hide' : 'Show'}
                    </button>
                  </td>
                </tr>
                {expandedRows.has(audit.id) && (
                  <tr style={{ background: '#f9fafb' }}>
                    <td colSpan={6} style={{ padding: '1rem' }}>
                      <div style={{ fontSize: '0.875rem' }}>
                        <div style={{ marginBottom: '1rem' }}>
                          <strong style={{ color: '#374151' }}>Resource ID:</strong>
                          <div style={{ fontFamily: 'monospace', fontSize: '0.8125rem', color: '#6b7280', marginTop: '0.25rem' }}>
                            {audit.resource_id || 'N/A'}
                          </div>
                        </div>

                        {audit.old_values && Object.keys(audit.old_values).length > 0 && (
                          <div style={{ marginBottom: '1rem' }}>
                            <strong style={{ color: '#374151' }}>Previous Values:</strong>
                            {formatJsonValue(audit.old_values)}
                          </div>
                        )}

                        {audit.new_values && Object.keys(audit.new_values).length > 0 && (
                          <div style={{ marginBottom: '1rem' }}>
                            <strong style={{ color: '#374151' }}>New Values:</strong>
                            {formatJsonValue(audit.new_values)}
                          </div>
                        )}

                        {audit.metadata && Object.keys(audit.metadata).length > 0 && (
                          <div>
                            <strong style={{ color: '#374151' }}>Metadata:</strong>
                            {formatJsonValue(audit.metadata)}
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
