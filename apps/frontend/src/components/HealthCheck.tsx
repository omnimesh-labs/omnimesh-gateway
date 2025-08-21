'use client';

import { useState, useEffect } from 'react';

const API_BASE_URL = 'http://localhost:8080';

export function HealthCheck() {
  const [status, setStatus] = useState<'checking' | 'healthy' | 'error'>('checking');
  const [message, setMessage] = useState('');

  const checkHealth = async () => {
    try {
      setStatus('checking');
      const response = await fetch(`${API_BASE_URL}/health`, {
        mode: 'cors',
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setStatus('healthy');
        setMessage('Backend is running and accessible');
      } else {
        setStatus('error');
        setMessage(`HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error) {
      setStatus('error');
      if (error instanceof TypeError && error.message.includes('fetch')) {
        setMessage('Cannot connect to backend. Please ensure the backend server is running on port 8080.');
      } else {
        setMessage(error instanceof Error ? error.message : 'Unknown error');
      }
    }
  };

  useEffect(() => {
    checkHealth();
  }, []);

  const getStatusColor = () => {
    switch (status) {
      case 'healthy':
        return '#10b981';
      case 'error':
        return '#ef4444';
      case 'checking':
        return '#f59e0b';
      default:
        return '#6b7280';
    }
  };

  const getStatusIcon = () => {
    switch (status) {
      case 'healthy':
        return '✅';
      case 'error':
        return '❌';
      case 'checking':
        return '⏳';
      default:
        return '❓';
    }
  };

  return (
    <div style={{
      background: 'white',
      border: '1px solid #e5e7eb',
      borderRadius: '8px',
      padding: '1rem',
      marginBottom: '1rem'
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <span style={{ fontSize: '1rem' }}>{getStatusIcon()}</span>
          <span style={{
            fontSize: '0.875rem',
            fontWeight: '500',
            color: getStatusColor()
          }}>
            Backend Status: {status.charAt(0).toUpperCase() + status.slice(1)}
          </span>
        </div>
        
        <button
          onClick={checkHealth}
          disabled={status === 'checking'}
          style={{
            background: '#f3f4f6',
            color: '#374151',
            padding: '0.375rem 0.75rem',
            borderRadius: '4px',
            border: 'none',
            fontSize: '0.75rem',
            cursor: status === 'checking' ? 'not-allowed' : 'pointer',
            opacity: status === 'checking' ? 0.5 : 1,
            transition: 'background-color 0.2s'
          }}
          onMouseOver={(e) => {
            if (status !== 'checking') {
              e.currentTarget.style.background = '#e5e7eb';
            }
          }}
          onMouseOut={(e) => {
            if (status !== 'checking') {
              e.currentTarget.style.background = '#f3f4f6';
            }
          }}
        >
          {status === 'checking' ? 'Checking...' : 'Refresh'}
        </button>
      </div>
      
      {message && (
        <div style={{
          fontSize: '0.75rem',
          color: '#6b7280',
          marginTop: '0.5rem',
          padding: '0.5rem',
          background: '#f9fafb',
          borderRadius: '4px'
        }}>
          {message}
        </div>
      )}
    </div>
  );
}
