'use client';

import { useState } from 'react';
import { A2AAgent } from '../../lib/api';

interface A2AAgentListProps {
  agents: A2AAgent[];
  onEdit: (agent: A2AAgent) => void;
  onDelete: (id: string) => void;
  onToggle: (id: string, active: boolean) => void;
  onTest: (id: string) => void;
  showInactive: boolean;
}

export function A2AAgentList({ agents, onEdit, onDelete, onToggle, onTest, showInactive }: A2AAgentListProps) {
  const [loadingActions, setLoadingActions] = useState<Record<string, boolean>>({});
  const [testResults, setTestResults] = useState<Record<string, { visible: boolean; result: any }>>({});

  const handleAction = async (agentId: string, action: () => Promise<void>) => {
    setLoadingActions(prev => ({ ...prev, [agentId]: true }));
    try {
      await action();
    } finally {
      setLoadingActions(prev => ({ ...prev, [agentId]: false }));
    }
  };

  const handleTest = async (agentId: string) => {
    setLoadingActions(prev => ({ ...prev, [`test-${agentId}`]: true }));
    try {
      await onTest(agentId);
      // Show test result area
      setTestResults(prev => ({
        ...prev,
        [agentId]: { visible: true, result: { success: true, message: 'Test completed successfully' } }
      }));
    } catch (error) {
      setTestResults(prev => ({
        ...prev,
        [agentId]: { visible: true, result: { success: false, message: 'Test failed' } }
      }));
    } finally {
      setLoadingActions(prev => ({ ...prev, [`test-${agentId}`]: false }));
    }
  };

  const getStatusBadge = (agent: A2AAgent) => {
    const statusColors = {
      active: { bg: '#dcfce7', text: '#166534' },
      inactive: { bg: '#fef2f2', text: '#dc2626' },
      healthy: { bg: '#dcfce7', text: '#166534' },
      unhealthy: { bg: '#fef3cd', text: '#92400e' },
    };

    const activeColor = agent.is_active ? statusColors.active : statusColors.inactive;
    const healthColor = agent.health_status === 'healthy' ? statusColors.healthy : statusColors.unhealthy;

    return { active: activeColor, health: healthColor };
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  if (agents.length === 0) {
    return (
      <div style={{
        background: 'white',
        borderRadius: '0.5rem',
        border: '1px solid #e5e7eb',
        padding: '3rem',
        textAlign: 'center',
        color: '#6b7280'
      }}>
        <p style={{ margin: 0, fontSize: '1.125rem' }}>
          {showInactive ? 'No A2A agents found' : 'No active A2A agents found'}
        </p>
        <p style={{ margin: '0.5rem 0 0 0', fontSize: '0.875rem' }}>
          Create your first A2A agent to get started.
        </p>
      </div>
    );
  }

  return (
    <div style={{
      background: 'white',
      borderRadius: '0.5rem',
      border: '1px solid #e5e7eb',
      overflow: 'hidden'
    }}>
      <div style={{ padding: '1.5rem 1.5rem 1rem 1.5rem' }}>
        <h3 style={{
          fontSize: '1.125rem',
          fontWeight: '600',
          color: '#111827',
          margin: '0 0 1rem 0'
        }}>
          Registered A2A Agents ({agents.length})
        </h3>
      </div>

      <div style={{ padding: '0 1.5rem 1.5rem 1.5rem' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {agents.map(agent => {
            const statusColors = getStatusBadge(agent);
            const isLoading = loadingActions[agent.id];
            const isTestLoading = loadingActions[`test-${agent.id}`];
            const testResult = testResults[agent.id];

            return (
              <div
                key={agent.id}
                style={{
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem',
                  padding: '1rem',
                  transition: 'all 0.2s'
                }}
              >
                <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
                  {/* Agent Info */}
                  <div style={{ flex: 1 }}>
                    <div style={{ marginBottom: '0.75rem' }}>
                      <h4 style={{
                        fontSize: '1.125rem',
                        fontWeight: '600',
                        color: '#111827',
                        margin: '0 0 0.25rem 0'
                      }}>
                        {agent.name}
                      </h4>
                      {agent.description && (
                        <p style={{
                          fontSize: '0.875rem',
                          color: '#6b7280',
                          margin: 0
                        }}>
                          {agent.description}
                        </p>
                      )}
                    </div>

                    {/* Status Badges */}
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', marginBottom: '0.75rem' }}>
                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '0.25rem 0.75rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        backgroundColor: statusColors.active.bg,
                        color: statusColors.active.text
                      }}>
                        {agent.is_active ? 'Active' : 'Inactive'}
                      </span>

                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '0.25rem 0.75rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        backgroundColor: statusColors.health.bg,
                        color: statusColors.health.text
                      }}>
                        {agent.health_status === 'healthy' ? 'Reachable' :
                         agent.health_status === 'unhealthy' ? 'Unreachable' : 'Unknown'}
                      </span>

                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '0.25rem 0.75rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        backgroundColor: '#eff6ff',
                        color: '#1d4ed8'
                      }}>
                        {agent.agent_type}
                      </span>

                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '0.25rem 0.75rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: '500',
                        backgroundColor: '#f3f4f6',
                        color: '#374151'
                      }}>
                        Auth: {agent.auth_type || 'None'}
                      </span>
                    </div>

                    {/* Agent Details */}
                    <div style={{
                      fontSize: '0.75rem',
                      color: '#6b7280',
                      lineHeight: '1.5',
                      marginBottom: '0.75rem'
                    }}>
                      <div>Endpoint: {agent.endpoint_url}</div>
                      <div>Created: {formatDate(agent.created_at)}</div>
                      {agent.last_health_check && (
                        <div>Last Health Check: {formatDate(agent.last_health_check)}</div>
                      )}
                      {agent.health_error && (
                        <div style={{ color: '#dc2626' }}>Health Error: {agent.health_error}</div>
                      )}
                    </div>

                    {/* Tags */}
                    {agent.tags && agent.tags.length > 0 && (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.25rem', marginBottom: '0.75rem' }}>
                        {agent.tags.map(tag => (
                          <span
                            key={tag}
                            style={{
                              display: 'inline-flex',
                              alignItems: 'center',
                              padding: '0.125rem 0.5rem',
                              borderRadius: '0.25rem',
                              fontSize: '0.75rem',
                              backgroundColor: '#f3f4f6',
                              color: '#374151'
                            }}
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}

                    {/* Test Result */}
                    {testResult && testResult.visible && (
                      <div style={{
                        padding: '0.75rem',
                        borderRadius: '0.375rem',
                        backgroundColor: testResult.result.success ? '#f0f9ff' : '#fef2f2',
                        border: `1px solid ${testResult.result.success ? '#bae6fd' : '#fecaca'}`,
                        marginBottom: '0.75rem'
                      }}>
                        <div style={{
                          fontSize: '0.875rem',
                          color: testResult.result.success ? '#0c4a6e' : '#991b1b',
                          fontWeight: '500'
                        }}>
                          Test {testResult.result.success ? 'Successful' : 'Failed'}
                        </div>
                        {testResult.result.message && (
                          <div style={{
                            fontSize: '0.75rem',
                            color: testResult.result.success ? '#0369a1' : '#b91c1c',
                            marginTop: '0.25rem'
                          }}>
                            {testResult.result.message}
                          </div>
                        )}
                      </div>
                    )}
                  </div>

                  {/* Action Buttons */}
                  <div style={{ display: 'flex', gap: '0.5rem', marginLeft: '1rem' }}>
                    <button
                      onClick={() => handleTest(agent.id)}
                      disabled={isTestLoading}
                      style={{
                        padding: '0.375rem 0.75rem',
                        fontSize: '0.875rem',
                        fontWeight: '500',
                        color: '#1d4ed8',
                        backgroundColor: '#eff6ff',
                        border: '1px solid #bfdbfe',
                        borderRadius: '0.375rem',
                        cursor: isTestLoading ? 'not-allowed' : 'pointer',
                        opacity: isTestLoading ? 0.5 : 1,
                        transition: 'all 0.2s'
                      }}
                      onMouseEnter={(e) => {
                        if (!isTestLoading) {
                          e.currentTarget.style.backgroundColor = '#dbeafe';
                          e.currentTarget.style.borderColor = '#93c5fd';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (!isTestLoading) {
                          e.currentTarget.style.backgroundColor = '#eff6ff';
                          e.currentTarget.style.borderColor = '#bfdbfe';
                        }
                      }}
                    >
                      {isTestLoading ? 'Testing...' : 'Test'}
                    </button>

                    <button
                      onClick={() => onEdit(agent)}
                      style={{
                        padding: '0.375rem 0.75rem',
                        fontSize: '0.875rem',
                        fontWeight: '500',
                        color: '#374151',
                        backgroundColor: '#f9fafb',
                        border: '1px solid #d1d5db',
                        borderRadius: '0.375rem',
                        cursor: 'pointer',
                        transition: 'all 0.2s'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.backgroundColor = '#f3f4f6';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.backgroundColor = '#f9fafb';
                      }}
                    >
                      Edit
                    </button>

                    <button
                      onClick={() => handleAction(agent.id, async () => onToggle(agent.id, !agent.is_active))}
                      disabled={isLoading}
                      style={{
                        padding: '0.375rem 0.75rem',
                        fontSize: '0.875rem',
                        fontWeight: '500',
                        color: agent.is_active ? '#dc2626' : '#16a34a',
                        backgroundColor: agent.is_active ? '#fef2f2' : '#f0f9f0',
                        border: agent.is_active ? '1px solid #fecaca' : '1px solid #bbf7d0',
                        borderRadius: '0.375rem',
                        cursor: isLoading ? 'not-allowed' : 'pointer',
                        opacity: isLoading ? 0.5 : 1,
                        transition: 'all 0.2s'
                      }}
                      onMouseEnter={(e) => {
                        if (!isLoading) {
                          if (agent.is_active) {
                            e.currentTarget.style.backgroundColor = '#fee2e2';
                            e.currentTarget.style.borderColor = '#fca5a5';
                          } else {
                            e.currentTarget.style.backgroundColor = '#dcfce7';
                            e.currentTarget.style.borderColor = '#86efac';
                          }
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (!isLoading) {
                          if (agent.is_active) {
                            e.currentTarget.style.backgroundColor = '#fef2f2';
                            e.currentTarget.style.borderColor = '#fecaca';
                          } else {
                            e.currentTarget.style.backgroundColor = '#f0f9f0';
                            e.currentTarget.style.borderColor = '#bbf7d0';
                          }
                        }
                      }}
                    >
                      {isLoading ? 'Loading...' : (agent.is_active ? 'Deactivate' : 'Activate')}
                    </button>

                    <button
                      onClick={() => handleAction(agent.id, async () => onDelete(agent.id))}
                      disabled={isLoading}
                      style={{
                        padding: '0.375rem 0.75rem',
                        fontSize: '0.875rem',
                        fontWeight: '500',
                        color: '#dc2626',
                        backgroundColor: '#fef2f2',
                        border: '1px solid #fecaca',
                        borderRadius: '0.375rem',
                        cursor: isLoading ? 'not-allowed' : 'pointer',
                        opacity: isLoading ? 0.5 : 1,
                        transition: 'all 0.2s'
                      }}
                      onMouseEnter={(e) => {
                        if (!isLoading) {
                          e.currentTarget.style.backgroundColor = '#fee2e2';
                          e.currentTarget.style.borderColor = '#fca5a5';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (!isLoading) {
                          e.currentTarget.style.backgroundColor = '#fef2f2';
                          e.currentTarget.style.borderColor = '#fecaca';
                        }
                      }}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
