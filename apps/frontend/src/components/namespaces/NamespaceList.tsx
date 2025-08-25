'use client';

import { useState } from 'react';
import Link from 'next/link';
import { EditNamespaceModal } from './EditNamespaceModal';

interface Namespace {
  id: string;
  name: string;
  description?: string;
  servers?: string[];
  server_count?: number;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  metadata?: Record<string, any>;
}

interface NamespaceListProps {
  namespaces: Namespace[];
  onDelete: (id: string) => void;
  onRefresh: () => void;
}

export function NamespaceList({ namespaces, onDelete, onRefresh }: NamespaceListProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [editingNamespace, setEditingNamespace] = useState<Namespace | null>(null);

  const filteredNamespaces = (namespaces || []).filter(ns =>
    ns.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    ns.description?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div>
      {/* Search Bar */}
      <div style={{ marginBottom: '1.5rem' }}>
        <input
          type="text"
          placeholder="Search namespaces..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          style={{
            width: '100%',
            maxWidth: '24rem',
            padding: '0.625rem 1rem',
            border: '1px solid #e5e7eb',
            borderRadius: '0.375rem',
            fontSize: '0.875rem'
          }}
        />
      </div>

      {/* Namespaces Table */}
      <div style={{
        backgroundColor: 'white',
        border: '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        overflow: 'hidden'
      }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ backgroundColor: '#f9fafb', borderBottom: '1px solid #e5e7eb' }}>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Name
              </th>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Description
              </th>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Servers
              </th>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Status
              </th>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Created
              </th>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'center', fontWeight: '600', color: '#374151', fontSize: '0.875rem' }}>
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {filteredNamespaces.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ padding: '2rem', textAlign: 'center', color: '#6b7280' }}>
                  {searchTerm ? 'No namespaces found matching your search' : 'No namespaces created yet'}
                </td>
              </tr>
            ) : (
              filteredNamespaces.map((namespace) => (
                <tr key={namespace.id} style={{ borderBottom: '1px solid #e5e7eb' }}>
                  <td style={{ padding: '1rem' }}>
                    <Link
                      href={`/namespaces/${namespace.id}`}
                      style={{
                        color: '#3b82f6',
                        textDecoration: 'none',
                        fontWeight: '500',
                        fontSize: '0.875rem'
                      }}
                      onMouseEnter={(e) => e.currentTarget.style.textDecoration = 'underline'}
                      onMouseLeave={(e) => e.currentTarget.style.textDecoration = 'none'}
                    >
                      {namespace.name}
                    </Link>
                  </td>
                  <td style={{ padding: '1rem', color: '#6b7280', fontSize: '0.875rem' }}>
                    {namespace.description || '-'}
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'center' }}>
                    <span style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      padding: '0.25rem 0.75rem',
                      backgroundColor: '#dbeafe',
                      color: '#1e40af',
                      borderRadius: '9999px',
                      fontSize: '0.75rem',
                      fontWeight: '500'
                    }}>
                      {namespace.server_count ?? namespace.servers?.length ?? 0} servers
                    </span>
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'center' }}>
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
                  </td>
                  <td style={{ padding: '1rem', color: '#6b7280', fontSize: '0.875rem' }}>
                    {formatDate(namespace.created_at)}
                  </td>
                  <td style={{ padding: '1rem' }}>
                    <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                      <button
                        onClick={() => setEditingNamespace(namespace)}
                        style={{
                          padding: '0.375rem 0.75rem',
                          backgroundColor: 'white',
                          color: '#374151',
                          border: '1px solid #e5e7eb',
                          borderRadius: '0.375rem',
                          fontSize: '0.75rem',
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
                        onClick={() => onDelete(namespace.id)}
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
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Summary */}
      <div style={{
        marginTop: '1rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        color: '#6b7280',
        fontSize: '0.875rem'
      }}>
        <span>
          Showing {filteredNamespaces.length} of {namespaces?.length || 0} namespaces
        </span>
        <button
          onClick={onRefresh}
          style={{
            padding: '0.375rem 0.75rem',
            backgroundColor: 'white',
            color: '#374151',
            border: '1px solid #e5e7eb',
            borderRadius: '0.375rem',
            fontSize: '0.75rem',
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
          Refresh
        </button>
      </div>

      {/* Edit Modal */}
      {editingNamespace && (
        <EditNamespaceModal
          namespace={editingNamespace}
          onClose={() => setEditingNamespace(null)}
          onUpdate={() => {
            setEditingNamespace(null);
            onRefresh();
          }}
        />
      )}
    </div>
  );
}