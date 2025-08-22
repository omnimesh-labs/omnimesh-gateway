'use client';

import { useState, useEffect } from 'react';
import { ServerTable } from '@/components/servers/ServerTable';
import { AvailableServersList } from '@/components/servers/AvailableServersList';
import { RegisterServerModal } from '@/components/servers/RegisterServerModal';
import { HealthCheck } from '@/components/HealthCheck';
import { useToast } from '@/components/Toast';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { serverApi, type MCPServer } from '@/lib/api';

export default function ServersPage() {
  const [registeredServers, setRegisteredServers] = useState<MCPServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'registered' | 'available'>('registered');
  const [showRegisterModal, setShowRegisterModal] = useState(false);
  const { success, error: showError, ToastContainer } = useToast();

  const loadServers = async () => {
    try {
      setLoading(true);
      const servers = await serverApi.listServers();
      setRegisteredServers(servers);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load servers');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadServers();
  }, []);

  const handleServerRegistered = (server: MCPServer) => {
    setRegisteredServers(prev => [...prev, server]);
    setShowRegisterModal(false);
    success(`Server "${server.name}" registered successfully!`);
  };

  const handleServerUnregistered = async (serverId: string) => {
    try {
      const serverToRemove = registeredServers.find(s => s.id === serverId);
      await serverApi.unregisterServer(serverId);
      setRegisteredServers(prev => prev.filter(server => server.id !== serverId));
      success(`Server "${serverToRemove?.name || 'Server'}" unregistered successfully!`);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to unregister server';
      setError(errorMessage);
      showError(errorMessage);
    }
  };

  const handleRefresh = () => {
    loadServers();
  };

  if (loading && registeredServers.length === 0) {
    return (
      <ProtectedRoute>
        <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <div style={{ fontSize: '1.125rem', color: '#666' }}>Loading servers...</div>
          </div>
        </div>
      </ProtectedRoute>
    );
  }

  return (
    <ProtectedRoute>
      <ToastContainer />
      <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
      <header style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#333', marginBottom: '0.5rem' }}>
          Server Management
        </h1>
        <p style={{ fontSize: '1rem', color: '#666' }}>
          Manage your MCP servers and discover new ones
        </p>
      </header>

      <HealthCheck />

      {error && (
        <div style={{
          background: '#fef2f2',
          border: '1px solid #fecaca',
          borderRadius: '8px',
          padding: '1rem',
          marginBottom: '1.5rem',
          color: '#dc2626'
        }}>
          {error}
        </div>
      )}

      {/* Tab Navigation */}
      <div style={{ marginBottom: '2rem' }}>
        <div style={{
          borderBottom: '1px solid #e5e7eb',
          display: 'flex',
          gap: '2rem'
        }}>
          <button
            onClick={() => setActiveTab('registered')}
            style={{
              padding: '0.75rem 0',
              fontSize: '1rem',
              fontWeight: activeTab === 'registered' ? '600' : '400',
              color: activeTab === 'registered' ? '#3b82f6' : '#6b7280',
              background: 'none',
              border: 'none',
              borderBottom: activeTab === 'registered' ? '2px solid #3b82f6' : '2px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.2s'
            }}
          >
            Registered Servers ({registeredServers.length})
          </button>
          <button
            onClick={() => setActiveTab('available')}
            style={{
              padding: '0.75rem 0',
              fontSize: '1rem',
              fontWeight: activeTab === 'available' ? '600' : '400',
              color: activeTab === 'available' ? '#3b82f6' : '#6b7280',
              background: 'none',
              border: 'none',
              borderBottom: activeTab === 'available' ? '2px solid #3b82f6' : '2px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.2s'
            }}
          >
            Available Servers
          </button>
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'registered' && (
        <div>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '1.5rem'
          }}>
            <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333' }}>
              Your Registered Servers
            </h2>
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <button
                onClick={handleRefresh}
                style={{
                  background: '#f3f4f6',
                  color: '#374151',
                  padding: '0.5rem 1rem',
                  borderRadius: '6px',
                  border: 'none',
                  fontSize: '0.875rem',
                  cursor: 'pointer',
                  transition: 'background-color 0.2s'
                }}
                onMouseOver={(e) => e.currentTarget.style.background = '#e5e7eb'}
                onMouseOut={(e) => e.currentTarget.style.background = '#f3f4f6'}
              >
                Refresh
              </button>
              <button
                onClick={() => setShowRegisterModal(true)}
                style={{
                  background: '#3b82f6',
                  color: 'white',
                  padding: '0.5rem 1rem',
                  borderRadius: '6px',
                  border: 'none',
                  fontSize: '0.875rem',
                  cursor: 'pointer',
                  transition: 'background-color 0.2s'
                }}
                onMouseOver={(e) => e.currentTarget.style.background = '#2563eb'}
                onMouseOut={(e) => e.currentTarget.style.background = '#3b82f6'}
              >
                Register Server
              </button>
            </div>
          </div>

          <ServerTable
            servers={registeredServers}
            onUnregister={handleServerUnregistered}
            loading={loading}
          />
        </div>
      )}

      {activeTab === 'available' && (
        <div>
          <div style={{ marginBottom: '1.5rem' }}>
            <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333', marginBottom: '0.5rem' }}>
              Available MCP Servers
            </h2>
            <p style={{ color: '#666', fontSize: '0.875rem' }}>
              Discover and register new MCP servers from the community
            </p>
          </div>

          <AvailableServersList
            onRegister={handleServerRegistered}
          />
        </div>
      )}

      {/* Register Server Modal */}
      {showRegisterModal && (
        <RegisterServerModal
          onClose={() => setShowRegisterModal(false)}
          onRegister={handleServerRegistered}
        />
      )}
    </div>
    </ProtectedRoute>
  );
}
