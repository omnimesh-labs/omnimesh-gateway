'use client';

import { useState, useEffect } from 'react';
import { A2AAgent, A2AAgentSpec } from '../../lib/api';

interface A2AAgentFormProps {
  agent?: A2AAgent | null;
  onSubmit: (data: A2AAgentSpec) => Promise<void>;
  onCancel: () => void;
}

export function A2AAgentForm({ agent, onSubmit, onCancel }: A2AAgentFormProps) {
  const [formData, setFormData] = useState<A2AAgentSpec>({
    name: '',
    description: '',
    endpoint_url: '',
    agent_type: 'generic',
    auth_type: 'none',
    auth_value: '',
    is_active: true,
    tags: [],
    protocol_version: '1.0',
  });
  const [tagsInput, setTagsInput] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (agent) {
      setFormData({
        name: agent.name,
        description: agent.description || '',
        endpoint_url: agent.endpoint_url,
        agent_type: agent.agent_type,
        auth_type: agent.auth_type,
        auth_value: '', // Never pre-fill auth value for security
        is_active: agent.is_active,
        tags: agent.tags || [],
        protocol_version: agent.protocol_version || '1.0',
        capabilities: agent.capabilities,
        config: agent.config,
        metadata: agent.metadata,
      });
      setTagsInput((agent.tags || []).join(', '));
    }
  }, [agent]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const tags = tagsInput
        .split(',')
        .map(tag => tag.trim())
        .filter(tag => tag.length > 0);

      const submitData: A2AAgentSpec = {
        ...formData,
        tags: tags.length > 0 ? tags : undefined,
      };

      // Only include auth_value if it's not empty
      if (!submitData.auth_value?.trim()) {
        delete submitData.auth_value;
      }

      await onSubmit(submitData);
    } catch (error) {
      console.error('Form submission error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (field: keyof A2AAgentSpec, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
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
        boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
        width: '90%',
        maxWidth: '600px',
        maxHeight: '90vh',
        overflow: 'auto'
      }}>
        <div style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <h2 style={{ 
              fontSize: '1.5rem', 
              fontWeight: 'bold', 
              color: '#111827',
              margin: 0
            }}>
              {agent ? 'Edit A2A Agent' : 'Add New A2A Agent'}
            </h2>
            <button
              onClick={onCancel}
              style={{
                width: '2rem',
                height: '2rem',
                borderRadius: '50%',
                border: 'none',
                background: '#f3f4f6',
                color: '#6b7280',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '1.25rem',
                transition: 'all 0.2s'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = '#e5e7eb';
                e.currentTarget.style.color = '#374151';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = '#f3f4f6';
                e.currentTarget.style.color = '#6b7280';
              }}
            >
              ×
            </button>
          </div>

          <form onSubmit={handleSubmit}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              {/* Agent Name */}
              <div>
                <label style={{ 
                  display: 'block', 
                  fontSize: '0.875rem', 
                  fontWeight: '500', 
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Agent Name *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  required
                  placeholder="my-assistant-agent"
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

              {/* Endpoint URL */}
              <div>
                <label style={{ 
                  display: 'block', 
                  fontSize: '0.875rem', 
                  fontWeight: '500', 
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Endpoint URL *
                </label>
                <input
                  type="url"
                  value={formData.endpoint_url}
                  onChange={(e) => handleInputChange('endpoint_url', e.target.value)}
                  required
                  placeholder="https://api.example.com/agent"
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
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              {/* Agent Type */}
              <div>
                <label style={{ 
                  display: 'block', 
                  fontSize: '0.875rem', 
                  fontWeight: '500', 
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Agent Type
                </label>
                <select
                  value={formData.agent_type}
                  onChange={(e) => handleInputChange('agent_type', e.target.value as A2AAgentSpec['agent_type'])}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    backgroundColor: 'white'
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = '#3b82f6';
                    e.target.style.boxShadow = '0 0 0 1px #3b82f6';
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#d1d5db';
                    e.target.style.boxShadow = 'none';
                  }}
                >
                  <option value="generic">Generic</option>
                  <option value="openai">OpenAI</option>
                  <option value="anthropic">Anthropic</option>
                  <option value="custom">Custom</option>
                </select>
              </div>

              {/* Authentication Type */}
              <div>
                <label style={{ 
                  display: 'block', 
                  fontSize: '0.875rem', 
                  fontWeight: '500', 
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Authentication Type
                </label>
                <select
                  value={formData.auth_type}
                  onChange={(e) => handleInputChange('auth_type', e.target.value as A2AAgentSpec['auth_type'])}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    backgroundColor: 'white'
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = '#3b82f6';
                    e.target.style.boxShadow = '0 0 0 1px #3b82f6';
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#d1d5db';
                    e.target.style.boxShadow = 'none';
                  }}
                >
                  <option value="none">None</option>
                  <option value="api_key">API Key</option>
                  <option value="bearer">Bearer Token</option>
                  <option value="oauth">OAuth</option>
                </select>
              </div>
            </div>

            {/* Description */}
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ 
                display: 'block', 
                fontSize: '0.875rem', 
                fontWeight: '500', 
                color: '#374151',
                marginBottom: '0.25rem'
              }}>
                Description
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                rows={3}
                placeholder="Description of the agent's capabilities"
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem',
                  outline: 'none',
                  resize: 'vertical'
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

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              {/* Authentication Value */}
              <div>
                <label style={{ 
                  display: 'block', 
                  fontSize: '0.875rem', 
                  fontWeight: '500', 
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Authentication Value
                  {formData.auth_type !== 'none' && (
                    <span style={{ color: '#ef4444', fontSize: '0.75rem', marginLeft: '0.25rem' }}>
                      (leave empty to keep existing)
                    </span>
                  )}
                </label>
                <input
                  type="password"
                  value={formData.auth_value || ''}
                  onChange={(e) => handleInputChange('auth_value', e.target.value)}
                  placeholder="API key, token, etc."
                  disabled={formData.auth_type === 'none'}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    backgroundColor: formData.auth_type === 'none' ? '#f9fafb' : 'white',
                    opacity: formData.auth_type === 'none' ? 0.5 : 1
                  }}
                  onFocus={(e) => {
                    if (formData.auth_type !== 'none') {
                      e.target.style.borderColor = '#3b82f6';
                      e.target.style.boxShadow = '0 0 0 1px #3b82f6';
                    }
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#d1d5db';
                    e.target.style.boxShadow = 'none';
                  }}
                />
              </div>

              {/* Protocol Version */}
              <div>
                <label style={{ 
                  display: 'block', 
                  fontSize: '0.875rem', 
                  fontWeight: '500', 
                  color: '#374151',
                  marginBottom: '0.25rem'
                }}>
                  Protocol Version
                </label>
                <input
                  type="text"
                  value={formData.protocol_version || '1.0'}
                  onChange={(e) => handleInputChange('protocol_version', e.target.value)}
                  placeholder="1.0"
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
            </div>

            {/* Tags */}
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ 
                display: 'block', 
                fontSize: '0.875rem', 
                fontWeight: '500', 
                color: '#374151',
                marginBottom: '0.25rem'
              }}>
                Tags
              </label>
              <input
                type="text"
                value={tagsInput}
                onChange={(e) => setTagsInput(e.target.value)}
                placeholder="ai,assistant,production (comma-separated)"
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

            {/* Active Status */}
            <div style={{ marginBottom: '2rem' }}>
              <label style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: '0.5rem',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                cursor: 'pointer'
              }}>
                <input
                  type="checkbox"
                  checked={formData.is_active}
                  onChange={(e) => handleInputChange('is_active', e.target.checked)}
                  style={{ margin: 0 }}
                />
                Active (enable this agent immediately)
              </label>
            </div>

            {/* Buttons */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem' }}>
              <button
                type="button"
                onClick={onCancel}
                disabled={loading}
                style={{
                  padding: '0.5rem 1rem',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  background: '#f9fafb',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  cursor: loading ? 'not-allowed' : 'pointer',
                  opacity: loading ? 0.5 : 1,
                  transition: 'all 0.2s'
                }}
                onMouseEnter={(e) => {
                  if (!loading) {
                    e.currentTarget.style.background = '#f3f4f6';
                  }
                }}
                onMouseLeave={(e) => {
                  if (!loading) {
                    e.currentTarget.style.background = '#f9fafb';
                  }
                }}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading || !formData.name || !formData.endpoint_url}
                style={{
                  padding: '0.5rem 1rem',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: 'white',
                  background: loading || !formData.name || !formData.endpoint_url ? '#9ca3af' : '#3b82f6',
                  border: 'none',
                  borderRadius: '0.375rem',
                  cursor: loading || !formData.name || !formData.endpoint_url ? 'not-allowed' : 'pointer',
                  transition: 'background-color 0.2s'
                }}
                onMouseEnter={(e) => {
                  if (!loading && formData.name && formData.endpoint_url) {
                    e.currentTarget.style.backgroundColor = '#2563eb';
                  }
                }}
                onMouseLeave={(e) => {
                  if (!loading && formData.name && formData.endpoint_url) {
                    e.currentTarget.style.backgroundColor = '#3b82f6';
                  }
                }}
              >
                {loading ? 'Saving...' : (agent ? 'Update Agent' : 'Create Agent')}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}