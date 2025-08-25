'use client';

import { useState } from 'react';
import { Policy, policyApi } from '@/lib/api';

interface PolicyTableProps {
    policies: Policy[];
    onRefresh: () => void;
    onEdit: (policy: Policy) => void;
}

export function PolicyTable({ policies, onRefresh, onEdit }: PolicyTableProps) {
    const [loadingDelete, setLoadingDelete] = useState<string | null>(null);

    const handleDelete = async (policy: Policy) => {
        if (!confirm(`Are you sure you want to delete the policy "${policy.name}"?`)) {
            return;
        }

        setLoadingDelete(policy.id);
        try {
            await policyApi.deletePolicy(policy.id);
            onRefresh();
        } catch (error) {
            console.error('Error deleting policy:', error);
            alert('Failed to delete policy: ' + (error instanceof Error ? error.message : 'Unknown error'));
        } finally {
            setLoadingDelete(null);
        }
    };

    const getStatusBadge = (isActive: boolean) => ({
        padding: '4px 8px',
        borderRadius: '4px',
        fontSize: '12px',
        fontWeight: '500',
        backgroundColor: isActive ? '#dcfce7' : '#fef3c7',
        color: isActive ? '#166534' : '#92400e',
    });

    const getTypeBadge = (type: string) => {
        const typeColors: Record<string, { bg: string; text: string }> = {
            access: { bg: '#dbeafe', text: '#1e40af' },
            rate_limit: { bg: '#fef3c7', text: '#92400e' },
            routing: { bg: '#f3e8ff', text: '#7c2d12' },
            security: { bg: '#fee2e2', text: '#dc2626' },
        };

        const colors = typeColors[type] || { bg: '#f3f4f6', text: '#374151' };

        return {
            padding: '4px 8px',
            borderRadius: '4px',
            fontSize: '12px',
            fontWeight: '500',
            backgroundColor: colors.bg,
            color: colors.text,
        };
    };

    if (policies.length === 0) {
        return (
            <div style={{
                textAlign: 'center',
                padding: '48px 24px',
                color: '#6b7280',
                backgroundColor: '#f9fafb',
                borderRadius: '8px',
                border: '1px solid #e5e7eb',
            }}>
                <div style={{ fontSize: '18px', marginBottom: '8px' }}>📋</div>
                <div style={{ fontSize: '16px', fontWeight: '500', marginBottom: '4px' }}>No policies found</div>
                <div style={{ fontSize: '14px' }}>Create your first policy to get started</div>
            </div>
        );
    }

    return (
        <div style={{
            backgroundColor: '#ffffff',
            border: '1px solid #e5e7eb',
            borderRadius: '8px',
            overflow: 'hidden',
        }}>
            <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                        <tr style={{ backgroundColor: '#f9fafb' }}>
                            <th style={{
                                padding: '12px 16px',
                                textAlign: 'left',
                                fontSize: '12px',
                                fontWeight: '500',
                                color: '#374151',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Name
                            </th>
                            <th style={{
                                padding: '12px 16px',
                                textAlign: 'left',
                                fontSize: '12px',
                                fontWeight: '500',
                                color: '#374151',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Type
                            </th>
                            <th style={{
                                padding: '12px 16px',
                                textAlign: 'center',
                                fontSize: '12px',
                                fontWeight: '500',
                                color: '#374151',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Priority
                            </th>
                            <th style={{
                                padding: '12px 16px',
                                textAlign: 'center',
                                fontSize: '12px',
                                fontWeight: '500',
                                color: '#374151',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Status
                            </th>
                            <th style={{
                                padding: '12px 16px',
                                textAlign: 'left',
                                fontSize: '12px',
                                fontWeight: '500',
                                color: '#374151',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Created
                            </th>
                            <th style={{
                                padding: '12px 16px',
                                textAlign: 'center',
                                fontSize: '12px',
                                fontWeight: '500',
                                color: '#374151',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Actions
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {policies.map((policy) => (
                            <tr key={policy.id} style={{ borderTop: '1px solid #e5e7eb' }}>
                                <td style={{ padding: '16px' }}>
                                    <div style={{ fontWeight: '500', color: '#111827', marginBottom: '4px' }}>
                                        {policy.name}
                                    </div>
                                    {policy.description && (
                                        <div style={{ fontSize: '14px', color: '#6b7280', lineHeight: '1.4' }}>
                                            {policy.description}
                                        </div>
                                    )}
                                </td>
                                <td style={{ padding: '16px' }}>
                                    <span style={getTypeBadge(policy.type)}>
                                        {policy.type.replace('_', ' ')}
                                    </span>
                                </td>
                                <td style={{ padding: '16px', textAlign: 'center' }}>
                                    <span style={{
                                        fontSize: '14px',
                                        fontWeight: '500',
                                        color: '#111827',
                                    }}>
                                        {policy.priority}
                                    </span>
                                </td>
                                <td style={{ padding: '16px', textAlign: 'center' }}>
                                    <span style={getStatusBadge(policy.is_active)}>
                                        {policy.is_active ? 'Active' : 'Inactive'}
                                    </span>
                                </td>
                                <td style={{ padding: '16px' }}>
                                    <div style={{ fontSize: '14px', color: '#6b7280' }}>
                                        {new Date(policy.created_at).toLocaleDateString()}
                                    </div>
                                    <div style={{ fontSize: '12px', color: '#9ca3af' }}>
                                        {new Date(policy.created_at).toLocaleTimeString()}
                                    </div>
                                </td>
                                <td style={{ padding: '16px', textAlign: 'center' }}>
                                    <div style={{ display: 'flex', gap: '8px', justifyContent: 'center' }}>
                                        <button
                                            onClick={() => onEdit(policy)}
                                            style={{
                                                backgroundColor: '#3b82f6',
                                                color: '#ffffff',
                                                border: 'none',
                                                padding: '6px 12px',
                                                borderRadius: '4px',
                                                fontSize: '12px',
                                                fontWeight: '500',
                                                cursor: 'pointer',
                                                transition: 'background-color 0.2s',
                                            }}
                                            onMouseOver={(e) => {
                                                e.currentTarget.style.backgroundColor = '#2563eb';
                                            }}
                                            onMouseOut={(e) => {
                                                e.currentTarget.style.backgroundColor = '#3b82f6';
                                            }}
                                        >
                                            Edit
                                        </button>
                                        <button
                                            onClick={() => handleDelete(policy)}
                                            disabled={loadingDelete === policy.id}
                                            style={{
                                                backgroundColor: loadingDelete === policy.id ? '#9ca3af' : '#ef4444',
                                                color: '#ffffff',
                                                border: 'none',
                                                padding: '6px 12px',
                                                borderRadius: '4px',
                                                fontSize: '12px',
                                                fontWeight: '500',
                                                cursor: loadingDelete === policy.id ? 'not-allowed' : 'pointer',
                                                transition: 'background-color 0.2s',
                                            }}
                                            onMouseOver={(e) => {
                                                if (loadingDelete !== policy.id) {
                                                    e.currentTarget.style.backgroundColor = '#dc2626';
                                                }
                                            }}
                                            onMouseOut={(e) => {
                                                if (loadingDelete !== policy.id) {
                                                    e.currentTarget.style.backgroundColor = '#ef4444';
                                                }
                                            }}
                                        >
                                            {loadingDelete === policy.id ? '...' : 'Delete'}
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
