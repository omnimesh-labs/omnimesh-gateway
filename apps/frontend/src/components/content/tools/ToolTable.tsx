'use client';

import { useState } from 'react';
import { Tool } from '@/lib/api';

interface ToolTableProps {
  tools: Tool[];
  onEdit: (tool: Tool) => void;
  onDelete: (tool: Tool) => void;
  onView: (tool: Tool) => void;
  loading?: boolean;
}

export function ToolTable({ tools, onEdit, onDelete, onView, loading = false }: ToolTableProps) {
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [showConfirmDelete, setShowConfirmDelete] = useState<Tool | null>(null);

  const handleDeleteConfirm = async (tool: Tool) => {
    setDeletingId(tool.id);
    try {
      await onDelete(tool);
      setShowConfirmDelete(null);
    } catch (error) {
      // Error handling is done in parent component
    } finally {
      setDeletingId(null);
    }
  };

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      general: '#6b7280',     // gray
      data: '#3b82f6',        // blue
      file: '#10b981',        // green
      web: '#8b5cf6',         // purple
      system: '#f59e0b',      // yellow
      ai: '#ef4444',          // red
      dev: '#06b6d4',         // cyan
      custom: '#6b7280'       // gray
    };
    return colors[category] || colors.custom;
  };

  const getImplementationTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      internal: '#10b981',    // green
      external: '#3b82f6',    // blue
      webhook: '#8b5cf6',     // purple
      script: '#f59e0b'       // yellow
    };
    return colors[type] || colors.internal;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const truncateText = (text: string, maxLength: number) => {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };

  if (loading && (!tools || tools.length === 0)) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <div style={{ color: '#666' }}>Loading tools...</div>
      </div>
    );
  }

  if (!tools || tools.length === 0) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <div style={{ color: '#666', marginBottom: '1rem' }}>No tools found</div>
        <p style={{ color: '#999', fontSize: '0.875rem' }}>
          Create your first tool to get started
        </p>
      </div>
    );
  }

  return (
    <>
      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '2px solid #e5e7eb' }}>
              <th style={{
                padding: '0.75rem',
                textAlign: 'left',
                fontWeight: '600',
                color: '#374151',
                fontSize: '0.875rem',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Name & Category
              </th>
              <th style={{
                padding: '0.75rem',
                textAlign: 'left',
                fontWeight: '600',
                color: '#374151',
                fontSize: '0.875rem',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Function & Type
              </th>
              <th style={{
                padding: '0.75rem',
                textAlign: 'left',
                fontWeight: '600',
                color: '#374151',
                fontSize: '0.875rem',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Usage & Status
              </th>
              <th style={{
                padding: '0.75rem',
                textAlign: 'left',
                fontWeight: '600',
                color: '#374151',
                fontSize: '0.875rem',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Updated
              </th>
              <th style={{
                padding: '0.75rem',
                textAlign: 'right',
                fontWeight: '600',
                color: '#374151',
                fontSize: '0.875rem',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {tools.map((tool) => (
              <tr
                key={tool.id}
                style={{
                  borderBottom: '1px solid #f3f4f6',
                  backgroundColor: 'white',
                  transition: 'background-color 0.2s'
                }}
                onMouseOver={(e) => e.currentTarget.style.backgroundColor = '#f9fafb'}
                onMouseOut={(e) => e.currentTarget.style.backgroundColor = 'white'}
              >
                <td style={{ padding: '1rem 0.75rem' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <div
                      style={{
                        width: '8px',
                        height: '8px',
                        borderRadius: '50%',
                        backgroundColor: getCategoryColor(tool.category),
                        flexShrink: 0
                      }}
                    />
                    <div>
                      <div style={{
                        fontWeight: '600',
                        color: '#111827',
                        marginBottom: '0.25rem',
                        fontSize: '0.875rem'
                      }}>
                        {tool.name}
                        {tool.is_public && (
                          <span style={{
                            marginLeft: '0.5rem',
                            padding: '0.125rem 0.25rem',
                            fontSize: '0.625rem',
                            fontWeight: '500',
                            backgroundColor: '#dbeafe',
                            color: '#1e40af',
                            borderRadius: '0.25rem'
                          }}>
                            PUBLIC
                          </span>
                        )}
                      </div>
                      <div style={{
                        display: 'inline-block',
                        padding: '0.125rem 0.5rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        textTransform: 'uppercase',
                        letterSpacing: '0.025em',
                        backgroundColor: getCategoryColor(tool.category) + '20',
                        color: getCategoryColor(tool.category)
                      }}>
                        {tool.category}
                      </div>
                    </div>
                  </div>
                </td>

                <td style={{ padding: '1rem 0.75rem' }}>
                  <div style={{ fontSize: '0.875rem' }}>
                    <div style={{
                      color: '#374151',
                      marginBottom: '0.25rem',
                      fontFamily: 'monospace',
                      fontWeight: '500'
                    }}>
                      {tool.function_name}
                    </div>
                    <div style={{
                      display: 'inline-block',
                      padding: '0.125rem 0.5rem',
                      borderRadius: '9999px',
                      fontSize: '0.75rem',
                      fontWeight: '500',
                      textTransform: 'uppercase',
                      letterSpacing: '0.025em',
                      backgroundColor: getImplementationTypeColor(tool.implementation_type) + '20',
                      color: getImplementationTypeColor(tool.implementation_type),
                      marginBottom: '0.25rem'
                    }}>
                      {tool.implementation_type}
                    </div>
                    {tool.description && (
                      <div style={{ color: '#6b7280', fontSize: '0.75rem' }}>
                        {truncateText(tool.description, 60)}
                      </div>
                    )}
                    {tool.tags && tool.tags.length > 0 && (
                      <div style={{ marginTop: '0.5rem', display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
                        {tool.tags.slice(0, 2).map((tag, index) => (
                          <span
                            key={index}
                            style={{
                              padding: '0.125rem 0.375rem',
                              backgroundColor: '#f3f4f6',
                              color: '#374151',
                              borderRadius: '0.375rem',
                              fontSize: '0.625rem',
                              fontWeight: '500'
                            }}
                          >
                            {tag}
                          </span>
                        ))}
                        {tool.tags.length > 2 && (
                          <span style={{ fontSize: '0.625rem', color: '#6b7280' }}>
                            +{tool.tags.length - 2}
                          </span>
                        )}
                      </div>
                    )}
                  </div>
                </td>

                <td style={{ padding: '1rem 0.75rem' }}>
                  <div style={{ fontSize: '0.875rem' }}>
                    <div style={{
                      color: '#374151',
                      marginBottom: '0.5rem',
                      fontWeight: '500'
                    }}>
                      {tool.usage_count.toLocaleString()} uses
                    </div>
                    <div
                      style={{
                        display: 'inline-block',
                        padding: '0.25rem 0.75rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: '600',
                        backgroundColor: tool.is_active ? '#dcfce7' : '#fef2f2',
                        color: tool.is_active ? '#166534' : '#dc2626'
                      }}
                    >
                      {tool.is_active ? 'Active' : 'Inactive'}
                    </div>
                    <div style={{
                      color: '#6b7280',
                      fontSize: '0.75rem',
                      marginTop: '0.25rem'
                    }}>
                      {tool.timeout_seconds}s timeout • {tool.max_retries} retries
                    </div>
                  </div>
                </td>

                <td style={{ padding: '1rem 0.75rem', fontSize: '0.875rem', color: '#6b7280' }}>
                  {formatDate(tool.updated_at)}
                </td>

                <td style={{ padding: '1rem 0.75rem', textAlign: 'right' }}>
                  <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                    <button
                      onClick={() => onView(tool)}
                      style={{
                        padding: '0.375rem 0.75rem',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        color: '#374151',
                        backgroundColor: '#f9fafb',
                        border: '1px solid #e5e7eb',
                        borderRadius: '0.375rem',
                        cursor: 'pointer',
                        transition: 'all 0.2s'
                      }}
                      onMouseOver={(e) => {
                        e.currentTarget.style.backgroundColor = '#f3f4f6';
                        e.currentTarget.style.borderColor = '#d1d5db';
                      }}
                      onMouseOut={(e) => {
                        e.currentTarget.style.backgroundColor = '#f9fafb';
                        e.currentTarget.style.borderColor = '#e5e7eb';
                      }}
                    >
                      View
                    </button>
                    <button
                      onClick={() => onEdit(tool)}
                      style={{
                        padding: '0.375rem 0.75rem',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        color: '#3b82f6',
                        backgroundColor: 'transparent',
                        border: '1px solid #3b82f6',
                        borderRadius: '0.375rem',
                        cursor: 'pointer',
                        transition: 'all 0.2s'
                      }}
                      onMouseOver={(e) => {
                        e.currentTarget.style.backgroundColor = '#3b82f6';
                        e.currentTarget.style.color = 'white';
                      }}
                      onMouseOut={(e) => {
                        e.currentTarget.style.backgroundColor = 'transparent';
                        e.currentTarget.style.color = '#3b82f6';
                      }}
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => setShowConfirmDelete(tool)}
                      disabled={deletingId === tool.id}
                      style={{
                        padding: '0.375rem 0.75rem',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        color: deletingId === tool.id ? '#9ca3af' : '#dc2626',
                        backgroundColor: 'transparent',
                        border: `1px solid ${deletingId === tool.id ? '#9ca3af' : '#dc2626'}`,
                        borderRadius: '0.375rem',
                        cursor: deletingId === tool.id ? 'not-allowed' : 'pointer',
                        transition: 'all 0.2s'
                      }}
                      onMouseOver={(e) => {
                        if (deletingId !== tool.id) {
                          e.currentTarget.style.backgroundColor = '#dc2626';
                          e.currentTarget.style.color = 'white';
                        }
                      }}
                      onMouseOut={(e) => {
                        if (deletingId !== tool.id) {
                          e.currentTarget.style.backgroundColor = 'transparent';
                          e.currentTarget.style.color = '#dc2626';
                        }
                      }}
                    >
                      {deletingId === tool.id ? 'Deleting...' : 'Delete'}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Confirmation Modal */}
      {showConfirmDelete && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000
        }}>
          <div style={{
            backgroundColor: 'white',
            padding: '1.5rem',
            borderRadius: '8px',
            maxWidth: '400px',
            width: '90%',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
          }}>
            <h3 style={{ margin: '0 0 1rem 0', fontSize: '1.125rem', fontWeight: '600', color: '#111827' }}>
              Delete Tool
            </h3>
            <p style={{ margin: '0 0 1.5rem 0', color: '#6b7280', fontSize: '0.875rem' }}>
              Are you sure you want to delete &quot;{showConfirmDelete.name}&quot;? This action cannot be undone.
            </p>
            <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'flex-end' }}>
              <button
                onClick={() => setShowConfirmDelete(null)}
                disabled={deletingId === showConfirmDelete.id}
                style={{
                  padding: '0.5rem 1rem',
                  fontSize: '0.875rem',
                  color: '#374151',
                  backgroundColor: '#f9fafb',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.375rem',
                  cursor: deletingId === showConfirmDelete.id ? 'not-allowed' : 'pointer',
                  opacity: deletingId === showConfirmDelete.id ? 0.5 : 1
                }}
              >
                Cancel
              </button>
              <button
                onClick={() => handleDeleteConfirm(showConfirmDelete)}
                disabled={deletingId === showConfirmDelete.id}
                style={{
                  padding: '0.5rem 1rem',
                  fontSize: '0.875rem',
                  color: 'white',
                  backgroundColor: deletingId === showConfirmDelete.id ? '#9ca3af' : '#dc2626',
                  border: 'none',
                  borderRadius: '0.375rem',
                  cursor: deletingId === showConfirmDelete.id ? 'not-allowed' : 'pointer'
                }}
              >
                {deletingId === showConfirmDelete.id ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
