'use client';

import { useState } from 'react';
import { useAuth } from '@/components/AuthContext';
import { authApi, UpdateProfileRequest } from '@/lib/api';

export default function ProfileSettingsPage() {
  const { user, isAuthenticated } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  // Profile form state
  const [profileForm, setProfileForm] = useState({
    email: user?.email || '',
    current_password: '',
    new_password: '',
    confirm_password: '',
  });

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();

    if (profileForm.new_password && profileForm.new_password !== profileForm.confirm_password) {
      setMessage({ type: 'error', text: 'New passwords do not match' });
      return;
    }

    if (!profileForm.current_password && (profileForm.email !== user?.email || profileForm.new_password)) {
      setMessage({ type: 'error', text: 'Current password is required to make changes' });
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

  if (!isAuthenticated || !user) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <p>Please log in to view your profile settings.</p>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '600px', margin: '0 auto', padding: '2rem' }}>
      {/* Header */}
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          Account Settings
        </h1>
        <p style={{ color: '#6b7280' }}>
          Update your email address and password.
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

      {/* Settings Form */}
      <div style={{ backgroundColor: 'white', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb' }}>
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
                boxSizing: 'border-box',
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
                boxSizing: 'border-box',
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
                boxSizing: 'border-box',
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
                boxSizing: 'border-box',
              }}
              placeholder="Confirm new password"
            />
          </div>

          <div style={{ display: 'flex', gap: '0.75rem' }}>
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
              {isLoading ? 'Updating...' : 'Update Settings'}
            </button>

            <button
              type="button"
              onClick={() => window.history.back()}
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

      {/* Account Information */}
      <div style={{ backgroundColor: 'white', padding: '1.5rem', borderRadius: '8px', border: '1px solid #e5e7eb', marginTop: '1.5rem' }}>
        <h3 style={{ fontSize: '1.125rem', fontWeight: '600', marginBottom: '1rem', color: '#111827' }}>
          Account Information
        </h3>

        <div style={{ display: 'grid', gap: '1rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 0', borderBottom: '1px solid #f3f4f6' }}>
            <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>User ID</span>
            <span style={{ fontSize: '0.875rem', color: '#111827', fontFamily: 'monospace' }}>{user.id}</span>
          </div>

          <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 0', borderBottom: '1px solid #f3f4f6' }}>
            <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>Role</span>
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

          <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 0', borderBottom: '1px solid #f3f4f6' }}>
            <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>Organization</span>
            <span style={{ fontSize: '0.875rem', color: '#111827', fontFamily: 'monospace' }}>{user.organization_id}</span>
          </div>

          <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 0' }}>
            <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>Member Since</span>
            <span style={{ fontSize: '0.875rem', color: '#111827' }}>{new Date(user.created_at).toLocaleDateString()}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
