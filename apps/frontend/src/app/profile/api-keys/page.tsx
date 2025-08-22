'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/components/AuthContext';
import { authApi, ApiKey, CreateApiKeyRequest } from '@/lib/api';

export default function ApiKeysPage() {
  const { user, isAuthenticated } = useAuth();
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingKeys, setIsLoadingKeys] = useState(true);
  const [showCreateKeyForm, setShowCreateKeyForm] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  
  const [newKeyForm, setNewKeyForm] = useState({
    name: '',
    role: 'user' as 'admin' | 'user' | 'viewer',
    expires_at: '',
  });

  // Load API keys on component mount
  useEffect(() => {
    loadApiKeys();
  }, []);

  const loadApiKeys = async () => {
    setIsLoadingKeys(true);
    try {
      const keys = await authApi.listApiKeys();
      setApiKeys(keys);
    } catch (error) {
      console.error('Failed to load API keys:', error);
      setMessage({ type: 'error', text: 'Failed to load API keys' });
    } finally {
      setIsLoadingKeys(false);
    }
  };

  const handleCreateApiKey = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setMessage(null);
    
    try {
      const keyData: CreateApiKeyRequest = {
        name: newKeyForm.name,
        role: newKeyForm.role,
        expires_at: newKeyForm.expires_at || undefined,
      };
      
      const response = await authApi.createApiKey(keyData);
      setCreatedKey(response.key);
      setMessage({ type: 'success', text: 'API key created successfully' });
      setNewKeyForm({ name: '', role: 'user', expires_at: '' });
      setShowCreateKeyForm(false);
      loadApiKeys();
    } catch (error) {
      setMessage({ type: 'error', text: error instanceof Error ? error.message : 'Failed to create API key' });
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteApiKey = async (keyId: string, keyName: string) => {
    if (!confirm(`Are you sure you want to delete the API key "${keyName}"? This action cannot be undone.`)) {
      return;
    }

    try {
      await authApi.deleteApiKey(keyId);
      setMessage({ type: 'success', text: 'API key deleted successfully' });
      loadApiKeys();
    } catch (error) {
      setMessage({ type: 'error', text: error instanceof Error ? error.message : 'Failed to delete API key' });
    }
  };

  if (!isAuthenticated || !user) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <p>Please log in to manage your API keys.</p>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto', padding: '2rem' }}>
      {/* Header */}
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          API Keys
        </h1>
        <p style={{ color: '#6b7280' }}>
          Create and manage API keys for programmatic access to MCP Gateway.
        </p>
      </div>

      {/* Message */}
      {message && (
        <div style={{
          padding: '0.75rem 1rem',
          borderRadius: '6px',
          marginBottom: '1.5rem',
          backgroundColor: message.type === 'success' ? '#f0fdf4' : '#fef2f2',
          border: `1px solid ${message.type === 'success' ? '#bbf7d0' : '#fecaca'}`,
          color: message.type === 'success' ? '#166534' : '#dc2626',
        }}>
          {message.text}
        </div>
      )}

      {/* Created API Key Display */}
      {createdKey && (
        <div style={{
          padding: '1rem',
          borderRadius: '6px',
          marginBottom: '1.5rem',
          backgroundColor: '#fffbeb',
          border: '1px solid #fed7aa',
        }}>
          <h4 style={{ color: '#92400e', marginBottom: '0.5rem', fontSize: '0.875rem', fontWeight: '600' }}>
            API Key Created - Save This Key!
          </h4>
          <p style={{ color: '#92400e', fontSize: '0.875rem', marginBottom: '0.5rem' }}>
            This key will only be shown once. Please copy and save it securely.
          </p>
          <div style={{
            backgroundColor: '#fef3c7',
            padding: '0.75rem',
            borderRadius: '4px',
            fontFamily: 'monospace',
            fontSize: '0.875rem',
            wordBreak: 'break-all',
            color: '#92400e',
            border: '1px solid #fbbf24',
          }}>
            {createdKey}
          </div>
          <div style={{ marginTop: '0.75rem', display: 'flex', gap: '0.5rem' }}>
            <button
              onClick={() => navigator.clipboard.writeText(createdKey)}
              style={{
                fontSize: '0.75rem',
                color: '#92400e',
                backgroundColor: '#fef3c7',
                border: '1px solid #fbbf24',
                borderRadius: '4px',
                padding: '0.25rem 0.5rem',
                cursor: 'pointer',
              }}
            >
              Copy to Clipboard
            </button>
            <button
              onClick={() => setCreatedKey(null)}
              style={{
                fontSize: '0.75rem',
                color: '#92400e',
                backgroundColor: 'transparent',
                border: 'none',
                cursor: 'pointer',
                textDecoration: 'underline',
              }}
            >
              Dismiss
            </button>
          </div>
        </div>
      )}

      {/* Create New Key Section */}
      <div style={{ backgroundColor: 'white', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb', marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <h3 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#111827' }}>
            Create New API Key
          </h3>
          <button
            onClick={() => setShowCreateKeyForm(!showCreateKeyForm)}
            style={{
              backgroundColor: '#3b82f6',
              color: 'white',
              padding: '0.5rem 1rem',
              border: 'none',
              borderRadius: '4px',
              fontSize: '0.875rem',
              fontWeight: '500',
              cursor: 'pointer',
            }}
          >
            {showCreateKeyForm ? 'Cancel' : 'Create New Key'}
          </button>
        </div>

        {/* Create Key Form */}
        {showCreateKeyForm && (
          <form onSubmit={handleCreateApiKey}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                  Key Name *
                </label>
                <input
                  type="text"
                  value={newKeyForm.name}
                  onChange={(e) => setNewKeyForm(prev => ({ ...prev, name: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '4px',
                    fontSize: '0.875rem',
                    boxSizing: 'border-box',
                  }}
                  placeholder="e.g., Production API, Development"
                  required
                />
              </div>
              
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                  Role *
                </label>
                <select
                  value={newKeyForm.role}
                  onChange={(e) => setNewKeyForm(prev => ({ ...prev, role: e.target.value as any }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '4px',
                    fontSize: '0.875rem',
                    boxSizing: 'border-box',
                  }}
                >
                  <option value="viewer">Viewer - Read-only access</option>
                  <option value="user">User - Standard access</option>
                  <option value="admin">Admin - Full access</option>
                </select>
              </div>
            </div>
            
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                Expiration Date (Optional)
              </label>
              <input
                type="datetime-local"
                value={newKeyForm.expires_at}
                onChange={(e) => setNewKeyForm(prev => ({ ...prev, expires_at: e.target.value }))}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem',
                  boxSizing: 'border-box',
                }}
              />
              <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
                Leave blank for no expiration
              </p>
            </div>
            
            <button
              type="submit"
              disabled={isLoading}
              style={{
                backgroundColor: '#10b981',
                color: 'white',
                padding: '0.5rem 1rem',
                border: 'none',
                borderRadius: '4px',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: isLoading ? 'not-allowed' : 'pointer',
                opacity: isLoading ? 0.7 : 1,
              }}
            >
              {isLoading ? 'Creating...' : 'Create API Key'}
            </button>
          </form>
        )}
      </div>

      {/* Existing API Keys */}
      <div style={{ backgroundColor: 'white', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb' }}>
        <h3 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
          Your API Keys
        </h3>

        {isLoadingKeys ? (
          <div style={{ textAlign: 'center', padding: '2rem', color: '#6b7280' }}>
            Loading API keys...
          </div>
        ) : apiKeys.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2rem', color: '#6b7280' }}>
            <svg style={{ width: '48px', height: '48px', margin: '0 auto 1rem', color: '#d1d5db' }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
            <p>No API keys found.</p>
            <p style={{ fontSize: '0.875rem' }}>Create your first API key to get started with programmatic access.</p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {apiKeys.map((key) => (
              <div
                key={key.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '1rem',
                  border: '1px solid #e5e7eb',
                  borderRadius: '6px',
                  backgroundColor: key.is_active ? 'white' : '#f9fafb',
                }}
              >
                <div style={{ flex: 1 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.5rem' }}>
                    <h4 style={{ fontSize: '0.875rem', fontWeight: '600', color: '#111827' }}>
                      {key.name}
                    </h4>
                    <span style={{
                      backgroundColor: key.role === 'admin' ? '#dc2626' : key.role === 'user' ? '#3b82f6' : '#6b7280',
                      color: 'white',
                      padding: '0.125rem 0.5rem',
                      borderRadius: '12px',
                      fontSize: '0.75rem',
                      fontWeight: '500',
                      textTransform: 'capitalize',
                    }}>
                      {key.role}
                    </span>
                    {!key.is_active && (
                      <span style={{
                        backgroundColor: '#ef4444',
                        color: 'white',
                        padding: '0.125rem 0.5rem',
                        borderRadius: '12px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                      }}>
                        Inactive
                      </span>
                    )}
                    {key.expires_at && new Date(key.expires_at) < new Date() && (
                      <span style={{
                        backgroundColor: '#f59e0b',
                        color: 'white',
                        padding: '0.125rem 0.5rem',
                        borderRadius: '12px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                      }}>
                        Expired
                      </span>
                    )}
                  </div>
                  
                  <div style={{ fontSize: '0.75rem', color: '#6b7280', display: 'flex', flexWrap: 'wrap', gap: '1rem' }}>
                    <span>Created: {new Date(key.created_at).toLocaleDateString()}</span>
                    {key.expires_at && (
                      <span>Expires: {new Date(key.expires_at).toLocaleDateString()}</span>
                    )}
                    <span>
                      {key.last_used_at ? `Last used: ${new Date(key.last_used_at).toLocaleDateString()}` : 'Never used'}
                    </span>
                  </div>
                  
                  <div style={{
                    fontSize: '0.75rem',
                    color: '#6b7280',
                    fontFamily: 'monospace',
                    marginTop: '0.25rem',
                  }}>
                    {key.key_hash.substring(0, 8)}...
                  </div>
                </div>
                
                <button
                  onClick={() => handleDeleteApiKey(key.id, key.name)}
                  style={{
                    backgroundColor: '#ef4444',
                    color: 'white',
                    padding: '0.25rem 0.75rem',
                    border: 'none',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: '500',
                    cursor: 'pointer',
                  }}
                >
                  Delete
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Usage Guidelines */}
      <div style={{ backgroundColor: '#f9fafb', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb', marginTop: '1.5rem' }}>
        <h4 style={{ fontSize: '1rem', fontWeight: '600', color: '#111827', marginBottom: '0.75rem' }}>
          API Key Usage
        </h4>
        <div style={{ fontSize: '0.875rem', color: '#6b7280', lineHeight: '1.5' }}>
          <p style={{ marginBottom: '0.5rem' }}>
            • Include your API key in the <code style={{ backgroundColor: '#e5e7eb', padding: '0.125rem 0.25rem', borderRadius: '3px' }}>Authorization</code> header: <code style={{ backgroundColor: '#e5e7eb', padding: '0.125rem 0.25rem', borderRadius: '3px' }}>Bearer your-api-key</code>
          </p>
          <p style={{ marginBottom: '0.5rem' }}>
            • API keys inherit the permissions of your user role and the key's assigned role (whichever is more restrictive)
          </p>
          <p style={{ marginBottom: '0.5rem' }}>
            • Keys can be set to expire automatically for enhanced security
          </p>
          <p>
            • Delete unused keys regularly to maintain security
          </p>
        </div>
      </div>
    </div>
  );
}