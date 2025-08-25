'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { namespaceApi, serverApi, endpointApi } from '../../../lib/api';
import { Toast } from '../../../components/Toast';
import { EditNamespaceModal } from '../../../components/namespaces/EditNamespaceModal';
import { NamespaceToolsManager } from '../../../components/namespaces/NamespaceToolsManager';
import { CreateEndpointModal } from '../../../components/endpoints/CreateEndpointModal';

interface NamespaceServer {
  server_id: string;
  server_name: string;
  status: string;
  priority: number;
  joined_at: string;
}

interface Namespace {
  id: string;
  name: string;
  description: string;
  servers: NamespaceServer[];
  created_at: string;
  updated_at: string;
  is_active: boolean;
  metadata?: Record<string, any>;
}

interface Server {
  id: string;
  name: string;
  protocol: string;
  status: string;
  description?: string;
}

interface Endpoint {
  id: string;
  name: string;
  description?: string;
  enable_api_key_auth: boolean;
  enable_oauth: boolean;
  enable_public_access: boolean;
  use_query_param_auth: boolean;
  is_active: boolean;
  urls?: {
    sse: string;
    http: string;
    websocket: string;
    openapi: string;
    documentation: string;
  };
}

export default function NamespaceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const namespaceId = params.id as string;

  const [namespace, setNamespace] = useState<Namespace | null>(null);
  const [servers, setServers] = useState<Server[]>([]);
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showCreateEndpointModal, setShowCreateEndpointModal] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const fetchNamespace = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await namespaceApi.getNamespace(namespaceId);
      setNamespace(data);

      // Fetch server details if namespace has servers
      if (data.servers && data.servers.length > 0) {
        const serverPromises = data.servers.map(serverInfo =>
          serverApi.getServer(serverInfo.server_id).catch(() => null)
        );
        const serverData = await Promise.all(serverPromises);
        setServers(serverData.filter(s => s !== null) as Server[]);
      }

      // Fetch endpoints for this namespace
      try {
        const allEndpoints = await endpointApi.listEndpoints();
        const namespaceEndpoints = allEndpoints.filter(ep => ep.namespace_id === namespaceId);
        setEndpoints(namespaceEndpoints);
      } catch (err) {
        console.error('Failed to fetch endpoints:', err);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch namespace details';
      setError(errorMessage);
      setToast({ message: errorMessage, type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNamespace();
  }, [namespaceId]);

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this namespace?')) {
      return;
    }

    try {
      await namespaceApi.deleteNamespace(namespaceId);
      setToast({ message: 'Namespace deleted successfully', type: 'success' });
      router.push('/namespaces');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete namespace';
      setToast({ message: errorMessage, type: 'error' });
    }
  };

  const handleRemoveServer = async (serverId: string) => {
    if (!confirm('Are you sure you want to remove this server from the namespace?')) {
      return;
    }

    try {
      await namespaceApi.removeServerFromNamespace(namespaceId, serverId);
      setToast({ message: 'Server removed from namespace', type: 'success' });
      fetchNamespace();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to remove server';
      setToast({ message: errorMessage, type: 'error' });
    }
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

  if (loading) {
    return (
      <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '2rem' }}>
        <div style={{ textAlign: 'center', padding: '3rem', color: '#6b7280' }}>
          Loading namespace details...
        </div>
      </div>
    );
  }

  if (error || !namespace) {
    return (
      <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '2rem' }}>
        <div style={{
          backgroundColor: '#fef2f2',
          color: '#dc2626',
          padding: '1rem',
          borderRadius: '0.375rem'
        }}>
          {error || 'Namespace not found'}
        </div>
        <Link
          href="/namespaces"
          style={{
            display: 'inline-block',
            marginTop: '1rem',
            color: '#3b82f6',
            textDecoration: 'none'
          }}
        >
          ← Back to Namespaces
        </Link>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '2rem' }}>
      {/* Header */}
      <div style={{ marginBottom: '2rem' }}>
        <Link
          href="/namespaces"
          style={{
            color: '#3b82f6',
            textDecoration: 'none',
            fontSize: '0.875rem',
            display: 'inline-flex',
            alignItems: 'center',
            marginBottom: '1rem'
          }}
          onMouseEnter={(e) => e.currentTarget.style.textDecoration = 'underline'}
          onMouseLeave={(e) => e.currentTarget.style.textDecoration = 'none'}
        >
          ← Back to Namespaces
        </Link>

        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start'
        }}>
          <div>
            <h1 style={{
              fontSize: '1.875rem',
              fontWeight: 'bold',
              color: '#111827',
              marginBottom: '0.5rem'
            }}>
              {namespace.name}
            </h1>
            {namespace.description && (
              <p style={{ color: '#6b7280' }}>
                {namespace.description}
              </p>
            )}
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button
              onClick={() => setShowEditModal(true)}
              style={{
                padding: '0.625rem 1.25rem',
                backgroundColor: 'white',
                color: '#374151',
                border: '1px solid #e5e7eb',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: 'pointer',
                transition: 'all 0.2s'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.backgroundColor = '#f9fafb';
                e.currentTarget.style.borderColor = '#d1d5db';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.backgroundColor = 'white';
                e.currentTarget.style.borderColor = '#e5e7eb';
              }}
            >
              Edit
            </button>
            <button
              onClick={handleDelete}
              style={{
                padding: '0.625rem 1.25rem',
                backgroundColor: '#dc2626',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: 'pointer',
                transition: 'background-color 0.2s'
              }}
              onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#b91c1c'}
              onMouseLeave={(e) => e.currentTarget.style.backgroundColor = '#dc2626'}
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      {/* Basic Information */}
      <div style={{
        backgroundColor: 'white',
        border: '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        marginBottom: '1.5rem'
      }}>
        <h2 style={{
          fontSize: '1.125rem',
          fontWeight: '600',
          color: '#111827',
          marginBottom: '1rem'
        }}>
          Basic Information
        </h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '1rem' }}>
          <div>
            <label style={{ fontSize: '0.875rem', color: '#6b7280' }}>ID</label>
            <p style={{ fontSize: '0.875rem', color: '#111827', fontFamily: 'monospace' }}>
              {namespace.id}
            </p>
          </div>
          <div>
            <label style={{ fontSize: '0.875rem', color: '#6b7280' }}>Status</label>
            <p>
              <span style={{
                display: 'inline-flex',
                alignItems: 'center',
                padding: '0.25rem 0.75rem',
                backgroundColor: namespace.is_active ? '#d1fae5' : '#fee2e2',
                color: namespace.is_active ? '#065f46' : '#991b1b',
                borderRadius: '9999px',
                fontSize: '0.75rem',
                fontWeight: '500'
              }}>
                {namespace.is_active ? 'Active' : 'Inactive'}
              </span>
            </p>
          </div>
          <div>
            <label style={{ fontSize: '0.875rem', color: '#6b7280' }}>Created</label>
            <p style={{ fontSize: '0.875rem', color: '#111827' }}>
              {formatDate(namespace.created_at)}
            </p>
          </div>
          <div>
            <label style={{ fontSize: '0.875rem', color: '#6b7280' }}>Last Updated</label>
            <p style={{ fontSize: '0.875rem', color: '#111827' }}>
              {formatDate(namespace.updated_at)}
            </p>
          </div>
        </div>
      </div>

      {/* Servers */}
      <div style={{
        backgroundColor: 'white',
        border: '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        marginBottom: '1.5rem'
      }}>
        <h2 style={{
          fontSize: '1.125rem',
          fontWeight: '600',
          color: '#111827',
          marginBottom: '1rem'
        }}>
          Servers ({servers.length})
        </h2>
        {servers.length === 0 ? (
          <p style={{ color: '#6b7280', fontSize: '0.875rem' }}>
            No servers added to this namespace yet.
          </p>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid #e5e7eb' }}>
                  <th style={{ padding: '0.75rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                    Name
                  </th>
                  <th style={{ padding: '0.75rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                    Protocol
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
                {servers.map((server) => (
                  <tr key={server.id} style={{ borderBottom: '1px solid #e5e7eb' }}>
                    <td style={{ padding: '0.75rem' }}>
                      <Link
                        href={`/servers/${server.id}`}
                        style={{
                          color: '#3b82f6',
                          textDecoration: 'none',
                          fontWeight: '500',
                          fontSize: '0.875rem'
                        }}
                        onMouseEnter={(e) => e.currentTarget.style.textDecoration = 'underline'}
                        onMouseLeave={(e) => e.currentTarget.style.textDecoration = 'none'}
                      >
                        {server.name}
                      </Link>
                      {server.description && (
                        <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
                          {server.description}
                        </p>
                      )}
                    </td>
                    <td style={{ padding: '0.75rem', color: '#6b7280', fontSize: '0.875rem' }}>
                      {server.protocol}
                    </td>
                    <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '0.25rem 0.75rem',
                        backgroundColor: server.status === 'active' ? '#d1fae5' : '#fee2e2',
                        color: server.status === 'active' ? '#065f46' : '#991b1b',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: '500'
                      }}>
                        {server.status}
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                      <button
                        onClick={() => handleRemoveServer(server.id)}
                        style={{
                          padding: '0.375rem 0.75rem',
                          backgroundColor: 'white',
                          color: '#dc2626',
                          border: '1px solid #fca5a5',
                          borderRadius: '0.375rem',
                          fontSize: '0.75rem',
                          fontWeight: '500',
                          cursor: 'pointer',
                          transition: 'all 0.2s'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.backgroundColor = '#fef2f2';
                          e.currentTarget.style.borderColor = '#f87171';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.backgroundColor = 'white';
                          e.currentTarget.style.borderColor = '#fca5a5';
                        }}
                      >
                        Remove
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Endpoints */}
      <div style={{
        backgroundColor: 'white',
        border: '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        marginBottom: '1.5rem'
      }}>
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '1rem'
        }}>
          <h2 style={{
            fontSize: '1.125rem',
            fontWeight: '600',
            color: '#111827'
          }}>
            Endpoints ({endpoints.length})
          </h2>
          <button
            onClick={() => setShowCreateEndpointModal(true)}
            style={{
              padding: '0.375rem 0.75rem',
              backgroundColor: '#3b82f6',
              color: 'white',
              border: 'none',
              borderRadius: '0.375rem',
              fontSize: '0.75rem',
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

        {endpoints.length === 0 ? (
          <div style={{
            padding: '2rem',
            textAlign: 'center',
            backgroundColor: '#f9fafb',
            borderRadius: '0.375rem'
          }}>
            <svg
              style={{ width: '2.5rem', height: '2.5rem', margin: '0 auto', color: '#9ca3af' }}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            <p style={{ color: '#6b7280', fontSize: '0.875rem', marginTop: '1rem' }}>
              No endpoints created for this namespace yet.
            </p>
            <p style={{ color: '#6b7280', fontSize: '0.875rem', marginTop: '0.5rem' }}>
              Endpoints provide external access to namespace tools via various transport protocols.
            </p>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid #e5e7eb' }}>
                  <th style={{ padding: '0.75rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                    Name
                  </th>
                  <th style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                    Authentication
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
                  <tr key={endpoint.id} style={{ borderBottom: '1px solid #e5e7eb' }}>
                    <td style={{ padding: '0.75rem' }}>
                      <Link
                        href={`/endpoints`}
                        style={{
                          color: '#3b82f6',
                          textDecoration: 'none',
                          fontWeight: '500',
                          fontSize: '0.875rem'
                        }}
                        onMouseEnter={(e) => e.currentTarget.style.textDecoration = 'underline'}
                        onMouseLeave={(e) => e.currentTarget.style.textDecoration = 'none'}
                      >
                        {endpoint.name}
                      </Link>
                      {endpoint.description && (
                        <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
                          {endpoint.description}
                        </p>
                      )}
                      {endpoint.urls && (
                        <div style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
                          <span style={{ fontFamily: 'monospace' }}>
                            {endpoint.urls.http}
                          </span>
                        </div>
                      )}
                    </td>
                    <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                      <div style={{ display: 'flex', justifyContent: 'center', gap: '0.25rem', flexWrap: 'wrap' }}>
                        {endpoint.enable_public_access && (
                          <span style={{
                            display: 'inline-block',
                            padding: '0.125rem 0.375rem',
                            backgroundColor: '#fef3c7',
                            color: '#92400e',
                            borderRadius: '0.25rem',
                            fontSize: '0.625rem',
                            fontWeight: '500'
                          }}>
                            Public
                          </span>
                        )}
                        {endpoint.enable_api_key_auth && (
                          <span style={{
                            display: 'inline-block',
                            padding: '0.125rem 0.375rem',
                            backgroundColor: '#dbeafe',
                            color: '#1e40af',
                            borderRadius: '0.25rem',
                            fontSize: '0.625rem',
                            fontWeight: '500'
                          }}>
                            API Key
                          </span>
                        )}
                        {endpoint.enable_oauth && (
                          <span style={{
                            display: 'inline-block',
                            padding: '0.125rem 0.375rem',
                            backgroundColor: '#e9d5ff',
                            color: '#6b21a8',
                            borderRadius: '0.25rem',
                            fontSize: '0.625rem',
                            fontWeight: '500'
                          }}>
                            OAuth
                          </span>
                        )}
                      </div>
                    </td>
                    <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '0.25rem 0.5rem',
                        backgroundColor: endpoint.is_active ? '#d1fae5' : '#fee2e2',
                        color: endpoint.is_active ? '#065f46' : '#991b1b',
                        borderRadius: '9999px',
                        fontSize: '0.625rem',
                        fontWeight: '500'
                      }}>
                        {endpoint.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                      <Link
                        href={`/endpoints`}
                        style={{
                          padding: '0.375rem 0.75rem',
                          backgroundColor: 'white',
                          color: '#3b82f6',
                          border: '1px solid #3b82f6',
                          borderRadius: '0.375rem',
                          fontSize: '0.75rem',
                          fontWeight: '500',
                          textDecoration: 'none',
                          display: 'inline-block'
                        }}
                      >
                        Manage
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Tools Management */}
      <NamespaceToolsManager namespaceId={namespaceId} />

      {/* Edit Modal */}
      {showEditModal && (
        <EditNamespaceModal
          namespace={namespace}
          onClose={() => setShowEditModal(false)}
          onUpdate={() => {
            setShowEditModal(false);
            fetchNamespace();
          }}
        />
      )}

      {/* Create Endpoint Modal */}
      {showCreateEndpointModal && (
        <CreateEndpointModal
          preselectedNamespaceId={namespaceId}
          onClose={() => setShowCreateEndpointModal(false)}
          onCreate={() => {
            setShowCreateEndpointModal(false);
            fetchNamespace();
            setToast({ message: 'Endpoint created successfully', type: 'success' });
          }}
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
