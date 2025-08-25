'use client';

import { useState, useEffect } from 'react';
import { NamespaceList } from '../../components/namespaces/NamespaceList';
import { CreateNamespaceModal } from '../../components/namespaces/CreateNamespaceModal';
import { Toast } from '../../components/Toast';
import { namespaceApi } from '../../lib/api';

interface Namespace {
  id: string;
  name: string;
  description?: string;
  servers?: string[];
  created_at: string;
  updated_at: string;
  is_active: boolean;
  metadata?: Record<string, any>;
}

export default function NamespacesPage() {
  const [namespaces, setNamespaces] = useState<Namespace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const fetchNamespaces = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await namespaceApi.listNamespaces();
      setNamespaces(data);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch namespaces';
      setError(errorMessage);
      setToast({ message: errorMessage, type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNamespaces();
  }, []);

  const handleCreateNamespace = async (namespaceData: {
    name: string;
    description: string;
    server_ids: string[];
  }) => {
    try {
      await namespaceApi.createNamespace(namespaceData);
      setToast({ message: 'Namespace created successfully', type: 'success' });
      setShowCreateModal(false);
      fetchNamespaces();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to create namespace';
      setToast({ message: errorMessage, type: 'error' });
    }
  };

  const handleDeleteNamespace = async (id: string) => {
    if (!confirm('Are you sure you want to delete this namespace?')) {
      return;
    }

    try {
      await namespaceApi.deleteNamespace(id);
      setToast({ message: 'Namespace deleted successfully', type: 'success' });
      fetchNamespaces();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete namespace';
      setToast({ message: errorMessage, type: 'error' });
    }
  };

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '2rem' }}>
      {/* Header */}
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: '2rem'
      }}>
        <div>
          <h1 style={{ 
            fontSize: '1.875rem', 
            fontWeight: 'bold', 
            color: '#111827',
            marginBottom: '0.5rem'
          }}>
            Namespaces
          </h1>
          <p style={{ color: '#6b7280' }}>
            Group and organize your MCP servers into logical namespaces
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          style={{
            backgroundColor: '#3b82f6',
            color: 'white',
            padding: '0.625rem 1.25rem',
            borderRadius: '0.375rem',
            border: 'none',
            cursor: 'pointer',
            fontWeight: '500',
            fontSize: '0.875rem',
            transition: 'background-color 0.2s'
          }}
          onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#2563eb'}
          onMouseLeave={(e) => e.currentTarget.style.backgroundColor = '#3b82f6'}
        >
          + Create Namespace
        </button>
      </div>

      {/* Content */}
      {loading ? (
        <div style={{ 
          textAlign: 'center', 
          padding: '3rem',
          color: '#6b7280'
        }}>
          Loading namespaces...
        </div>
      ) : error ? (
        <div style={{
          backgroundColor: '#fef2f2',
          color: '#dc2626',
          padding: '1rem',
          borderRadius: '0.375rem',
          marginBottom: '1rem'
        }}>
          {error}
        </div>
      ) : (
        <NamespaceList
          namespaces={namespaces}
          onDelete={handleDeleteNamespace}
          onRefresh={fetchNamespaces}
        />
      )}

      {/* Create Namespace Modal */}
      {showCreateModal && (
        <CreateNamespaceModal
          onClose={() => setShowCreateModal(false)}
          onCreate={handleCreateNamespace}
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