'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/components/AuthContext';
import { authApi, User, UpdateProfileRequest, ApiKey, CreateApiKeyRequest } from '@/lib/api';

export default function ProfilePage() {
  const { user, isAuthenticated } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<'profile' | 'security' | 'api-keys'>('profile');
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  // Profile form state
  const [profileForm, setProfileForm] = useState({
    email: '',
    current_password: '',
    new_password: '',
    confirm_password: '',
  });

  // API Keys state
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [isLoadingKeys, setIsLoadingKeys] = useState(false);
  const [showCreateKeyForm, setShowCreateKeyForm] = useState(false);
  const [newKeyForm, setNewKeyForm] = useState({
    name: '',
    role: 'user' as 'admin' | 'user' | 'viewer',
    expires_at: '',
  });
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  // Initialize profile form with user data
  useEffect(() => {
    if (user) {
      setProfileForm(prev => ({
        ...prev,
        email: user.email,
      }));
    }
  }, [user]);

  // Load API keys when API Keys tab is active
  useEffect(() => {
    if (activeTab === 'api-keys') {
      loadApiKeys();
    }
  }, [activeTab]);

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

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();

    if (profileForm.new_password && profileForm.new_password !== profileForm.confirm_password) {
      setMessage({ type: 'error', text: 'New passwords do not match' });
      return;
    }

    setIsLoading(true);
    setMessage(null);

    try {
      const updates: UpdateProfileRequest = {};

      if (profileForm.email !== user?.email) {
        updates.email = profileForm.email;
      }

      if (profileForm.new_password) {
        updates.current_password = profileForm.current_password;
        updates.new_password = profileForm.new_password;
      }

      if (Object.keys(updates).length > 0) {
        await authApi.updateProfile(updates);
        setMessage({ type: 'success', text: 'Profile updated successfully' });
        setProfileForm(prev => ({
          ...prev,
          current_password: '',
          new_password: '',
          confirm_password: '',
        }));
      } else {
        setMessage({ type: 'error', text: 'No changes to save' });
      }
    } catch (error) {
      setMessage({ type: 'error', text: error instanceof Error ? error.message : 'Failed to update profile' });
    } finally {
      setIsLoading(false);
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

  const handleDeleteApiKey = async (keyId: string) => {
    if (!confirm('Are you sure you want to delete this API key? This action cannot be undone.')) {
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
        <p>Please log in to view your profile.</p>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto', padding: '2rem' }}>
      {/* Header */}
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          Profile Settings
        </h1>
        <p style={{ color: '#6b7280' }}>
          Manage your account settings and preferences.
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
            padding: '0.5rem',
            borderRadius: '4px',
            fontFamily: 'monospace',
            fontSize: '0.875rem',
            wordBreak: 'break-all',
            color: '#92400e',
          }}>
            {createdKey}
          </div>
          <button
            onClick={() => setCreatedKey(null)}
            style={{
              marginTop: '0.5rem',
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
      )}

      {/* Tabs */}
      <div style={{ borderBottom: '1px solid #e5e7eb', marginBottom: '2rem' }}>
        <div style={{ display: 'flex', gap: '2rem' }}>
          {[
            { id: 'profile', label: 'Profile' },
            { id: 'security', label: 'Security' },
            { id: 'api-keys', label: 'API Keys' },
          ].map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              style={{
                padding: '0.75rem 0',
                border: 'none',
                background: 'none',
                fontSize: '0.875rem',
                fontWeight: activeTab === tab.id ? '600' : '400',
                color: activeTab === tab.id ? '#3b82f6' : '#6b7280',
                borderBottom: activeTab === tab.id ? '2px solid #3b82f6' : '2px solid transparent',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Profile Tab */}
      {activeTab === 'profile' && (
        <div style={{ backgroundColor: 'white', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb' }}>
          <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '1rem', color: '#111827' }}>
            Account Information
          </h3>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem', marginBottom: '2rem' }}>
            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                User ID
              </label>
              <div style={{
                padding: '0.5rem 0.75rem',
                backgroundColor: '#f9fafb',
                border: '1px solid #e5e7eb',
                borderRadius: '4px',
                fontSize: '0.875rem',
                color: '#6b7280',
                fontFamily: 'monospace',
              }}>
                {user.id}
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                Role
              </label>
              <div style={{
                padding: '0.5rem 0.75rem',
                backgroundColor: '#f9fafb',
                border: '1px solid #e5e7eb',
                borderRadius: '4px',
                fontSize: '0.875rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
              }}>
                <span style={{
                  backgroundColor: user?.role === 'admin' ? '#dc2626' : user?.role === 'user' ? '#3b82f6' : '#6b7280',
                  color: 'white',
                  padding: '0.125rem 0.5rem',
                  borderRadius: '12px',
                  fontSize: '0.75rem',
                  fontWeight: '500',
                  textTransform: 'capitalize',
                }}>
                  {user.role}
                </span>
              </div>
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                Organization ID
              </label>
              <div style={{
                padding: '0.5rem 0.75rem',
                backgroundColor: '#f9fafb',
                border: '1px solid #e5e7eb',
                borderRadius: '4px',
                fontSize: '0.875rem',
                color: '#6b7280',
                fontFamily: 'monospace',
              }}>
                {user.organization_id}
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                Account Created
              </label>
              <div style={{
                padding: '0.5rem 0.75rem',
                backgroundColor: '#f9fafb',
                border: '1px solid #e5e7eb',
                borderRadius: '4px',
                fontSize: '0.875rem',
                color: '#6b7280',
              }}>
                {new Date(user.created_at).toLocaleDateString()}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Security Tab */}
      {activeTab === 'security' && (
        <div style={{ backgroundColor: 'white', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb' }}>
          <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '1rem', color: '#111827' }}>
            Security Settings
          </h3>

          <form onSubmit={handleProfileUpdate}>
            <div style={{ marginBottom: '1.5rem' }}>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                Email Address
              </label>
              <input
                type="email"
                value={profileForm.email}
                onChange={(e) => setProfileForm(prev => ({ ...prev, email: e.target.value }))}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem',
                }}
                required
              />
            </div>

            <div style={{ marginBottom: '1.5rem' }}>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                Current Password
              </label>
              <input
                type="password"
                value={profileForm.current_password}
                onChange={(e) => setProfileForm(prev => ({ ...prev, current_password: e.target.value }))}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem',
                }}
                placeholder="Enter current password to make changes"
              />
            </div>

            <div style={{ marginBottom: '1.5rem' }}>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                New Password
              </label>
              <input
                type="password"
                value={profileForm.new_password}
                onChange={(e) => setProfileForm(prev => ({ ...prev, new_password: e.target.value }))}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem',
                }}
                placeholder="Leave blank to keep current password"
              />
            </div>

            <div style={{ marginBottom: '1.5rem' }}>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                Confirm New Password
              </label>
              <input
                type="password"
                value={profileForm.confirm_password}
                onChange={(e) => setProfileForm(prev => ({ ...prev, confirm_password: e.target.value }))}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem',
                }}
                placeholder="Confirm new password"
              />
            </div>

            <button
              type="submit"
              disabled={isLoading}
              style={{
                backgroundColor: '#3b82f6',
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
              {isLoading ? 'Updating...' : 'Update Profile'}
            </button>
          </form>
        </div>
      )}

      {/* API Keys Tab */}
      {activeTab === 'api-keys' && (
        <div style={{ backgroundColor: 'white', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <h3 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#111827' }}>
              API Keys
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
              Create New Key
            </button>
          </div>

          {/* Create Key Form */}
          {showCreateKeyForm && (
            <div style={{
              backgroundColor: '#f9fafb',
              padding: '1rem',
              borderRadius: '6px',
              marginBottom: '1.5rem',
              border: '1px solid #e5e7eb',
            }}>
              <form onSubmit={handleCreateApiKey}>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
                  <div>
                    <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                      Key Name
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
                      }}
                      placeholder="e.g., Production API"
                      required
                    />
                  </div>

                  <div>
                    <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                      Role
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
                      }}
                    >
                      <option value="viewer">Viewer</option>
                      <option value="user">User</option>
                      <option value="admin">Admin</option>
                    </select>
                  </div>

                  <div>
                    <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
                      Expires At (Optional)
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
                      }}
                    />
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '0.5rem' }}>
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
                    {isLoading ? 'Creating...' : 'Create Key'}
                  </button>

                  <button
                    type="button"
                    onClick={() => setShowCreateKeyForm(false)}
                    style={{
                      backgroundColor: '#6b7280',
                      color: 'white',
                      padding: '0.5rem 1rem',
                      border: 'none',
                      borderRadius: '4px',
                      fontSize: '0.875rem',
                      fontWeight: '500',
                      cursor: 'pointer',
                    }}
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* API Keys List */}
          {isLoadingKeys ? (
            <div style={{ textAlign: 'center', padding: '2rem', color: '#6b7280' }}>
              Loading API keys...
            </div>
          ) : apiKeys.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2rem', color: '#6b7280' }}>
              No API keys found. Create your first API key to get started.
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
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.25rem' }}>
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
                    </div>
                    <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>
                      Created: {new Date(key.created_at).toLocaleDateString()}{' '}
                      {key.expires_at && `• Expires: ${new Date(key.expires_at).toLocaleDateString()}`}{' '}
                      {key.last_used_at ? `• Last used: ${new Date(key.last_used_at).toLocaleDateString()}` : '• Never used'}
                    </div>
                  </div>

                  <button
                    onClick={() => handleDeleteApiKey(key.id)}
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
      )}
    </div>
  );
}
