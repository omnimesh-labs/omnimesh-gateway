'use client';

import { useState, useEffect } from 'react';
import { namespaceApi } from '../../lib/api';

interface NamespaceTool {
  tool_id: string;
  tool_name: string;
  prefixed_name: string;
  server_id: string;
  server_name: string;
  status: 'ACTIVE' | 'INACTIVE';
}

interface NamespaceToolsManagerProps {
  namespaceId: string;
}

export function NamespaceToolsManager({ namespaceId }: NamespaceToolsManagerProps) {
  const [tools, setTools] = useState<NamespaceTool[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  const fetchTools = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await namespaceApi.getNamespaceTools(namespaceId);
      setTools(data);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch tools';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTools();
  }, [namespaceId]);

  const handleToggleStatus = async (toolId: string, currentStatus: string) => {
    const newStatus = currentStatus === 'ACTIVE' ? 'INACTIVE' : 'ACTIVE';
    try {
      await namespaceApi.updateToolStatus(namespaceId, toolId, newStatus);
      // Update local state optimistically
      setTools(prev => prev.map(tool =>
        tool.tool_id === toolId ? { ...tool, status: newStatus as 'ACTIVE' | 'INACTIVE' } : tool
      ));
    } catch (error) {
      console.error('Failed to update tool status:', error);
      // Revert on error
      fetchTools();
    }
  };

  const filteredTools = (tools || []).filter(tool =>
    tool.tool_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    tool.prefixed_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    tool.server_name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const groupedTools = filteredTools.reduce((acc, tool) => {
    if (!acc[tool.server_name]) {
      acc[tool.server_name] = [];
    }
    acc[tool.server_name].push(tool);
    return acc;
  }, {} as Record<string, NamespaceTool[]>);

  return (
    <div style={{
      backgroundColor: 'white',
      border: '1px solid #e5e7eb',
      borderRadius: '0.5rem',
      padding: '1.5rem'
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
          Tools Management
        </h2>
        <button
          onClick={fetchTools}
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
          Refresh Tools
        </button>
      </div>

      {/* Search Bar */}
      <div style={{ marginBottom: '1rem' }}>
        <input
          type="text"
          placeholder="Search tools..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          style={{
            width: '100%',
            maxWidth: '24rem',
            padding: '0.5rem 1rem',
            border: '1px solid #e5e7eb',
            borderRadius: '0.375rem',
            fontSize: '0.875rem'
          }}
        />
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '2rem', color: '#6b7280' }}>
          Loading tools...
        </div>
      ) : error ? (
        <div style={{
          backgroundColor: '#fef2f2',
          color: '#dc2626',
          padding: '0.75rem',
          borderRadius: '0.375rem',
          fontSize: '0.875rem'
        }}>
          {error}
        </div>
      ) : Object.keys(groupedTools).length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem', color: '#6b7280' }}>
          {searchTerm ? 'No tools found matching your search' : 'No tools available in this namespace'}
        </div>
      ) : (
        <div style={{ maxHeight: '400px', overflow: 'auto' }}>
          {Object.entries(groupedTools).map(([serverName, serverTools]) => (
            <div key={serverName} style={{ marginBottom: '1.5rem' }}>
              <h3 style={{
                fontSize: '0.875rem',
                fontWeight: '600',
                color: '#374151',
                marginBottom: '0.5rem',
                padding: '0.5rem',
                backgroundColor: '#f9fafb',
                borderRadius: '0.25rem'
              }}>
                {serverName} ({serverTools.length} tools)
              </h3>
              <div style={{ paddingLeft: '1rem' }}>
                {serverTools.map(tool => (
                  <div
                    key={tool.tool_id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '0.5rem',
                      borderBottom: '1px solid #f3f4f6'
                    }}
                  >
                    <div style={{ flex: 1 }}>
                      <div style={{
                        fontWeight: '500',
                        fontSize: '0.875rem',
                        color: '#111827'
                      }}>
                        {tool.tool_name}
                      </div>
                      <div style={{
                        fontSize: '0.75rem',
                        color: '#6b7280',
                        fontFamily: 'monospace'
                      }}>
                        {tool.prefixed_name}
                      </div>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '0.125rem 0.5rem',
                        backgroundColor: tool.status === 'ACTIVE' ? '#d1fae5' : '#fee2e2',
                        color: tool.status === 'ACTIVE' ? '#065f46' : '#991b1b',
                        borderRadius: '9999px',
                        fontSize: '0.625rem',
                        fontWeight: '500'
                      }}>
                        {tool.status}
                      </span>
                      <button
                        onClick={() => handleToggleStatus(tool.tool_id, tool.status)}
                        style={{
                          padding: '0.25rem 0.5rem',
                          backgroundColor: 'white',
                          color: '#374151',
                          border: '1px solid #e5e7eb',
                          borderRadius: '0.25rem',
                          fontSize: '0.625rem',
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
                        {tool.status === 'ACTIVE' ? 'Disable' : 'Enable'}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Summary */}
      <div style={{
        marginTop: '1rem',
        paddingTop: '1rem',
        borderTop: '1px solid #e5e7eb',
        display: 'flex',
        justifyContent: 'space-between',
        fontSize: '0.75rem',
        color: '#6b7280'
      }}>
        <span>
          Total: {filteredTools.length} tools from {Object.keys(groupedTools).length} servers
        </span>
        <span>
          Active: {filteredTools.filter(t => t.status === 'ACTIVE').length} |
          Inactive: {filteredTools.filter(t => t.status === 'INACTIVE').length}
        </span>
      </div>
    </div>
  );
}
