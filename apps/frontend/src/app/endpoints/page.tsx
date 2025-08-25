'use client';

import { useState, useEffect } from 'react';
import { endpointApi } from '../../lib/api';
import { Toast } from '../../components/Toast';
import { CreateEndpointModal } from '../../components/endpoints/CreateEndpointModal';
import { EndpointTable } from '../../components/endpoints/EndpointTable';

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

export default function EndpointsPage() {
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const fetchEndpoints = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await endpointApi.listEndpoints();
      setEndpoints(data || []);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch endpoints';
      setError(errorMessage);
      setToast({ message: errorMessage, type: 'error' });
      setEndpoints([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEndpoints();
  }, []);

  const handleCreateEndpoint = async () => {
    await fetchEndpoints();
    setToast({ message: 'Endpoint created successfully', type: 'success' });
  };

  const handleDeleteEndpoint = async (id: string) => {
    try {
      await endpointApi.deleteEndpoint(id);
      await fetchEndpoints();
      setToast({ message: 'Endpoint deleted successfully', type: 'success' });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete endpoint';
      setToast({ message: errorMessage, type: 'error' });
    }
  };

  const handleUpdateEndpoint = async () => {
    await fetchEndpoints();
    setToast({ message: 'Endpoint updated successfully', type: 'success' });
  };

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '2rem' }}>
      {/* Header */}
      <div style={{
        marginBottom: '2rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <div>
          <h1 style={{
            fontSize: '1.875rem',
            fontWeight: 'bold',
            color: '#111827',
            marginBottom: '0.5rem'
          }}>
            Endpoints
          </h1>
          <p style={{ color: '#6b7280' }}>
            Manage public-facing endpoints for your namespaces. Endpoints provide authenticated access to namespace tools via various transport protocols.
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          style={{
            padding: '0.625rem 1.25rem',
            backgroundColor: '#3b82f6',
            color: 'white',
            border: 'none',
            borderRadius: '0.375rem',
            fontSize: '0.875rem',
            fontWeight: '500',
            cursor: 'pointer',
            transition: 'background-color 0.2s'
          }}
          onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#2563eb'}
          onMouseLeave={(e) => e.currentTarget.style.backgroundColor = '#3b82f6'}
        >
          Create Endpoint
        </button>
      </div>

      {/* Loading State */}
      {loading && (
        <div style={{
          backgroundColor: 'white',
          border: '1px solid #e5e7eb',
          borderRadius: '0.5rem',
          padding: '3rem',
          textAlign: 'center'
        }}>
          <p style={{ color: '#6b7280' }}>Loading endpoints...</p>
        </div>
      )}

      {/* Error State */}
      {error && !loading && (
        <div style={{
          backgroundColor: '#fef2f2',
          border: '1px solid #fecaca',
          borderRadius: '0.5rem',
          padding: '1rem',
          marginBottom: '1rem'
        }}>
          <p style={{ color: '#dc2626' }}>{error}</p>
        </div>
      )}

      {/* Empty State */}
      {!loading && !error && endpoints.length === 0 && (
        <div style={{
          backgroundColor: 'white',
          border: '2px dashed #e5e7eb',
          borderRadius: '0.5rem',
          padding: '3rem',
          textAlign: 'center'
        }}>
          <svg
            style={{ width: '3rem', height: '3rem', margin: '0 auto', color: '#9ca3af' }}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <h3 style={{
            fontSize: '1.125rem',
            fontWeight: '600',
            color: '#111827',
            marginTop: '1rem',
            marginBottom: '0.5rem'
          }}>
            No endpoints yet
          </h3>
          <p style={{ color: '#6b7280', marginBottom: '1.5rem' }}>
            Create your first endpoint to enable external access to your namespace tools.
          </p>
          <button
            onClick={() => setShowCreateModal(true)}
            style={{
              padding: '0.625rem 1.25rem',
              backgroundColor: '#3b82f6',
              color: 'white',
              border: 'none',
              borderRadius: '0.375rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#2563eb'}
            onMouseLeave={(e) => e.currentTarget.style.backgroundColor = '#3b82f6'}
          >
            Create Your First Endpoint
          </button>
        </div>
      )}

      {/* Endpoints Table */}
      {!loading && !error && endpoints.length > 0 && (
        <EndpointTable
          endpoints={endpoints}
          onDelete={handleDeleteEndpoint}
          onUpdate={handleUpdateEndpoint}
        />
      )}

      {/* Create Endpoint Modal */}
      {showCreateModal && (
        <CreateEndpointModal
          onClose={() => setShowCreateModal(false)}
          onCreate={handleCreateEndpoint}
        />
      )}

      {/* Toast Notifications */}
      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </div>
  );
}
