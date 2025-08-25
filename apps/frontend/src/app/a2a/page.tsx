'use client';

import { useState, useEffect, useCallback } from 'react';
import { a2aApi, A2AAgent, A2AAgentSpec, A2AStats } from '../../lib/api';
import { Toast } from '../../components/Toast';
import { A2AAgentForm } from '../../components/a2a/A2AAgentForm';
import { A2AAgentList } from '../../components/a2a/A2AAgentList';

export default function A2APage() {
  const [agents, setAgents] = useState<A2AAgent[]>([]);
  const [stats, setStats] = useState<A2AStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingAgent, setEditingAgent] = useState<A2AAgent | null>(null);
  const [showInactive, setShowInactive] = useState(false);
  const [tagFilter, setTagFilter] = useState('');
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [toast, setToast] = useState<{
    message: string;
    type: 'success' | 'error';
    show: boolean;
  } | null>(null);

  const showToast = useCallback((message: string, type: 'success' | 'error') => {
    setToast({ message, type, show: true });
  }, []);

  const hideToast = useCallback(() => {
    setToast(null);
  }, []);

  const loadAgents = useCallback(async () => {
    try {
      setLoading(true);
      const [agentsData, statsData] = await Promise.all([
        a2aApi.listAgents(),
        a2aApi.getStats()
      ]);

      setAgents(agentsData);
      setStats(statsData);

      // Extract unique tags from all agents
      const tags = new Set<string>();
      agentsData.forEach(agent => {
        if (agent.tags) {
          agent.tags.forEach(tag => tags.add(tag));
        }
      });
      setAvailableTags(Array.from(tags).sort());
    } catch (error) {
      console.error('Failed to load A2A agents:', error);
      showToast('Failed to load A2A agents', 'error');
    } finally {
      setLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    loadAgents();
  }, [loadAgents]);

  const handleCreateAgent = async (agentData: A2AAgentSpec) => {
    try {
      await a2aApi.createAgent(agentData);
      showToast('A2A agent created successfully', 'success');
      setShowForm(false);
      loadAgents();
    } catch (error) {
      console.error('Failed to create agent:', error);
      showToast('Failed to create A2A agent', 'error');
    }
  };

  const handleUpdateAgent = async (id: string, agentData: Partial<A2AAgentSpec>) => {
    try {
      await a2aApi.updateAgent(id, agentData);
      showToast('A2A agent updated successfully', 'success');
      setEditingAgent(null);
      loadAgents();
    } catch (error) {
      console.error('Failed to update agent:', error);
      showToast('Failed to update A2A agent', 'error');
    }
  };

  const handleDeleteAgent = async (id: string) => {
    if (!confirm('Are you sure you want to delete this A2A agent?')) {
      return;
    }

    try {
      await a2aApi.deleteAgent(id);
      showToast('A2A agent deleted successfully', 'success');
      loadAgents();
    } catch (error) {
      console.error('Failed to delete agent:', error);
      showToast('Failed to delete A2A agent', 'error');
    }
  };

  const handleToggleAgent = async (id: string, active: boolean) => {
    try {
      await a2aApi.toggleAgent(id, active);
      showToast(`A2A agent ${active ? 'activated' : 'deactivated'} successfully`, 'success');
      loadAgents();
    } catch (error) {
      console.error('Failed to toggle agent:', error);
      showToast('Failed to toggle A2A agent status', 'error');
    }
  };

  const handleTestAgent = async (id: string) => {
    try {
      const result = await a2aApi.testAgent(id, {
        message: 'Hello, this is a test message to verify agent connectivity.',
        max_tokens: 100
      });

      if (result.success) {
        showToast(`Test successful! Response: ${result.content?.substring(0, 100)}...`, 'success');
      } else {
        showToast(`Test failed: ${result.error}`, 'error');
      }
    } catch (error) {
      console.error('Failed to test agent:', error);
      showToast('Failed to test A2A agent', 'error');
    }
  };

  const filteredAgents = agents.filter(agent => {
    if (!showInactive && !agent.is_active) return false;

    if (tagFilter) {
      const filterTags = tagFilter.split(',').map(t => t.trim().toLowerCase()).filter(t => t);
      if (filterTags.length > 0) {
        const agentTags = (agent.tags || []).map(t => t.toLowerCase());
        if (!filterTags.some(filterTag => agentTags.includes(filterTag))) return false;
      }
    }

    return true;
  });

  const clearTagFilter = () => {
    setTagFilter('');
  };

  const addTagToFilter = (tag: string) => {
    const currentTags = tagFilter.split(',').map(t => t.trim()).filter(t => t);
    if (!currentTags.includes(tag)) {
      const newFilter = [...currentTags, tag].join(', ');
      setTagFilter(newFilter);
    }
  };

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '200px',
        color: '#6b7280'
      }}>
        Loading A2A agents...
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '0 2rem' }}>
      {toast && toast.show && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={hideToast}
        />
      )}

      {/* Header */}
      <div style={{ marginBottom: '2rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
          <div>
            <h1 style={{
              fontSize: '2rem',
              fontWeight: 'bold',
              color: '#111827',
              margin: '0 0 0.5rem 0'
            }}>
              A2A Agents Catalog
            </h1>
            <p style={{
              fontSize: '0.875rem',
              color: '#6b7280',
              margin: 0
            }}>
              Manage Agent-to-Agent compatible agents that can be integrated as tools
            </p>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <label style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151'
            }}>
              <input
                type="checkbox"
                checked={showInactive}
                onChange={(e) => setShowInactive(e.target.checked)}
                style={{ margin: 0 }}
              />
              Show Inactive
            </label>
          </div>
        </div>

        {/* Stats */}
        {stats && (
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: '1rem',
            marginBottom: '2rem'
          }}>
            <div style={{
              background: 'white',
              padding: '1rem',
              borderRadius: '0.5rem',
              border: '1px solid #e5e7eb'
            }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#111827' }}>
                {stats.total}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                Total Agents
              </div>
            </div>
            <div style={{
              background: 'white',
              padding: '1rem',
              borderRadius: '0.5rem',
              border: '1px solid #e5e7eb'
            }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#10b981' }}>
                {stats.active}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                Active Agents
              </div>
            </div>
            <div style={{
              background: 'white',
              padding: '1rem',
              borderRadius: '0.5rem',
              border: '1px solid #e5e7eb'
            }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#3b82f6' }}>
                {Object.keys(stats.by_type).length}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                Agent Types
              </div>
            </div>
            <div style={{
              background: 'white',
              padding: '1rem',
              borderRadius: '0.5rem',
              border: '1px solid #e5e7eb'
            }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#10b981' }}>
                {stats.by_health.healthy || 0}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                Healthy
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Filter Section */}
      <div style={{
        background: 'white',
        padding: '1rem',
        borderRadius: '0.5rem',
        border: '1px solid #e5e7eb',
        marginBottom: '2rem'
      }}>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem', alignItems: 'center' }}>
          <div style={{ flex: '1', minWidth: '250px' }}>
            <label style={{
              display: 'block',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              marginBottom: '0.25rem'
            }}>
              Filter by Tags:
            </label>
            <input
              type="text"
              value={tagFilter}
              onChange={(e) => setTagFilter(e.target.value)}
              placeholder="e.g., ai,assistant (comma-separated)"
              style={{
                width: '100%',
                padding: '0.5rem 0.75rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                outline: 'none'
              }}
              onFocus={(e) => {
                e.target.style.borderColor = '#3b82f6';
                e.target.style.boxShadow = '0 0 0 1px #3b82f6';
              }}
              onBlur={(e) => {
                e.target.style.borderColor = '#d1d5db';
                e.target.style.boxShadow = 'none';
              }}
            />
          </div>

          {availableTags.length > 0 && (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
              {availableTags.slice(0, 8).map(tag => (
                <button
                  key={tag}
                  onClick={() => addTagToFilter(tag)}
                  style={{
                    padding: '0.25rem 0.5rem',
                    fontSize: '0.75rem',
                    fontWeight: '500',
                    color: '#3b82f6',
                    background: '#eff6ff',
                    border: '1px solid #bfdbfe',
                    borderRadius: '0.25rem',
                    cursor: 'pointer',
                    transition: 'all 0.2s'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = '#dbeafe';
                    e.currentTarget.style.borderColor = '#93c5fd';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = '#eff6ff';
                    e.currentTarget.style.borderColor = '#bfdbfe';
                  }}
                >
                  {tag}
                </button>
              ))}
            </div>
          )}

          <button
            onClick={clearTagFilter}
            style={{
              padding: '0.5rem 0.75rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              background: '#f9fafb',
              border: '1px solid #d1d5db',
              borderRadius: '0.375rem',
              cursor: 'pointer',
              transition: 'all 0.2s'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = '#f3f4f6';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = '#f9fafb';
            }}
          >
            Clear Filter
          </button>
        </div>
      </div>

      {/* Add Agent Button */}
      <div style={{ marginBottom: '2rem' }}>
        <button
          onClick={() => setShowForm(true)}
          style={{
            padding: '0.75rem 1.5rem',
            fontSize: '0.875rem',
            fontWeight: '500',
            color: 'white',
            background: '#3b82f6',
            border: 'none',
            borderRadius: '0.375rem',
            cursor: 'pointer',
            transition: 'background-color 0.2s'
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.backgroundColor = '#2563eb';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.backgroundColor = '#3b82f6';
          }}
        >
          Add New A2A Agent
        </button>
      </div>

      {/* Agent Form */}
      {(showForm || editingAgent) && (
        <A2AAgentForm
          agent={editingAgent}
          onSubmit={editingAgent ?
            (data: A2AAgentSpec) => handleUpdateAgent(editingAgent.id, data) :
            handleCreateAgent
          }
          onCancel={() => {
            setShowForm(false);
            setEditingAgent(null);
          }}
        />
      )}

      {/* Agents List */}
      <A2AAgentList
        agents={filteredAgents}
        onEdit={setEditingAgent}
        onDelete={handleDeleteAgent}
        onToggle={handleToggleAgent}
        onTest={handleTestAgent}
        showInactive={showInactive}
      />

      {filteredAgents.length === 0 && (
        <div style={{
          textAlign: 'center',
          padding: '4rem 2rem',
          color: '#6b7280'
        }}>
          {agents.length === 0 ? (
            <p>No A2A agents registered yet.</p>
          ) : (
            <p>No agents match the current filters.</p>
          )}
        </div>
      )}
    </div>
  );
}
