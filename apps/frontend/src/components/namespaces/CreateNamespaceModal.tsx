'use client';

import { useState, useEffect } from 'react';
import { serverApi } from '../../lib/api';

interface CreateNamespaceModalProps {
  onClose: () => void;
  onCreate: (data: {
    name: string;
    description: string;
    server_ids: string[];
  }) => void;
}

interface Server {
  id: string;
  name: string;
  protocol: string;
  status: string;
}

export function CreateNamespaceModal({ onClose, onCreate }: CreateNamespaceModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [selectedServers, setSelectedServers] = useState<string[]>([]);
  const [availableServers, setAvailableServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingServers, setLoadingServers] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchAvailableServers();
  }, []);

  const fetchAvailableServers = async () => {
    try {
      const servers = await serverApi.listServers();
      setAvailableServers(servers);
    } catch (error) {
      console.error('Failed to fetch servers:', error);
      setError('Failed to load available servers');
    } finally {
      setLoadingServers(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name.trim()) {
      setError('Namespace name is required');
      return;
    }

    if (!name.match(/^[a-zA-Z0-9_-]+$/)) {
      setError('Namespace name can only contain letters, numbers, hyphens, and underscores');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await onCreate({
        name: name.trim(),
        description: description.trim(),
        server_ids: selectedServers
      });
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to create namespace');
      setLoading(false);
    }
  };

  const toggleServerSelection = (serverId: string) => {
    setSelectedServers(prev =>
      prev.includes(serverId)
        ? prev.filter(id => id !== serverId)
        : [...prev, serverId]
    );
  };

  return (
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
        borderRadius: '0.5rem',
        padding: '2rem',
        width: '90%',
        maxWidth: '600px',
        maxHeight: '90vh',
        overflow: 'auto'
      }}>
        <h2 style={{
          fontSize: '1.5rem',
          fontWeight: 'bold',
          marginBottom: '1.5rem',
          color: '#111827'
        }}>
          Create Namespace
        </h2>

        {error && (
          <div style={{
            backgroundColor: '#fef2f2',
            color: '#dc2626',
            padding: '0.75rem',
            borderRadius: '0.375rem',
            marginBottom: '1rem',
            fontSize: '0.875rem'
          }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {/* Name Field */}
          <div style={{ marginBottom: '1.5rem' }}>
            <label style={{
              display: 'block',
              marginBottom: '0.5rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151'
            }}>
              Name *
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-namespace"
              style={{
                width: '100%',
                padding: '0.625rem',
                border: '1px solid #e5e7eb',
                borderRadius: '0.375rem',
                fontSize: '0.875rem'
              }}
              required
            />
            <p style={{
              marginTop: '0.25rem',
              fontSize: '0.75rem',
              color: '#6b7280'
            }}>
              Only letters, numbers, hyphens, and underscores allowed
            </p>
          </div>

          {/* Description Field */}
          <div style={{ marginBottom: '1.5rem' }}>
            <label style={{
              display: 'block',
              marginBottom: '0.5rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151'
            }}>
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe the purpose of this namespace..."
              rows={3}
              style={{
                width: '100%',
                padding: '0.625rem',
                border: '1px solid #e5e7eb',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                resize: 'vertical'
              }}
            />
          </div>

          {/* Server Selection */}
          <div style={{ marginBottom: '1.5rem' }}>
            <label style={{
              display: 'block',
              marginBottom: '0.5rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151'
            }}>
              Select Servers
            </label>
            <div style={{
              border: '1px solid #e5e7eb',
              borderRadius: '0.375rem',
              maxHeight: '200px',
              overflow: 'auto',
              padding: '0.5rem'
            }}>
              {loadingServers ? (
                <div style={{ padding: '1rem', textAlign: 'center', color: '#6b7280' }}>
                  Loading servers...
                </div>
              ) : availableServers.length === 0 ? (
                <div style={{ padding: '1rem', textAlign: 'center', color: '#6b7280' }}>
                  No servers available. Create servers first.
                </div>
              ) : (
                availableServers.map(server => (
                  <label
                    key={server.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      padding: '0.5rem',
                      cursor: 'pointer',
                      borderRadius: '0.25rem',
                      transition: 'background-color 0.2s'
                    }}
                    onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f9fafb'}
                    onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                  >
                    <input
                      type="checkbox"
                      checked={selectedServers.includes(server.id)}
                      onChange={() => toggleServerSelection(server.id)}
                      style={{ marginRight: '0.75rem' }}
                    />
                    <div style={{ flex: 1 }}>
                      <div style={{ fontWeight: '500', fontSize: '0.875rem', color: '#111827' }}>
                        {server.name}
                      </div>
                      <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>
                        {server.protocol} • {server.status}
                      </div>
                    </div>
                  </label>
                ))
              )}
            </div>
            <p style={{
              marginTop: '0.25rem',
              fontSize: '0.75rem',
              color: '#6b7280'
            }}>
              {selectedServers.length} server(s) selected
            </p>
          </div>

          {/* Action Buttons */}
          <div style={{
            display: 'flex',
            gap: '1rem',
            justifyContent: 'flex-end'
          }}>
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              style={{
                padding: '0.625rem 1.25rem',
                backgroundColor: 'white',
                color: '#374151',
                border: '1px solid #e5e7eb',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: loading ? 'not-allowed' : 'pointer',
                opacity: loading ? 0.5 : 1,
                transition: 'all 0.2s'
              }}
              onMouseEnter={(e) => !loading && (e.currentTarget.style.backgroundColor = '#f9fafb')}
              onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'white'}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim()}
              style={{
                padding: '0.625rem 1.25rem',
                backgroundColor: '#3b82f6',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: loading || !name.trim() ? 'not-allowed' : 'pointer',
                opacity: loading || !name.trim() ? 0.5 : 1,
                transition: 'background-color 0.2s'
              }}
              onMouseEnter={(e) => !loading && name.trim() && (e.currentTarget.style.backgroundColor = '#2563eb')}
              onMouseLeave={(e) => e.currentTarget.style.backgroundColor = '#3b82f6'}
            >
              {loading ? 'Creating...' : 'Create Namespace'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}