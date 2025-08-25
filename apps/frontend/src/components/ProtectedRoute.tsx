'use client';

import React, { ReactNode } from 'react';
import { useAuth } from './AuthContext';
import { LoginForm } from './LoginForm';

interface ProtectedRouteProps {
    children: ReactNode;
    fallback?: ReactNode;
    requireRole?: 'admin' | 'user' | 'viewer' | 'system_admin';
}

export function ProtectedRoute({ children, fallback, requireRole }: ProtectedRouteProps) {
    const { isAuthenticated, isLoading, user } = useAuth();

    if (isLoading) {
        return (
            <div style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                minHeight: '60vh',
                fontSize: '1.125rem',
                color: '#6b7280',
            }}>
                Loading...
            </div>
        );
    }

    if (!isAuthenticated) {
        return fallback || (
            <div style={{
                padding: '2rem',
                minHeight: '60vh',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
            }}>
                <div style={{ marginBottom: '2rem', textAlign: 'center' }}>
                    <h1 style={{
                        fontSize: '1.5rem',
                        fontWeight: 'bold',
                        marginBottom: '0.5rem',
                        color: '#111827',
                    }}>
                        Authentication Required
                    </h1>
                    <p style={{ color: '#6b7280' }}>
                        Please sign in to access the MCP Gateway dashboard.
                    </p>
                </div>
                <LoginForm />
            </div>
        );
    }

    // Check role requirements
    if (requireRole && user) {
        const roleHierarchy = { viewer: 1, user: 2, admin: 3, system_admin: 4 };
        const userRoleLevel = roleHierarchy[user.role] || 0;
        const requiredRoleLevel = roleHierarchy[requireRole] || 0;

        if (userRoleLevel < requiredRoleLevel) {
            return (
                <div style={{
                    padding: '2rem',
                    textAlign: 'center',
                    color: '#dc2626',
                    backgroundColor: '#fef2f2',
                    border: '1px solid #fecaca',
                    borderRadius: '8px',
                    margin: '1rem',
                }}>
                    <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', marginBottom: '0.5rem' }}>
                        Access Denied
                    </h2>
                    <p>
                        You need {requireRole} role or higher to access this resource.
                        Your current role: {user.role}
                    </p>
                </div>
            );
        }
    }

    return <>{children}</>;
}
