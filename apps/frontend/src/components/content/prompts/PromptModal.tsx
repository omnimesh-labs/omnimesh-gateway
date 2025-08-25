'use client';

import { useState, useEffect } from 'react';
import { Prompt, CreatePromptRequest, UpdatePromptRequest } from '@/lib/api';

interface PromptModalProps {
  prompt?: Prompt;
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: CreatePromptRequest | UpdatePromptRequest) => Promise<void>;
}

const PROMPT_CATEGORIES = [
  { value: 'general', label: 'General' },
  { value: 'coding', label: 'Coding' },
  { value: 'analysis', label: 'Analysis' },
  { value: 'creative', label: 'Creative' },
  { value: 'educational', label: 'Educational' },
  { value: 'business', label: 'Business' },
  { value: 'custom', label: 'Custom' }
];

export function PromptModal({ prompt, isOpen, onClose, onSave }: PromptModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    prompt_template: '',
    parameters: '[]',
    category: 'general',
    metadata: '{}',
    tags: ''
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (prompt) {
      setFormData({
        name: prompt.name || '',
        description: prompt.description || '',
        prompt_template: prompt.prompt_template || '',
        parameters: JSON.stringify(prompt.parameters || [], null, 2),
        category: prompt.category || 'general',
        metadata: JSON.stringify(prompt.metadata || {}, null, 2),
        tags: prompt.tags?.join(', ') || ''
      });
    } else {
      setFormData({
        name: '',
        description: '',
        prompt_template: '',
        parameters: '[]',
        category: 'general',
        metadata: '{}',
        tags: ''
      });
    }
    setErrors({});
  }, [prompt, isOpen]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    } else if (formData.name.length < 2) {
      newErrors.name = 'Name must be at least 2 characters';
    }

    if (!formData.prompt_template.trim()) {
      newErrors.prompt_template = 'Prompt template is required';
    }

    if (!formData.category) {
      newErrors.category = 'Category is required';
    }

    // Validate parameters JSON
    if (formData.parameters.trim()) {
      try {
        const parsed = JSON.parse(formData.parameters);
        if (!Array.isArray(parsed)) {
          newErrors.parameters = 'Parameters must be a JSON array';
        }
      } catch (e) {
        newErrors.parameters = 'Parameters must be valid JSON array';
      }
    }

    // Validate metadata JSON
    if (formData.metadata.trim()) {
      try {
        JSON.parse(formData.metadata);
      } catch (e) {
        newErrors.metadata = 'Metadata must be valid JSON';
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
      const submitData: CreatePromptRequest | UpdatePromptRequest = {
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
        prompt_template: formData.prompt_template.trim(),
        parameters: formData.parameters.trim() ? JSON.parse(formData.parameters) : undefined,
        category: formData.category,
        metadata: formData.metadata.trim() ? JSON.parse(formData.metadata) : undefined,
        tags: formData.tags.trim()
          ? formData.tags.split(',').map(tag => tag.trim()).filter(tag => tag)
          : undefined
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
        maxWidth: '700px',
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
            {prompt ? 'Edit Prompt' : 'Create Prompt'}
          </h2>
        </div>

        <div style={{
          padding: '1.5rem',
          maxHeight: 'calc(90vh - 140px)',
          overflowY: 'auto'
        }}>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {/* Name and Category */}
            <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem' }}>
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
                  placeholder="Enter prompt name"
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
                  {PROMPT_CATEGORIES.map(category => (
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
                placeholder="Brief description of the prompt"
              />
            </div>

            {/* Prompt Template */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Prompt Template <span style={{ color: '#dc2626' }}>*</span>
              </label>
              <textarea
                value={formData.prompt_template}
                onChange={(e) => setFormData(prev => ({ ...prev, prompt_template: e.target.value }))}
                rows={8}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: `1px solid ${errors.prompt_template ? '#dc2626' : '#d1d5db'}`,
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem',
                  fontFamily: 'monospace',
                  outline: 'none',
                  transition: 'border-color 0.2s',
                  resize: 'vertical',
                  minHeight: '150px'
                }}
                onFocus={(e) => e.target.style.borderColor = errors.prompt_template ? '#dc2626' : '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = errors.prompt_template ? '#dc2626' : '#d1d5db'}
                placeholder="Enter the prompt template. Use {{variable}} for parameters."
              />
              {errors.prompt_template && (
                <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                  {errors.prompt_template}
                </p>
              )}
            </div>

            {/* Parameters */}
            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Parameters (JSON Array)
              </label>
              <textarea
                value={formData.parameters}
                onChange={(e) => setFormData(prev => ({ ...prev, parameters: e.target.value }))}
                rows={4}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: `1px solid ${errors.parameters ? '#dc2626' : '#d1d5db'}`,
                  borderRadius: '0.375rem',
                  fontSize: '0.75rem',
                  fontFamily: 'monospace',
                  outline: 'none',
                  transition: 'border-color 0.2s',
                  resize: 'vertical',
                  minHeight: '80px'
                }}
                onFocus={(e) => e.target.style.borderColor = errors.parameters ? '#dc2626' : '#3b82f6'}
                onBlur={(e) => e.target.style.borderColor = errors.parameters ? '#dc2626' : '#d1d5db'}
                placeholder='[{"name": "variable", "type": "string", "description": "Variable description"}]'
              />
              {errors.parameters && (
                <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                  {errors.parameters}
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
                placeholder="Comma-separated tags (e.g., creative, writing, template)"
              />
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
            {saving ? 'Saving...' : (prompt ? 'Update' : 'Create')}
          </button>
        </div>
      </div>
    </div>
  );
}
