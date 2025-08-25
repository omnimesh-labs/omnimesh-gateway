'use client';

import { useState, useEffect } from 'react';
import { endpointApi, namespaceApi } from '../../lib/api';

interface CreateEndpointModalProps {
  onClose: () => void;
  onCreate: () => void;
  preselectedNamespaceId?: string;
}

export function CreateEndpointModal({ onClose, onCreate, preselectedNamespaceId }: CreateEndpointModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [namespaceId, setNamespaceId] = useState(preselectedNamespaceId || '');
  const [namespaces, setNamespaces] = useState<Array<{ id: string; name: string; description?: string }>>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Authentication settings
  const [enableApiKeyAuth, setEnableApiKeyAuth] = useState(true);
  const [enableOauth, setEnableOauth] = useState(false);
  const [enablePublicAccess, setEnablePublicAccess] = useState(false);
  const [useQueryParamAuth, setUseQueryParamAuth] = useState(false);

  // Rate limiting settings
  const [rateLimitRequests, setRateLimitRequests] = useState(100);
  const [rateLimitWindow, setRateLimitWindow] = useState(60);

  // CORS settings
  const [allowedOrigins, setAllowedOrigins] = useState('*');
  const [allowedMethods, setAllowedMethods] = useState('GET,POST,OPTIONS');

  useEffect(() => {
    fetchNamespaces();
  }, []);

  const fetchNamespaces = async () => {
    try {
      const data = await namespaceApi.listNamespaces();
      setNamespaces(data);
    } catch (error) {
      console.error('Failed to fetch namespaces:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await endpointApi.createEndpoint({
        namespace_id: namespaceId,
        name: name.toLowerCase().replace(/[^a-z0-9-_]/g, '-'),
        description,
        enable_api_key_auth: enableApiKeyAuth,
        enable_oauth: enableOauth,
        enable_public_access: enablePublicAccess,
        use_query_param_auth: useQueryParamAuth,
        rate_limit_requests: rateLimitRequests,
        rate_limit_window: rateLimitWindow,
        allowed_origins: allowedOrigins.split(',').map(s => s.trim()),
        allowed_methods: allowedMethods.split(',').map(s => s.trim()),
      });
      onCreate();
      onClose();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to create endpoint';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // Handle escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [onClose]);

  return (
    <>
      {/* Modal Overlay */}
      <div
        style={{
          position: 'fixed',
          inset: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          zIndex: 9998,
          animation: 'fadeIn 0.2s ease-out'
        }}
        onClick={onClose}
      />

      {/* Modal Content */}
      <div
        style={{
          position: 'fixed',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          padding: '1.5rem',
          width: '90%',
          maxWidth: '600px',
          maxHeight: '90vh',
          overflowY: 'auto',
          zIndex: 9999,
          boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
        }}
      >
        <h2 style={{
          fontSize: '1.25rem',
          fontWeight: '600',
          color: '#111827',
          marginBottom: '1rem'
        }}>
          Create Endpoint
        </h2>

        <form onSubmit={handleSubmit}>
          {/* Namespace Selection */}
          <div style={{ marginBottom: '1rem' }}>
            <label style={{
              display: 'block',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              marginBottom: '0.25rem'
            }}>
              Namespace *
            </label>
            <select
              value={namespaceId}
              onChange={(e) => setNamespaceId(e.target.value)}
              required
              disabled={!!preselectedNamespaceId}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                backgroundColor: preselectedNamespaceId ? '#f9fafb' : 'white'
              }}
            >
              <option value="">Select a namespace</option>
              {namespaces.map((ns) => (
                <option key={ns.id} value={ns.id}>
                  {ns.name} {ns.description ? `- ${ns.description}` : ''}
                </option>
              ))}
            </select>
          </div>

          {/* Name */}
          <div style={{ marginBottom: '1rem' }}>
            <label style={{
              display: 'block',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              marginBottom: '0.25rem'
            }}>
              Endpoint Name *
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              placeholder="my-endpoint"
              pattern="[a-zA-Z0-9-_]+"
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem'
              }}
            />
            <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
              URL-safe name (alphanumeric, hyphens, underscores). Will be accessible at /api/public/endpoints/{name}/
            </p>
          </div>

          {/* Description */}
          <div style={{ marginBottom: '1rem' }}>
            <label style={{
              display: 'block',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              marginBottom: '0.25rem'
            }}>
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe what this endpoint provides..."
              rows={3}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                resize: 'vertical'
              }}
            />
          </div>

          {/* Authentication Settings */}
          <div style={{
            marginBottom: '1rem',
            padding: '1rem',
            backgroundColor: '#f9fafb',
            borderRadius: '0.375rem'
          }}>
            <h3 style={{
              fontSize: '0.875rem',
              fontWeight: '600',
              color: '#111827',
              marginBottom: '0.75rem'
            }}>
              Authentication Settings
            </h3>

            <div style={{ marginBottom: '0.5rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={enableApiKeyAuth}
                  onChange={(e) => setEnableApiKeyAuth(e.target.checked)}
                  style={{ marginRight: '0.5rem' }}
                />
                <span style={{ fontSize: '0.875rem' }}>Enable API Key Authentication</span>
              </label>
            </div>

            {enableApiKeyAuth && (
              <div style={{ marginBottom: '0.5rem', marginLeft: '1.5rem' }}>
                <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={useQueryParamAuth}
                    onChange={(e) => setUseQueryParamAuth(e.target.checked)}
                    style={{ marginRight: '0.5rem' }}
                  />
                  <span style={{ fontSize: '0.875rem' }}>Allow API key in query parameters</span>
                </label>
                <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
                  Enables ?api_key=... in addition to Authorization header
                </p>
              </div>
            )}

            <div style={{ marginBottom: '0.5rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={enableOauth}
                  onChange={(e) => setEnableOauth(e.target.checked)}
                  style={{ marginRight: '0.5rem' }}
                />
                <span style={{ fontSize: '0.875rem' }}>Enable OAuth Authentication</span>
              </label>
            </div>

            <div style={{ marginBottom: '0.5rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={enablePublicAccess}
                  onChange={(e) => setEnablePublicAccess(e.target.checked)}
                  style={{ marginRight: '0.5rem' }}
                />
                <span style={{ fontSize: '0.875rem' }}>Enable Public Access (No Auth Required)</span>
              </label>
              <p style={{ fontSize: '0.75rem', color: '#dc2626', marginTop: '0.25rem' }}>
                ⚠️ Warning: This allows anyone to access this endpoint without authentication
              </p>
            </div>
          </div>

          {/* Rate Limiting */}
          <div style={{
            marginBottom: '1rem',
            padding: '1rem',
            backgroundColor: '#f9fafb',
            borderRadius: '0.375rem'
          }}>
            <h3 style={{
              fontSize: '0.875rem',
              fontWeight: '600',
              color: '#111827',
              marginBottom: '0.75rem'
            }}>
              Rate Limiting
            </h3>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.75rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Requests
                </label>
                <input
                  type="number"
                  value={rateLimitRequests}
                  onChange={(e) => setRateLimitRequests(parseInt(e.target.value) || 100)}
                  min={1}
                  style={{
                    width: '100%',
                    padding: '0.375rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem'
                  }}
                />
              </div>
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.75rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Window (seconds)
                </label>
                <input
                  type="number"
                  value={rateLimitWindow}
                  onChange={(e) => setRateLimitWindow(parseInt(e.target.value) || 60)}
                  min={1}
                  style={{
                    width: '100%',
                    padding: '0.375rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem'
                  }}
                />
              </div>
            </div>
            <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
              Allow {rateLimitRequests} requests per {rateLimitWindow} seconds
            </p>
          </div>

          {/* CORS Settings */}
          <div style={{
            marginBottom: '1rem',
            padding: '1rem',
            backgroundColor: '#f9fafb',
            borderRadius: '0.375rem'
          }}>
            <h3 style={{
              fontSize: '0.875rem',
              fontWeight: '600',
              color: '#111827',
              marginBottom: '0.75rem'
            }}>
              CORS Settings
            </h3>

            <div style={{ marginBottom: '0.5rem' }}>
              <label style={{
                display: 'block',
                fontSize: '0.75rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.25rem'
              }}>
                Allowed Origins (comma-separated)
              </label>
              <input
                type="text"
                value={allowedOrigins}
                onChange={(e) => setAllowedOrigins(e.target.value)}
                placeholder="*, https://example.com"
                style={{
                  width: '100%',
                  padding: '0.375rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem'
                }}
              />
            </div>

            <div>
              <label style={{
                display: 'block',
                fontSize: '0.75rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.25rem'
              }}>
                Allowed Methods (comma-separated)
              </label>
              <input
                type="text"
                value={allowedMethods}
                onChange={(e) => setAllowedMethods(e.target.value)}
                placeholder="GET, POST, OPTIONS"
                style={{
                  width: '100%',
                  padding: '0.375rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem'
                }}
              />
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div style={{
              padding: '0.75rem',
              backgroundColor: '#fef2f2',
              border: '1px solid #fecaca',
              borderRadius: '0.375rem',
              marginBottom: '1rem'
            }}>
              <p style={{ color: '#dc2626', fontSize: '0.875rem' }}>{error}</p>
            </div>
          )}

          {/* Actions */}
          <div style={{
            display: 'flex',
            justifyContent: 'flex-end',
            gap: '0.5rem',
            marginTop: '1.5rem'
          }}>
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              style={{
                padding: '0.625rem 1.25rem',
                backgroundColor: 'white',
                color: '#374151',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: loading ? 'not-allowed' : 'pointer',
                opacity: loading ? 0.5 : 1
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !namespaceId || !name}
              style={{
                padding: '0.625rem 1.25rem',
                backgroundColor: '#3b82f6',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: loading || !namespaceId || !name ? 'not-allowed' : 'pointer',
                opacity: loading || !namespaceId || !name ? 0.5 : 1
              }}
            >
              {loading ? 'Creating...' : 'Create Endpoint'}
            </button>
          </div>
        </form>
      </div>

      <style jsx>{`
        @keyframes fadeIn {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }
      `}</style>
    </>
  );
}
