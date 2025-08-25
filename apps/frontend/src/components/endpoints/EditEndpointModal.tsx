'use client';

import { useState, useEffect } from 'react';
import { endpointApi } from '../../lib/api';

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
}

interface EditEndpointModalProps {
  endpoint: Endpoint;
  onClose: () => void;
  onUpdate: () => void;
}

export function EditEndpointModal({ endpoint, onClose, onUpdate }: EditEndpointModalProps) {
  const [description, setDescription] = useState(endpoint.description || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Authentication settings
  const [enableApiKeyAuth, setEnableApiKeyAuth] = useState(endpoint.enable_api_key_auth);
  const [enableOauth, setEnableOauth] = useState(endpoint.enable_oauth);
  const [enablePublicAccess, setEnablePublicAccess] = useState(endpoint.enable_public_access);
  const [useQueryParamAuth, setUseQueryParamAuth] = useState(endpoint.use_query_param_auth);

  // Rate limiting settings
  const [rateLimitRequests, setRateLimitRequests] = useState(endpoint.rate_limit_requests);
  const [rateLimitWindow, setRateLimitWindow] = useState(endpoint.rate_limit_window);

  // CORS settings
  const [allowedOrigins, setAllowedOrigins] = useState(endpoint.allowed_origins?.join(', ') || '*');
  const [allowedMethods, setAllowedMethods] = useState(endpoint.allowed_methods?.join(', ') || 'GET,POST,OPTIONS');

  // Status
  const [isActive, setIsActive] = useState(endpoint.is_active);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await endpointApi.updateEndpoint(endpoint.id, {
        description,
        enable_api_key_auth: enableApiKeyAuth,
        enable_oauth: enableOauth,
        enable_public_access: enablePublicAccess,
        use_query_param_auth: useQueryParamAuth,
        rate_limit_requests: rateLimitRequests,
        rate_limit_window: rateLimitWindow,
        allowed_origins: allowedOrigins.split(',').map(s => s.trim()),
        allowed_methods: allowedMethods.split(',').map(s => s.trim()),
        is_active: isActive,
      });
      onUpdate();
      onClose();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to update endpoint';
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
          Edit Endpoint: {endpoint.name}
        </h2>

        <form onSubmit={handleSubmit}>
          {/* Read-only fields */}
          <div style={{ marginBottom: '1rem' }}>
            <label style={{
              display: 'block',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              marginBottom: '0.25rem'
            }}>
              Namespace
            </label>
            <input
              type="text"
              value={endpoint.namespace?.name || 'Unknown'}
              disabled
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                backgroundColor: '#f9fafb',
                color: '#6b7280'
              }}
            />
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

          {/* Status */}
          <div style={{
            marginBottom: '1rem',
            padding: '1rem',
            backgroundColor: '#f9fafb',
            borderRadius: '0.375rem'
          }}>
            <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
              <input
                type="checkbox"
                checked={isActive}
                onChange={(e) => setIsActive(e.target.checked)}
                style={{ marginRight: '0.5rem' }}
              />
              <span style={{ fontSize: '0.875rem', fontWeight: '500' }}>Endpoint is Active</span>
            </label>
            <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem', marginLeft: '1.5rem' }}>
              Inactive endpoints will not accept any requests
            </p>
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
              disabled={loading}
              style={{
                padding: '0.625rem 1.25rem',
                backgroundColor: '#3b82f6',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: loading ? 'not-allowed' : 'pointer',
                opacity: loading ? 0.5 : 1
              }}
            >
              {loading ? 'Updating...' : 'Update Endpoint'}
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
