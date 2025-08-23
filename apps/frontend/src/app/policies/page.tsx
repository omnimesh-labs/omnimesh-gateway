'use client';

import { useState, useEffect } from 'react';
import { Policy, policyApi, PolicyListResponse } from '@/lib/api';
import { PolicyTable } from '@/components/policies/PolicyTable';
import { PolicyFormModal } from '@/components/policies/PolicyFormModal';
import { Toast } from '@/components/Toast';
import { ProtectedRoute } from '@/components/ProtectedRoute';

export default function PoliciesPage() {
    const [policies, setPolicies] = useState<Policy[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingPolicy, setEditingPolicy] = useState<Policy | undefined>();
    const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
    const [filters, setFilters] = useState({
        type: '',
        is_active: undefined as boolean | undefined,
    });

    const loadPolicies = async () => {
        try {
            setError(null);
            const response: PolicyListResponse = await policyApi.listPolicies({
                type: filters.type || undefined,
                is_active: filters.is_active,
            });
            setPolicies(response.policies || []);
        } catch (err) {
            console.error('Error loading policies:', err);
            setError(err instanceof Error ? err.message : 'Failed to load policies');
            setPolicies([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadPolicies();
    }, [filters.type, filters.is_active]);

    const handleCreatePolicy = () => {
        setEditingPolicy(undefined);
        setIsModalOpen(true);
    };

    const handleEditPolicy = (policy: Policy) => {
        setEditingPolicy(policy);
        setIsModalOpen(true);
    };

    const handleModalClose = () => {
        setIsModalOpen(false);
        setEditingPolicy(undefined);
    };

    const handleModalSave = () => {
        loadPolicies();
        setToast({
            type: 'success',
            message: editingPolicy ? 'Policy updated successfully' : 'Policy created successfully'
        });
    };


    return (
        <ProtectedRoute requireRole="admin">
            <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
                <header style={{ marginBottom: '2rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                        <div>
                            <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#333', marginBottom: '0.5rem' }}>
                                Policy Management
                            </h1>
                            <p style={{ fontSize: '1rem', color: '#666' }}>
                                Create and manage access control, rate limiting, and security policies
                            </p>
                        </div>
                        <button
                            onClick={handleCreatePolicy}
                            style={{
                                backgroundColor: '#3b82f6',
                                color: '#ffffff',
                                border: 'none',
                                padding: '0.75rem 1.25rem',
                                borderRadius: '8px',
                                fontSize: '0.875rem',
                                fontWeight: '500',
                                cursor: 'pointer',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.5rem',
                                transition: 'background-color 0.2s',
                            }}
                            onMouseOver={(e) => {
                                e.currentTarget.style.backgroundColor = '#2563eb';
                            }}
                            onMouseOut={(e) => {
                                e.currentTarget.style.backgroundColor = '#3b82f6';
                            }}
                        >
                            ➕ Create Policy
                        </button>
                    </div>
                </header>

                {/* Filters */}
                <div style={{
                    backgroundColor: '#ffffff',
                    border: '1px solid #e5e7eb',
                    borderRadius: '8px',
                    padding: '16px',
                    marginBottom: '24px',
                }}>
                    <div style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '12px', 
                                fontWeight: '500', 
                                color: '#374151',
                                marginBottom: '4px',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Filter by Type
                            </label>
                            <select
                                value={filters.type}
                                onChange={(e) => setFilters(prev => ({ ...prev, type: e.target.value }))}
                                style={{
                                    padding: '6px 8px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: '4px',
                                    fontSize: '14px',
                                    backgroundColor: '#ffffff',
                                    cursor: 'pointer',
                                    minWidth: '120px',
                                }}
                            >
                                <option value="">All Types</option>
                                <option value="access">Access</option>
                                <option value="rate_limit">Rate Limit</option>
                                <option value="routing">Routing</option>
                                <option value="security">Security</option>
                            </select>
                        </div>

                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '12px', 
                                fontWeight: '500', 
                                color: '#374151',
                                marginBottom: '4px',
                                textTransform: 'uppercase',
                                letterSpacing: '0.05em',
                            }}>
                                Filter by Status
                            </label>
                            <select
                                value={filters.is_active === undefined ? '' : filters.is_active.toString()}
                                onChange={(e) => setFilters(prev => ({ 
                                    ...prev, 
                                    is_active: e.target.value === '' ? undefined : e.target.value === 'true'
                                }))}
                                style={{
                                    padding: '6px 8px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: '4px',
                                    fontSize: '14px',
                                    backgroundColor: '#ffffff',
                                    cursor: 'pointer',
                                    minWidth: '120px',
                                }}
                            >
                                <option value="">All Status</option>
                                <option value="true">Active</option>
                                <option value="false">Inactive</option>
                            </select>
                        </div>

                        <div style={{ marginLeft: 'auto' }}>
                            <button
                                onClick={loadPolicies}
                                disabled={loading}
                                style={{
                                    backgroundColor: '#ffffff',
                                    color: '#374151',
                                    border: '1px solid #d1d5db',
                                    padding: '6px 12px',
                                    borderRadius: '4px',
                                    fontSize: '14px',
                                    fontWeight: '500',
                                    cursor: loading ? 'not-allowed' : 'pointer',
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '6px',
                                }}
                            >
                                🔄 {loading ? 'Loading...' : 'Refresh'}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Content */}
                {loading && policies.length === 0 ? (
                    <div style={{
                        textAlign: 'center',
                        padding: '48px',
                        backgroundColor: '#f9fafb',
                        border: '1px solid #e5e7eb',
                        borderRadius: '8px',
                    }}>
                        <div style={{ fontSize: '18px', marginBottom: '8px' }}>⏳</div>
                        <div style={{ fontSize: '16px', color: '#6b7280' }}>Loading policies...</div>
                    </div>
                ) : error ? (
                    <div style={{
                        textAlign: 'center',
                        padding: '48px',
                        backgroundColor: '#fef2f2',
                        border: '1px solid #fecaca',
                        borderRadius: '8px',
                    }}>
                        <div style={{ fontSize: '18px', marginBottom: '8px' }}>❌</div>
                        <div style={{ fontSize: '16px', color: '#dc2626', marginBottom: '16px' }}>
                            Failed to load policies
                        </div>
                        <div style={{ fontSize: '14px', color: '#7f1d1d', marginBottom: '16px' }}>
                            {error}
                        </div>
                        <button
                            onClick={loadPolicies}
                            style={{
                                backgroundColor: '#dc2626',
                                color: '#ffffff',
                                border: 'none',
                                padding: '8px 16px',
                                borderRadius: '4px',
                                fontSize: '14px',
                                fontWeight: '500',
                                cursor: 'pointer',
                            }}
                        >
                            Try Again
                        </button>
                    </div>
                ) : (
                    <>
                        {/* Policies Count */}
                        <div style={{ 
                            marginBottom: '16px',
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                        }}>
                            <div style={{ fontSize: '14px', color: '#6b7280' }}>
                                {policies.length === 0 
                                    ? 'No policies found' 
                                    : `${policies.length} ${policies.length === 1 ? 'policy' : 'policies'} found`
                                }
                            </div>
                        </div>

                        {/* Policy Table */}
                        <PolicyTable
                            policies={policies}
                            onRefresh={loadPolicies}
                            onEdit={handleEditPolicy}
                        />
                    </>
                )}

                {/* Policy Form Modal */}
                <PolicyFormModal
                    isOpen={isModalOpen}
                    onClose={handleModalClose}
                    onSave={handleModalSave}
                    policy={editingPolicy}
                />

                {/* Toast Notifications */}
                {toast && (
                    <Toast
                        type={toast.type}
                        message={toast.message}
                        onClose={() => setToast(null)}
                    />
                )}
            </div>
        </ProtectedRoute>
    );
}