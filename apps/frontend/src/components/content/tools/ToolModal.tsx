'use client';

import { useState, useEffect } from 'react';
import { Tool, CreateToolRequest, UpdateToolRequest } from '@/lib/api';

interface ToolModalProps {
  tool?: Tool;
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: CreateToolRequest | UpdateToolRequest) => Promise<void>;
}

const TOOL_CATEGORIES = [
  { value: 'general', label: 'General' },
  { value: 'data', label: 'Data Processing' },
  { value: 'file', label: 'File Operations' },
  { value: 'web', label: 'Web/API' },
  { value: 'system', label: 'System' },
  { value: 'ai', label: 'AI/ML' },
  { value: 'dev', label: 'Development' },
  { value: 'custom', label: 'Custom' }
];

const IMPLEMENTATION_TYPES = [
  { value: 'internal', label: 'Internal' },
  { value: 'external', label: 'External' },
  { value: 'webhook', label: 'Webhook' },
  { value: 'script', label: 'Script' }
];

export function ToolModal({ tool, isOpen, onClose, onSave }: ToolModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    function_name: '',
    schema: '{}',
    category: 'general',
    implementation_type: 'internal',
    endpoint_url: '',
    timeout_seconds: '30',
    max_retries: '3',
    is_public: false,
    metadata: '{}',
    tags: '',
    documentation: '',
    examples: '[]'
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (tool) {
      setFormData({
        name: tool.name || '',
        description: tool.description || '',
        function_name: tool.function_name || '',
        schema: JSON.stringify(tool.schema || {}, null, 2),
        category: tool.category || 'general',
        implementation_type: tool.implementation_type || 'internal',
        endpoint_url: tool.endpoint_url || '',
        timeout_seconds: tool.timeout_seconds?.toString() || '30',
        max_retries: tool.max_retries?.toString() || '3',
        is_public: tool.is_public || false,
        metadata: JSON.stringify(tool.metadata || {}, null, 2),
        tags: tool.tags?.join(', ') || '',
        documentation: tool.documentation || '',
        examples: JSON.stringify(tool.examples || [], null, 2)
      });
    } else {
      setFormData({
        name: '',
        description: '',
        function_name: '',
        schema: '{}',
        category: 'general',
        implementation_type: 'internal',
        endpoint_url: '',
        timeout_seconds: '30',
        max_retries: '3',
        is_public: false,
        metadata: '{}',
        tags: '',
        documentation: '',
        examples: '[]'
      });
    }
    setErrors({});
  }, [tool, isOpen]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    } else if (formData.name.length < 2) {
      newErrors.name = 'Name must be at least 2 characters';
    }

    if (!formData.function_name.trim()) {
      newErrors.function_name = 'Function name is required';
    } else if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(formData.function_name)) {
      newErrors.function_name = 'Function name must be a valid identifier (letters, numbers, underscores)';
    }

    if (!formData.category) {
      newErrors.category = 'Category is required';
    }

    if (!formData.implementation_type) {
      newErrors.implementation_type = 'Implementation type is required';
    }

    // Validate schema JSON
    if (formData.schema.trim()) {
      try {
        JSON.parse(formData.schema);
      } catch (e) {
        newErrors.schema = 'Schema must be valid JSON';
      }
    }

    // Validate timeout_seconds
    const timeout = Number(formData.timeout_seconds);
    if (!Number.isInteger(timeout) || timeout < 1 || timeout > 600) {
      newErrors.timeout_seconds = 'Timeout must be between 1 and 600 seconds';
    }

    // Validate max_retries
    const retries = Number(formData.max_retries);
    if (!Number.isInteger(retries) || retries < 0 || retries > 10) {
      newErrors.max_retries = 'Max retries must be between 0 and 10';
    }

    // Validate endpoint_url for external implementations
    if (['external', 'webhook'].includes(formData.implementation_type) && !formData.endpoint_url.trim()) {
      newErrors.endpoint_url = 'Endpoint URL is required for external/webhook implementations';
    }

    // Validate metadata JSON
    if (formData.metadata.trim()) {
      try {
        JSON.parse(formData.metadata);
      } catch (e) {
        newErrors.metadata = 'Metadata must be valid JSON';
      }
    }

    // Validate examples JSON
    if (formData.examples.trim()) {
      try {
        const parsed = JSON.parse(formData.examples);
        if (!Array.isArray(parsed)) {
          newErrors.examples = 'Examples must be a JSON array';
        }
      } catch (e) {
        newErrors.examples = 'Examples must be valid JSON array';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setSaving(true);
    try {
      const submitData: CreateToolRequest | UpdateToolRequest = {
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
        function_name: formData.function_name.trim(),
        schema: formData.schema.trim() ? JSON.parse(formData.schema) : {},
        category: formData.category,
        implementation_type: formData.implementation_type,
        endpoint_url: formData.endpoint_url.trim() || undefined,
        timeout_seconds: Number(formData.timeout_seconds),
        max_retries: Number(formData.max_retries),
        is_public: formData.is_public,
        metadata: formData.metadata.trim() ? JSON.parse(formData.metadata) : undefined,
        tags: formData.tags.trim() 
          ? formData.tags.split(',').map(tag => tag.trim()).filter(tag => tag)
          : undefined,
        documentation: formData.documentation.trim() || undefined,
        examples: formData.examples.trim() ? JSON.parse(formData.examples) : undefined
      };

      await onSave(submitData);
      onClose();
    } catch (error) {
      // Error handling is done in parent component
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    if (!saving) {
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div 
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
        padding: '1rem'
      }}
      onClick={(e) => e.target === e.currentTarget && handleClose()}
    >
      <div style={{
        backgroundColor: 'white',
        borderRadius: '8px',
        width: '100%',
        maxWidth: '800px',
        maxHeight: '90vh',
        overflow: 'hidden',
        boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
      }}>
        <div style={{
          padding: '1.5rem',
          borderBottom: '1px solid #e5e7eb'
        }}>
          <h2 style={{ 
            margin: 0, 
            fontSize: '1.25rem', 
            fontWeight: '600', 
            color: '#111827' 
          }}>
            {tool ? 'Edit Tool' : 'Create Tool'}
          </h2>
        </div>

        <div style={{ 
          padding: '1.5rem',
          maxHeight: 'calc(90vh - 140px)',
          overflowY: 'auto'
        }}>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {/* Name and Function Name */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Name <span style={{ color: '#dc2626' }}>*</span>
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.name ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    transition: 'border-color 0.2s'
                  }}
                  onFocus={(e) => e.target.style.borderColor = errors.name ? '#dc2626' : '#3b82f6'}
                  onBlur={(e) => e.target.style.borderColor = errors.name ? '#dc2626' : '#d1d5db'}
                  placeholder="Enter tool name"
                />
                {errors.name && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.name}
                  </p>
                )}
              </div>

              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Function Name <span style={{ color: '#dc2626' }}>*</span>
                </label>
                <input
                  type="text"
                  value={formData.function_name}
                  onChange={(e) => setFormData(prev => ({ ...prev, function_name: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.function_name ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    fontFamily: 'monospace',
                    outline: 'none',
                    transition: 'border-color 0.2s'
                  }}
                  onFocus={(e) => e.target.style.borderColor = errors.function_name ? '#dc2626' : '#3b82f6'}
                  onBlur={(e) => e.target.style.borderColor = errors.function_name ? '#dc2626' : '#d1d5db'}
                  placeholder="function_name"
                />
                {errors.function_name && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.function_name}
                  </p>
                )}
              </div>
            </div>

            {/* Description */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Description
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                rows={2}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem',
                  outline: 'none',
                  transition: 'border-color 0.2s',
                  resize: 'vertical',
                  minHeight: '60px'
                }}
                onFocus={(e) => e.target.style.borderColor = '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = '#d1d5db'}
                placeholder="Brief description of what this tool does"
              />
            </div>

            {/* Category, Implementation Type, and Public flag */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr auto', gap: '1rem', alignItems: 'end' }}>
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Category <span style={{ color: '#dc2626' }}>*</span>
                </label>
                <select
                  value={formData.category}
                  onChange={(e) => setFormData(prev => ({ ...prev, category: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.category ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    backgroundColor: 'white'
                  }}
                >
                  {TOOL_CATEGORIES.map(category => (
                    <option key={category.value} value={category.value}>
                      {category.label}
                    </option>
                  ))}
                </select>
                {errors.category && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.category}
                  </p>
                )}
              </div>

              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Implementation <span style={{ color: '#dc2626' }}>*</span>
                </label>
                <select
                  value={formData.implementation_type}
                  onChange={(e) => setFormData(prev => ({ ...prev, implementation_type: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.implementation_type ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    backgroundColor: 'white'
                  }}
                >
                  {IMPLEMENTATION_TYPES.map(type => (
                    <option key={type.value} value={type.value}>
                      {type.label}
                    </option>
                  ))}
                </select>
                {errors.implementation_type && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.implementation_type}
                  </p>
                )}
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 0' }}>
                <input
                  type="checkbox"
                  id="is_public"
                  checked={formData.is_public}
                  onChange={(e) => setFormData(prev => ({ ...prev, is_public: e.target.checked }))}
                  style={{ margin: 0 }}
                />
                <label
                  htmlFor="is_public"
                  style={{
                    fontSize: '0.875rem',
                    fontWeight: '500',
                    color: '#374151',
                    cursor: 'pointer'
                  }}
                >
                  Public
                </label>
              </div>
            </div>

            {/* Endpoint URL (conditional) */}
            {(['external', 'webhook'].includes(formData.implementation_type)) && (
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Endpoint URL <span style={{ color: '#dc2626' }}>*</span>
                </label>
                <input
                  type="url"
                  value={formData.endpoint_url}
                  onChange={(e) => setFormData(prev => ({ ...prev, endpoint_url: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.endpoint_url ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    transition: 'border-color 0.2s'
                  }}
                  onFocus={(e) => e.target.style.borderColor = errors.endpoint_url ? '#dc2626' : '#3b82f6'}
                  onBlur={(e) => e.target.style.borderColor = errors.endpoint_url ? '#dc2626' : '#d1d5db'}
                  placeholder="https://api.example.com/tool"
                />
                {errors.endpoint_url && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.endpoint_url}
                  </p>
                )}
              </div>
            )}

            {/* Timeout and Max Retries */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Timeout (seconds)
                </label>
                <input
                  type="number"
                  value={formData.timeout_seconds}
                  onChange={(e) => setFormData(prev => ({ ...prev, timeout_seconds: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.timeout_seconds ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    transition: 'border-color 0.2s'
                  }}
                  onFocus={(e) => e.target.style.borderColor = errors.timeout_seconds ? '#dc2626' : '#3b82f6'}
                  onBlur={(e) => e.target.style.borderColor = errors.timeout_seconds ? '#dc2626' : '#d1d5db'}
                  min="1"
                  max="600"
                />
                {errors.timeout_seconds && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.timeout_seconds}
                  </p>
                )}
              </div>

              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Max Retries
                </label>
                <input
                  type="number"
                  value={formData.max_retries}
                  onChange={(e) => setFormData(prev => ({ ...prev, max_retries: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.max_retries ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    transition: 'border-color 0.2s'
                  }}
                  onFocus={(e) => e.target.style.borderColor = errors.max_retries ? '#dc2626' : '#3b82f6'}
                  onBlur={(e) => e.target.style.borderColor = errors.max_retries ? '#dc2626' : '#d1d5db'}
                  min="0"
                  max="10"
                />
                {errors.max_retries && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.max_retries}
                  </p>
                )}
              </div>
            </div>

            {/* Schema */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Schema (JSON)
              </label>
              <textarea
                value={formData.schema}
                onChange={(e) => setFormData(prev => ({ ...prev, schema: e.target.value }))}
                rows={6}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: `1px solid ${errors.schema ? '#dc2626' : '#d1d5db'}`,
                  borderRadius: '0.375rem',
                  fontSize: '0.75rem',
                  fontFamily: 'monospace',
                  outline: 'none',
                  transition: 'border-color 0.2s',
                  resize: 'vertical',
                  minHeight: '120px'
                }}
                onFocus={(e) => e.target.style.borderColor = errors.schema ? '#dc2626' : '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = errors.schema ? '#dc2626' : '#d1d5db'}
                placeholder='{"type": "object", "properties": {...}}'
              />
              {errors.schema && (
                <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                  {errors.schema}
                </p>
              )}
            </div>

            {/* Tags */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Tags
              </label>
              <input
                type="text"
                value={formData.tags}
                onChange={(e) => setFormData(prev => ({ ...prev, tags: e.target.value }))}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem',
                  outline: 'none',
                  transition: 'border-color 0.2s'
                }}
                onFocus={(e) => e.target.style.borderColor = '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = '#d1d5db'}
                placeholder="Comma-separated tags (e.g., api, utility, data)"
              />
            </div>

            {/* Documentation */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Documentation
              </label>
              <textarea
                value={formData.documentation}
                onChange={(e) => setFormData(prev => ({ ...prev, documentation: e.target.value }))}
                rows={4}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem',
                  outline: 'none',
                  transition: 'border-color 0.2s',
                  resize: 'vertical',
                  minHeight: '80px'
                }}
                onFocus={(e) => e.target.style.borderColor = '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = '#d1d5db'}
                placeholder="Detailed documentation for this tool..."
              />
            </div>

            {/* Examples */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Examples (JSON Array)
              </label>
              <textarea
                value={formData.examples}
                onChange={(e) => setFormData(prev => ({ ...prev, examples: e.target.value }))}
                rows={3}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: `1px solid ${errors.examples ? '#dc2626' : '#d1d5db'}`,
                  borderRadius: '0.375rem',
                  fontSize: '0.75rem',
                  fontFamily: 'monospace',
                  outline: 'none',
                  transition: 'border-color 0.2s',
                  resize: 'vertical',
                  minHeight: '70px'
                }}
                onFocus={(e) => e.target.style.borderColor = errors.examples ? '#dc2626' : '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = errors.examples ? '#dc2626' : '#d1d5db'}
                placeholder='[{"description": "Example usage", "input": {...}, "output": {...}}]'
              />
              {errors.examples && (
                <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                  {errors.examples}
                </p>
              )}
            </div>

            {/* Metadata */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Metadata (JSON)
              </label>
              <textarea
                value={formData.metadata}
                onChange={(e) => setFormData(prev => ({ ...prev, metadata: e.target.value }))}
                rows={3}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: `1px solid ${errors.metadata ? '#dc2626' : '#d1d5db'}`,
                  borderRadius: '0.375rem',
                  fontSize: '0.75rem',
                  fontFamily: 'monospace',
                  outline: 'none',
                  transition: 'border-color 0.2s',
                  resize: 'vertical',
                  minHeight: '70px'
                }}
                onFocus={(e) => e.target.style.borderColor = errors.metadata ? '#dc2626' : '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = errors.metadata ? '#dc2626' : '#d1d5db'}
                placeholder='{"key": "value"}'
              />
              {errors.metadata && (
                <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                  {errors.metadata}
                </p>
              )}
            </div>
          </form>
        </div>

        <div style={{
          padding: '1.5rem',
          borderTop: '1px solid #e5e7eb',
          display: 'flex',
          gap: '0.75rem',
          justifyContent: 'flex-end'
        }}>
          <button
            type="button"
            onClick={handleClose}
            disabled={saving}
            style={{
              padding: '0.5rem 1rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              backgroundColor: '#f9fafb',
              border: '1px solid #e5e7eb',
              borderRadius: '0.375rem',
              cursor: saving ? 'not-allowed' : 'pointer',
              opacity: saving ? 0.5 : 1,
              transition: 'all 0.2s'
            }}
            onMouseOver={(e) => !saving && (e.currentTarget.style.backgroundColor = '#f3f4f6')}
            onMouseOut={(e) => !saving && (e.currentTarget.style.backgroundColor = '#f9fafb')}
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={handleSubmit}
            disabled={saving}
            style={{
              padding: '0.5rem 1rem',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: 'white',
              backgroundColor: saving ? '#9ca3af' : '#3b82f6',
              border: 'none',
              borderRadius: '0.375rem',
              cursor: saving ? 'not-allowed' : 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => !saving && (e.currentTarget.style.backgroundColor = '#2563eb')}
            onMouseOut={(e) => !saving && (e.currentTarget.style.backgroundColor = '#3b82f6')}
          >
            {saving ? 'Saving...' : (tool ? 'Update' : 'Create')}
          </button>
        </div>
      </div>
    </div>
  );
}