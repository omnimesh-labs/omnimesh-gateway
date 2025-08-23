'use client';

import { useState, useEffect } from 'react';
import { Policy, CreatePolicyRequest, UpdatePolicyRequest, policyApi } from '@/lib/api';

interface PolicyFormModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSave: () => void;
    policy?: Policy;
}

export function PolicyFormModal({ isOpen, onClose, onSave, policy }: PolicyFormModalProps) {
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        type: 'access' as 'access' | 'rate_limit' | 'security',
        priority: 100,
        conditions: '',
        actions: '',
        is_active: true,
    });
    const [loading, setLoading] = useState(false);
    const [errors, setErrors] = useState<Record<string, string>>({});

    // Reset form when modal opens/closes or policy changes
    useEffect(() => {
        if (isOpen) {
            if (policy) {
                // Editing existing policy
                setFormData({
                    name: policy.name,
                    description: policy.description,
                    type: policy.type,
                    priority: policy.priority,
                    conditions: JSON.stringify(policy.conditions, null, 2),
                    actions: JSON.stringify(policy.actions, null, 2),
                    is_active: policy.is_active,
                });
            } else {
                // Creating new policy
                setFormData({
                    name: '',
                    description: '',
                    type: 'access',
                    priority: 100,
                    conditions: '{}',
                    actions: '{}',
                    is_active: true,
                });
            }
            setErrors({});
        }
    }, [isOpen, policy]);

    const validateForm = () => {
        const newErrors: Record<string, string> = {};

        if (!formData.name.trim()) {
            newErrors.name = 'Name is required';
        }

        if (!formData.type) {
            newErrors.type = 'Type is required';
        }

        if (formData.priority < 1 || formData.priority > 1000) {
            newErrors.priority = 'Priority must be between 1 and 1000';
        }

        // Validate JSON
        try {
            JSON.parse(formData.conditions);
        } catch (e) {
            newErrors.conditions = 'Invalid JSON format';
        }

        try {
            JSON.parse(formData.actions);
        } catch (e) {
            newErrors.actions = 'Invalid JSON format';
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!validateForm()) {
            return;
        }

        setLoading(true);
        try {
            const conditions = JSON.parse(formData.conditions);
            const actions = JSON.parse(formData.actions);

            if (policy) {
                // Update existing policy
                const updateData: UpdatePolicyRequest = {
                    name: formData.name,
                    description: formData.description,
                    priority: formData.priority,
                    conditions,
                    actions,
                    is_active: formData.is_active,
                };
                await policyApi.updatePolicy(policy.id, updateData);
            } else {
                // Create new policy
                const createData: CreatePolicyRequest = {
                    name: formData.name,
                    description: formData.description,
                    type: formData.type,
                    priority: formData.priority,
                    conditions,
                    actions,
                };
                await policyApi.createPolicy(createData);
            }

            onSave();
            onClose();
        } catch (error) {
            console.error('Error saving policy:', error);
            setErrors({ submit: error instanceof Error ? error.message : 'Failed to save policy' });
        } finally {
            setLoading(false);
        }
    };

    const handleChange = (field: keyof typeof formData, value: any) => {
        setFormData(prev => ({ ...prev, [field]: value }));
        // Clear error when user starts typing
        if (errors[field]) {
            setErrors(prev => ({ ...prev, [field]: '' }));
        }
    };

    if (!isOpen) return null;

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
            zIndex: 1000,
            padding: '20px',
        }}>
            <div style={{
                backgroundColor: '#ffffff',
                borderRadius: '8px',
                padding: '24px',
                width: '100%',
                maxWidth: '600px',
                maxHeight: '90vh',
                overflowY: 'auto',
                boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
            }}>
                <div style={{ 
                    display: 'flex', 
                    justifyContent: 'space-between', 
                    alignItems: 'center', 
                    marginBottom: '24px' 
                }}>
                    <h2 style={{ 
                        fontSize: '20px', 
                        fontWeight: '600', 
                        color: '#111827', 
                        margin: 0 
                    }}>
                        {policy ? 'Edit Policy' : 'Create New Policy'}
                    </h2>
                    <button
                        onClick={onClose}
                        style={{
                            backgroundColor: 'transparent',
                            border: 'none',
                            fontSize: '24px',
                            cursor: 'pointer',
                            color: '#6b7280',
                            padding: '4px',
                        }}
                    >
                        ×
                    </button>
                </div>

                <form onSubmit={handleSubmit}>
                    <div style={{ display: 'grid', gap: '16px' }}>
                        {/* Name */}
                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '14px', 
                                fontWeight: '500', 
                                color: '#374151',
                                marginBottom: '4px'
                            }}>
                                Name *
                            </label>
                            <input
                                type="text"
                                value={formData.name}
                                onChange={(e) => handleChange('name', e.target.value)}
                                style={{
                                    width: '100%',
                                    padding: '8px 12px',
                                    border: `1px solid ${errors.name ? '#ef4444' : '#d1d5db'}`,
                                    borderRadius: '6px',
                                    fontSize: '14px',
                                    boxSizing: 'border-box',
                                }}
                                placeholder="Enter policy name"
                            />
                            {errors.name && (
                                <div style={{ fontSize: '12px', color: '#ef4444', marginTop: '4px' }}>
                                    {errors.name}
                                </div>
                            )}
                        </div>

                        {/* Description */}
                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '14px', 
                                fontWeight: '500', 
                                color: '#374151',
                                marginBottom: '4px'
                            }}>
                                Description
                            </label>
                            <textarea
                                value={formData.description}
                                onChange={(e) => handleChange('description', e.target.value)}
                                rows={3}
                                style={{
                                    width: '100%',
                                    padding: '8px 12px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: '6px',
                                    fontSize: '14px',
                                    boxSizing: 'border-box',
                                    resize: 'vertical',
                                }}
                                placeholder="Enter policy description"
                            />
                        </div>

                        {/* Type and Priority Row */}
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                            {/* Type */}
                            <div>
                                <label style={{ 
                                    display: 'block', 
                                    fontSize: '14px', 
                                    fontWeight: '500', 
                                    color: '#374151',
                                    marginBottom: '4px'
                                }}>
                                    Type *
                                </label>
                                <select
                                    value={formData.type}
                                    onChange={(e) => handleChange('type', e.target.value)}
                                    disabled={!!policy} // Disable when editing
                                    style={{
                                        width: '100%',
                                        padding: '8px 12px',
                                        border: `1px solid ${errors.type ? '#ef4444' : '#d1d5db'}`,
                                        borderRadius: '6px',
                                        fontSize: '14px',
                                        backgroundColor: policy ? '#f9fafb' : '#ffffff',
                                        cursor: policy ? 'not-allowed' : 'pointer',
                                        boxSizing: 'border-box',
                                    }}
                                >
                                    <option value="access">Access</option>
                                    <option value="rate_limit">Rate Limit</option>
                                    <option value="security">Security</option>
                                </select>
                                {errors.type && (
                                    <div style={{ fontSize: '12px', color: '#ef4444', marginTop: '4px' }}>
                                        {errors.type}
                                    </div>
                                )}
                            </div>

                            {/* Priority */}
                            <div>
                                <label style={{ 
                                    display: 'block', 
                                    fontSize: '14px', 
                                    fontWeight: '500', 
                                    color: '#374151',
                                    marginBottom: '4px'
                                }}>
                                    Priority * (1-1000)
                                </label>
                                <input
                                    type="number"
                                    min="1"
                                    max="1000"
                                    value={formData.priority}
                                    onChange={(e) => handleChange('priority', parseInt(e.target.value) || 0)}
                                    style={{
                                        width: '100%',
                                        padding: '8px 12px',
                                        border: `1px solid ${errors.priority ? '#ef4444' : '#d1d5db'}`,
                                        borderRadius: '6px',
                                        fontSize: '14px',
                                        boxSizing: 'border-box',
                                    }}
                                />
                                {errors.priority && (
                                    <div style={{ fontSize: '12px', color: '#ef4444', marginTop: '4px' }}>
                                        {errors.priority}
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Conditions */}
                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '14px', 
                                fontWeight: '500', 
                                color: '#374151',
                                marginBottom: '4px'
                            }}>
                                Conditions * (JSON)
                            </label>
                            <textarea
                                value={formData.conditions}
                                onChange={(e) => handleChange('conditions', e.target.value)}
                                rows={6}
                                style={{
                                    width: '100%',
                                    padding: '8px 12px',
                                    border: `1px solid ${errors.conditions ? '#ef4444' : '#d1d5db'}`,
                                    borderRadius: '6px',
                                    fontSize: '12px',
                                    fontFamily: 'monospace',
                                    boxSizing: 'border-box',
                                    resize: 'vertical',
                                }}
                                placeholder='{"path": "/api/*", "method": "GET"}'
                            />
                            {errors.conditions && (
                                <div style={{ fontSize: '12px', color: '#ef4444', marginTop: '4px' }}>
                                    {errors.conditions}
                                </div>
                            )}
                        </div>

                        {/* Actions */}
                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '14px', 
                                fontWeight: '500', 
                                color: '#374151',
                                marginBottom: '4px'
                            }}>
                                Actions * (JSON)
                            </label>
                            <textarea
                                value={formData.actions}
                                onChange={(e) => handleChange('actions', e.target.value)}
                                rows={6}
                                style={{
                                    width: '100%',
                                    padding: '8px 12px',
                                    border: `1px solid ${errors.actions ? '#ef4444' : '#d1d5db'}`,
                                    borderRadius: '6px',
                                    fontSize: '12px',
                                    fontFamily: 'monospace',
                                    boxSizing: 'border-box',
                                    resize: 'vertical',
                                }}
                                placeholder='{"allow": true, "rate_limit": 100}'
                            />
                            {errors.actions && (
                                <div style={{ fontSize: '12px', color: '#ef4444', marginTop: '4px' }}>
                                    {errors.actions}
                                </div>
                            )}
                        </div>

                        {/* Status (only when editing) */}
                        {policy && (
                            <div>
                                <label style={{ 
                                    display: 'flex', 
                                    alignItems: 'center',
                                    fontSize: '14px', 
                                    fontWeight: '500', 
                                    color: '#374151',
                                    cursor: 'pointer',
                                }}>
                                    <input
                                        type="checkbox"
                                        checked={formData.is_active}
                                        onChange={(e) => handleChange('is_active', e.target.checked)}
                                        style={{ marginRight: '8px' }}
                                    />
                                    Policy is active
                                </label>
                            </div>
                        )}

                        {/* Submit Error */}
                        {errors.submit && (
                            <div style={{ 
                                padding: '12px',
                                backgroundColor: '#fef2f2',
                                border: '1px solid #fecaca',
                                borderRadius: '6px',
                                color: '#dc2626',
                                fontSize: '14px',
                            }}>
                                {errors.submit}
                            </div>
                        )}

                        {/* Buttons */}
                        <div style={{ 
                            display: 'flex', 
                            justifyContent: 'flex-end', 
                            gap: '12px', 
                            marginTop: '8px' 
                        }}>
                            <button
                                type="button"
                                onClick={onClose}
                                style={{
                                    padding: '8px 16px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: '6px',
                                    backgroundColor: '#ffffff',
                                    color: '#374151',
                                    fontSize: '14px',
                                    fontWeight: '500',
                                    cursor: 'pointer',
                                }}
                            >
                                Cancel
                            </button>
                            <button
                                type="submit"
                                disabled={loading}
                                style={{
                                    padding: '8px 16px',
                                    border: 'none',
                                    borderRadius: '6px',
                                    backgroundColor: loading ? '#9ca3af' : '#3b82f6',
                                    color: '#ffffff',
                                    fontSize: '14px',
                                    fontWeight: '500',
                                    cursor: loading ? 'not-allowed' : 'pointer',
                                }}
                            >
                                {loading ? 'Saving...' : (policy ? 'Update Policy' : 'Create Policy')}
                            </button>
                        </div>
                    </div>
                </form>
            </div>
        </div>
    );
}