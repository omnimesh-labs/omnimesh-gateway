'use client';

import { Resource } from '@/lib/api';

interface ResourceCardProps {
  resource: Resource;
  isOpen: boolean;
  onClose: () => void;
  onEdit: () => void;
}

export function ResourceCard({ resource, isOpen, onClose, onEdit }: ResourceCardProps) {
  if (!isOpen) return null;

  const getResourceTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      file: '#8b5cf6',     // purple
      url: '#3b82f6',      // blue
      database: '#10b981', // green
      api: '#f59e0b',      // yellow
      memory: '#ef4444',   // red
      custom: '#6b7280'    // gray
    };
    return colors[type] || colors.custom;
  };

  const formatBytes = (bytes?: number) => {
    if (!bytes) return 'Unknown';
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      timeZoneName: 'short'
    });
  };

  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
        padding: '1rem'
      }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div style={{
        backgroundColor: 'white',
        borderRadius: '8px',
        width: '100%',
        maxWidth: '700px',
        maxHeight: '90vh',
        overflow: 'hidden',
        boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
      }}>
        {/* Header */}
        <div style={{
          padding: '1.5rem',
          borderBottom: '1px solid #e5e7eb',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div
              style={{
                width: '12px',
                height: '12px',
                borderRadius: '50%',
                backgroundColor: getResourceTypeColor(resource.resource_type),
                flexShrink: 0
              }}
            />
            <h2 style={{
              margin: 0,
              fontSize: '1.25rem',
              fontWeight: '600',
              color: '#111827'
            }}>
              {resource.name}
            </h2>
            <div style={{
              display: 'inline-block',
              padding: '0.25rem 0.5rem',
              borderRadius: '9999px',
              fontSize: '0.75rem',
              fontWeight: '500',
              textTransform: 'uppercase',
              letterSpacing: '0.025em',
              backgroundColor: getResourceTypeColor(resource.resource_type) + '20',
              color: getResourceTypeColor(resource.resource_type)
            }}>
              {resource.resource_type}
            </div>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button
              onClick={onEdit}
              style={{
                padding: '0.5rem 1rem',
                fontSize: '0.875rem',
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
              onClick={onClose}
              style={{
                padding: '0.5rem',
                backgroundColor: 'transparent',
                border: 'none',
                cursor: 'pointer',
                color: '#6b7280',
                fontSize: '1.25rem',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center'
              }}
              onMouseOver={(e) => e.currentTarget.style.color = '#374151'}
              onMouseOut={(e) => e.currentTarget.style.color = '#6b7280'}
            >
              ×
            </button>
          </div>
        </div>

        {/* Content */}
        <div style={{
          padding: '1.5rem',
          maxHeight: 'calc(90vh - 140px)',
          overflowY: 'auto'
        }}>
          {/* Status and Basic Info */}
          <div style={{ marginBottom: '2rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
              <div
                style={{
                  display: 'inline-block',
                  padding: '0.25rem 0.75rem',
                  borderRadius: '9999px',
                  fontSize: '0.75rem',
                  fontWeight: '600',
                  backgroundColor: resource.is_active ? '#dcfce7' : '#fef2f2',
                  color: resource.is_active ? '#166534' : '#dc2626'
                }}
              >
                {resource.is_active ? 'Active' : 'Inactive'}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                ID: {resource.id}
              </div>
            </div>

            {resource.description && (
              <p style={{
                color: '#374151',
                fontSize: '0.875rem',
                lineHeight: '1.5',
                margin: '0 0 1rem 0'
              }}>
                {resource.description}
              </p>
            )}
          </div>

          {/* Resource Details */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
            gap: '1.5rem',
            marginBottom: '2rem'
          }}>
            {/* URI */}
            <div>
              <h4 style={{
                margin: '0 0 0.5rem 0',
                fontSize: '0.875rem',
                fontWeight: '600',
                color: '#374151'
              }}>
                URI
              </h4>
              <p style={{
                margin: 0,
                fontSize: '0.875rem',
                color: '#111827',
                wordBreak: 'break-all',
                backgroundColor: '#f9fafb',
                padding: '0.5rem',
                borderRadius: '0.375rem',
                border: '1px solid #e5e7eb'
              }}>
                {resource.uri}
              </p>
            </div>

            {/* MIME Type */}
            {resource.mime_type && (
              <div>
                <h4 style={{
                  margin: '0 0 0.5rem 0',
                  fontSize: '0.875rem',
                  fontWeight: '600',
                  color: '#374151'
                }}>
                  MIME Type
                </h4>
                <p style={{
                  margin: 0,
                  fontSize: '0.875rem',
                  color: '#111827'
                }}>
                  {resource.mime_type}
                </p>
              </div>
            )}

            {/* Size */}
            {resource.size_bytes && (
              <div>
                <h4 style={{
                  margin: '0 0 0.5rem 0',
                  fontSize: '0.875rem',
                  fontWeight: '600',
                  color: '#374151'
                }}>
                  Size
                </h4>
                <p style={{
                  margin: 0,
                  fontSize: '0.875rem',
                  color: '#111827'
                }}>
                  {formatBytes(resource.size_bytes)}
                </p>
              </div>
            )}
          </div>

          {/* Tags */}
          {resource.tags && resource.tags.length > 0 && (
            <div style={{ marginBottom: '2rem' }}>
              <h4 style={{
                margin: '0 0 0.5rem 0',
                fontSize: '0.875rem',
                fontWeight: '600',
                color: '#374151'
              }}>
                Tags
              </h4>
              <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                {resource.tags.map((tag, index) => (
                  <span
                    key={index}
                    style={{
                      padding: '0.25rem 0.5rem',
                      backgroundColor: '#f3f4f6',
                      color: '#374151',
                      borderRadius: '0.375rem',
                      fontSize: '0.75rem',
                      fontWeight: '500'
                    }}
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Metadata */}
          {resource.metadata && Object.keys(resource.metadata).length > 0 && (
            <div style={{ marginBottom: '2rem' }}>
              <h4 style={{
                margin: '0 0 0.5rem 0',
                fontSize: '0.875rem',
                fontWeight: '600',
                color: '#374151'
              }}>
                Metadata
              </h4>
              <pre style={{
                margin: 0,
                fontSize: '0.75rem',
                color: '#111827',
                backgroundColor: '#f9fafb',
                padding: '0.75rem',
                borderRadius: '0.375rem',
                border: '1px solid #e5e7eb',
                overflow: 'auto',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace'
              }}>
                {JSON.stringify(resource.metadata, null, 2)}
              </pre>
            </div>
          )}

          {/* Access Permissions */}
          {resource.access_permissions && Object.keys(resource.access_permissions).length > 0 && (
            <div style={{ marginBottom: '2rem' }}>
              <h4 style={{
                margin: '0 0 0.5rem 0',
                fontSize: '0.875rem',
                fontWeight: '600',
                color: '#374151'
              }}>
                Access Permissions
              </h4>
              <pre style={{
                margin: 0,
                fontSize: '0.75rem',
                color: '#111827',
                backgroundColor: '#f9fafb',
                padding: '0.75rem',
                borderRadius: '0.375rem',
                border: '1px solid #e5e7eb',
                overflow: 'auto',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace'
              }}>
                {JSON.stringify(resource.access_permissions, null, 2)}
              </pre>
            </div>
          )}

          {/* Timestamps */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: '1rem',
            padding: '1rem',
            backgroundColor: '#f9fafb',
            borderRadius: '0.375rem',
            border: '1px solid #e5e7eb'
          }}>
            <div>
              <h4 style={{
                margin: '0 0 0.25rem 0',
                fontSize: '0.75rem',
                fontWeight: '600',
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Created
              </h4>
              <p style={{
                margin: 0,
                fontSize: '0.875rem',
                color: '#111827'
              }}>
                {formatDate(resource.created_at)}
              </p>
            </div>
            <div>
              <h4 style={{
                margin: '0 0 0.25rem 0',
                fontSize: '0.75rem',
                fontWeight: '600',
                color: '#6b7280',
                textTransform: 'uppercase',
                letterSpacing: '0.05em'
              }}>
                Updated
              </h4>
              <p style={{
                margin: 0,
                fontSize: '0.875rem',
                color: '#111827'
              }}>
                {formatDate(resource.updated_at)}
              </p>
            </div>
            {resource.created_by && (
              <div>
                <h4 style={{
                  margin: '0 0 0.25rem 0',
                  fontSize: '0.75rem',
                  fontWeight: '600',
                  color: '#6b7280',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em'
                }}>
                  Created By
                </h4>
                <p style={{
                  margin: 0,
                  fontSize: '0.875rem',
                  color: '#111827'
                }}>
                  {resource.created_by}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
