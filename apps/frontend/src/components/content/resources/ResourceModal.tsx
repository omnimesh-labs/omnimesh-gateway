'use client';

import { useState, useEffect } from 'react';
import { Resource, CreateResourceRequest, UpdateResourceRequest } from '@/lib/api';

interface ResourceModalProps {
  resource?: Resource;
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: CreateResourceRequest | UpdateResourceRequest) => Promise<void>;
}

const RESOURCE_TYPES = [
  { value: 'file', label: 'File' },
  { value: 'url', label: 'URL' },
  { value: 'database', label: 'Database' },
  { value: 'api', label: 'API' },
  { value: 'memory', label: 'Memory' },
  { value: 'custom', label: 'Custom' }
];

export function ResourceModal({ resource, isOpen, onClose, onSave }: ResourceModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    resource_type: 'file',
    uri: '',
    mime_type: '',
    size_bytes: '',
    metadata: '{}',
    tags: ''
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (resource) {
      setFormData({
        name: resource.name || '',
        description: resource.description || '',
        resource_type: resource.resource_type || 'file',
        uri: resource.uri || '',
        mime_type: resource.mime_type || '',
        size_bytes: resource.size_bytes?.toString() || '',
        metadata: JSON.stringify(resource.metadata || {}, null, 2),
        tags: resource.tags?.join(', ') || ''
      });
    } else {
      setFormData({
        name: '',
        description: '',
        resource_type: 'file',
        uri: '',
        mime_type: '',
        size_bytes: '',
        metadata: '{}',
        tags: ''
      });
    }
    setErrors({});
  }, [resource, isOpen]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    } else if (formData.name.length < 2) {
      newErrors.name = 'Name must be at least 2 characters';
    }

    if (!formData.uri.trim()) {
      newErrors.uri = 'URI is required';
    }

    if (!formData.resource_type) {
      newErrors.resource_type = 'Resource type is required';
    }

    // Validate metadata JSON
    if (formData.metadata.trim()) {
      try {
        JSON.parse(formData.metadata);
      } catch (e) {
        newErrors.metadata = 'Metadata must be valid JSON';
      }
    }

    // Validate size_bytes if provided
    if (formData.size_bytes && !Number.isInteger(Number(formData.size_bytes))) {
      newErrors.size_bytes = 'Size must be a valid number';
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
      const submitData: CreateResourceRequest | UpdateResourceRequest = {
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
        resource_type: formData.resource_type,
        uri: formData.uri.trim(),
        mime_type: formData.mime_type.trim() || undefined,
        size_bytes: formData.size_bytes ? Number(formData.size_bytes) : undefined,
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
        maxWidth: '600px',
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
            {resource ? 'Edit Resource' : 'Create Resource'}
          </h2>
        </div>

        <div style={{ 
          padding: '1.5rem',
          maxHeight: 'calc(90vh - 140px)',
          overflowY: 'auto'
        }}>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {/* Name */}
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
                placeholder="Enter resource name"
              />
              {errors.name && (
                <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                  {errors.name}
                </p>
              )}
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
                rows={3}
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
                placeholder="Optional description"
              />
            </div>

            {/* Resource Type and URI */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '1rem' }}>
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Type <span style={{ color: '#dc2626' }}>*</span>
                </label>
                <select
                  value={formData.resource_type}
                  onChange={(e) => setFormData(prev => ({ ...prev, resource_type: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.resource_type ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    backgroundColor: 'white'
                  }}
                >
                  {RESOURCE_TYPES.map(type => (
                    <option key={type.value} value={type.value}>
                      {type.label}
                    </option>
                  ))}
                </select>
                {errors.resource_type && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.resource_type}
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
                  URI <span style={{ color: '#dc2626' }}>*</span>
                </label>
                <input
                  type="text"
                  value={formData.uri}
                  onChange={(e) => setFormData(prev => ({ ...prev, uri: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.uri ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    transition: 'border-color 0.2s'
                  }}
                  onFocus={(e) => e.target.style.borderColor = errors.uri ? '#dc2626' : '#3b82f6'}
                  onBlur={(e) => e.target.style.borderColor = errors.uri ? '#dc2626' : '#d1d5db'}
                  placeholder="Resource URI or path"
                />
                {errors.uri && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.uri}
                  </p>
                )}
              </div>
            </div>

            {/* MIME Type and Size */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  MIME Type
                </label>
                <input
                  type="text"
                  value={formData.mime_type}
                  onChange={(e) => setFormData(prev => ({ ...prev, mime_type: e.target.value }))}
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
                  placeholder="e.g., application/json"
                />
              </div>

              <div>
                <label style={{
                  display: 'block',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Size (bytes)
                </label>
                <input
                  type="number"
                  value={formData.size_bytes}
                  onChange={(e) => setFormData(prev => ({ ...prev, size_bytes: e.target.value }))}
                  style={{
                    width: '100%',
                    padding: '0.5rem 0.75rem',
                    border: `1px solid ${errors.size_bytes ? '#dc2626' : '#d1d5db'}`,
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem',
                    outline: 'none',
                    transition: 'border-color 0.2s'
                  }}
                  onFocus={(e) => e.target.style.borderColor = errors.size_bytes ? '#dc2626' : '#3b82f6'}
                  onBlur={(e) => e.target.style.borderColor = errors.size_bytes ? '#dc2626' : '#d1d5db'}
                  placeholder="Optional size in bytes"
                  min="0"
                />
                {errors.size_bytes && (
                  <p style={{ color: '#dc2626', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                    {errors.size_bytes}
                  </p>
                )}
              </div>
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
                placeholder="Comma-separated tags (e.g., important, database, prod)"
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
                rows={4}
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
                  minHeight: '100px'
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
            {saving ? 'Saving...' : (resource ? 'Update' : 'Create')}
          </button>
        </div>
      </div>
    </div>
  );
}