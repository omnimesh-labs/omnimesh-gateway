'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { EditEndpointModal } from './EditEndpointModal';
import { EndpointURLs } from './EndpointURLs';

interface Endpoint {
  id: string;
  name: string;
  description?: string;
  namespace_id: string;
  namespace?: {
    id: string;
    name: string;
  };
  enable_api_key_auth: boolean;
  enable_oauth: boolean;
  enable_public_access: boolean;
  use_query_param_auth: boolean;
  rate_limit_requests: number;
  rate_limit_window: number;
  allowed_origins: string[];
  allowed_methods: string[];
  is_active: boolean;
  urls?: {
    sse: string;
    http: string;
    websocket: string;
    openapi: string;
    documentation: string;
  };
  created_at: string;
}

interface EndpointTableProps {
  endpoints: Endpoint[];
  onDelete: (id: string) => void;
  onUpdate: () => void;
}

export function EndpointTable({ endpoints, onDelete, onUpdate }: EndpointTableProps) {
  const [expandedEndpoint, setExpandedEndpoint] = useState<string | null>(null);
  const [editingEndpoint, setEditingEndpoint] = useState<Endpoint | null>(null);
  const [deletingEndpoint, setDeletingEndpoint] = useState<string | null>(null);

  const handleDelete = (endpoint: Endpoint) => {
    if (confirm(`Are you sure you want to delete the endpoint "${endpoint.name}"? This cannot be undone.`)) {
      setDeletingEndpoint(endpoint.id);
      onDelete(endpoint.id);
      setDeletingEndpoint(null);
    }
  };

  const toggleExpanded = (id: string) => {
    setExpandedEndpoint(expandedEndpoint === id ? null : id);
  };

  return (
    <>
      <div style={{
        backgroundColor: 'white',
        border: '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        overflow: 'hidden'
      }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid #e5e7eb', backgroundColor: '#f9fafb' }}>
              <th style={{ padding: '0.75rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Name / URLs
              </th>
              <th style={{ padding: '0.75rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Namespace
              </th>
              <th style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Authentication
              </th>
              <th style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Rate Limit
              </th>
              <th style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Status
              </th>
              <th style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {endpoints.map((endpoint) => (
              <React.Fragment key={endpoint.id}>
                <tr style={{ borderBottom: '1px solid #e5e7eb' }}>
                  <td style={{ padding: '0.75rem' }}>
                    <div>
                      <button
                        onClick={() => toggleExpanded(endpoint.id)}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5rem',
                          background: 'none',
                          border: 'none',
                          padding: 0,
                          cursor: 'pointer',
                          color: '#3b82f6',
                          fontWeight: '500',
                          fontSize: '0.875rem'
                        }}
                      >
                        <svg
                          style={{
                            width: '1rem',
                            height: '1rem',
                            transform: expandedEndpoint === endpoint.id ? 'rotate(90deg)' : 'rotate(0deg)',
                            transition: 'transform 0.2s'
                          }}
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                        {endpoint.name}
                      </button>
                      {endpoint.description && (
                        <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem', marginLeft: '1.5rem' }}>
                          {endpoint.description}
                        </p>
                      )}
                    </div>
                  </td>
                  <td style={{ padding: '0.75rem' }}>
                    {endpoint.namespace ? (
                      <Link
                        href={`/namespaces/${endpoint.namespace.id}`}
                        style={{
                          color: '#3b82f6',
                          textDecoration: 'none',
                          fontSize: '0.875rem'
                        }}
                        onMouseEnter={(e) => e.currentTarget.style.textDecoration = 'underline'}
                        onMouseLeave={(e) => e.currentTarget.style.textDecoration = 'none'}
                      >
                        {endpoint.namespace.name}
                      </Link>
                    ) : (
                      <span style={{ color: '#6b7280', fontSize: '0.875rem' }}>-</span>
                    )}
                  </td>
                  <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                    <div style={{ display: 'flex', justifyContent: 'center', gap: '0.25rem', flexWrap: 'wrap' }}>
                      {endpoint.enable_public_access && (
                        <span style={{
                          display: 'inline-block',
                          padding: '0.125rem 0.5rem',
                          backgroundColor: '#fef3c7',
                          color: '#92400e',
                          borderRadius: '0.25rem',
                          fontSize: '0.75rem',
                          fontWeight: '500'
                        }}>
                          Public
                        </span>
                      )}
                      {endpoint.enable_api_key_auth && (
                        <span style={{
                          display: 'inline-block',
                          padding: '0.125rem 0.5rem',
                          backgroundColor: '#dbeafe',
                          color: '#1e40af',
                          borderRadius: '0.25rem',
                          fontSize: '0.75rem',
                          fontWeight: '500'
                        }}>
                          API Key
                        </span>
                      )}
                      {endpoint.enable_oauth && (
                        <span style={{
                          display: 'inline-block',
                          padding: '0.125rem 0.5rem',
                          backgroundColor: '#e9d5ff',
                          color: '#6b21a8',
                          borderRadius: '0.25rem',
                          fontSize: '0.75rem',
                          fontWeight: '500'
                        }}>
                          OAuth
                        </span>
                      )}
                    </div>
                  </td>
                  <td style={{ padding: '0.75rem', textAlign: 'center', fontSize: '0.875rem', color: '#6b7280' }}>
                    {endpoint.rate_limit_requests}/{endpoint.rate_limit_window}s
                  </td>
                  <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                    <span style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      padding: '0.25rem 0.75rem',
                      backgroundColor: endpoint.is_active ? '#d1fae5' : '#fee2e2',
                      color: endpoint.is_active ? '#065f46' : '#991b1b',
                      borderRadius: '9999px',
                      fontSize: '0.75rem',
                      fontWeight: '500'
                    }}>
                      {endpoint.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                    <div style={{ display: 'flex', justifyContent: 'center', gap: '0.5rem' }}>
                      <button
                        onClick={() => setEditingEndpoint(endpoint)}
                        style={{
                          padding: '0.375rem 0.75rem',
                          backgroundColor: 'white',
                          color: '#374151',
                          border: '1px solid #d1d5db',
                          borderRadius: '0.375rem',
                          fontSize: '0.75rem',
                          fontWeight: '500',
                          cursor: 'pointer',
                          transition: 'all 0.2s'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.backgroundColor = '#f9fafb';
                          e.currentTarget.style.borderColor = '#9ca3af';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.backgroundColor = 'white';
                          e.currentTarget.style.borderColor = '#d1d5db';
                        }}
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(endpoint)}
                        disabled={deletingEndpoint === endpoint.id}
                        style={{
                          padding: '0.375rem 0.75rem',
                          backgroundColor: 'white',
                          color: '#dc2626',
                          border: '1px solid #fca5a5',
                          borderRadius: '0.375rem',
                          fontSize: '0.75rem',
                          fontWeight: '500',
                          cursor: deletingEndpoint === endpoint.id ? 'not-allowed' : 'pointer',
                          opacity: deletingEndpoint === endpoint.id ? 0.5 : 1,
                          transition: 'all 0.2s'
                        }}
                        onMouseEnter={(e) => {
                          if (deletingEndpoint !== endpoint.id) {
                            e.currentTarget.style.backgroundColor = '#fef2f2';
                            e.currentTarget.style.borderColor = '#f87171';
                          }
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.backgroundColor = 'white';
                          e.currentTarget.style.borderColor = '#fca5a5';
                        }}
                      >
                        {deletingEndpoint === endpoint.id ? 'Deleting...' : 'Delete'}
                      </button>
                    </div>
                  </td>
                </tr>
                {expandedEndpoint === endpoint.id && endpoint.urls && (
                  <tr style={{ backgroundColor: '#f9fafb' }}>
                    <td colSpan={6} style={{ padding: '1rem' }}>
                      <EndpointURLs
                        urls={endpoint.urls}
                        endpointName={endpoint.name}
                        useQueryParamAuth={endpoint.use_query_param_auth}
                      />
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>

      {/* Edit Endpoint Modal */}
      {editingEndpoint && (
        <EditEndpointModal
          endpoint={editingEndpoint}
          onClose={() => setEditingEndpoint(null)}
          onUpdate={onUpdate}
        />
      )}
    </>
  );
}
