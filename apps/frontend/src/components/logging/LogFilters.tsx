'use client';

import React, { useState } from 'react';
import { type LogQueryParams, type AuditQueryParams } from '@/lib/api';

interface LogFiltersProps {
  onFilterChange: (filters: LogQueryParams | AuditQueryParams) => void;
  type: 'logs' | 'audit';
  loading: boolean;
}

export function LogFilters({ onFilterChange, type, loading }: LogFiltersProps) {
  const [filters, setFilters] = useState<LogQueryParams | AuditQueryParams>(
    type === 'logs'
      ? { limit: 100, offset: 0 }
      : { limit: 100, offset: 0 }
  );
  const [showAdvanced, setShowAdvanced] = useState(false);

  const updateFilter = (key: string, value: string | number) => {
    const newFilters = { ...filters, [key]: value === '' ? undefined : value };
    setFilters(newFilters);
    onFilterChange(newFilters);
  };

  const clearFilters = () => {
    const clearedFilters = type === 'logs'
      ? { limit: 100, offset: 0 }
      : { limit: 100, offset: 0 };
    setFilters(clearedFilters);
    onFilterChange(clearedFilters);
  };

  const logFilters = filters as LogQueryParams;
  const auditFilters = filters as AuditQueryParams;

  return (
    <div style={{
      background: 'white',
      borderRadius: '8px',
      boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
      padding: '1.5rem',
      marginBottom: '1.5rem'
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: '1rem'
      }}>
        <h3 style={{ fontSize: '1rem', fontWeight: '500', color: '#374151' }}>
          Filter {type === 'logs' ? 'Logs' : 'Audit Trail'}
        </h3>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            onClick={() => setShowAdvanced(!showAdvanced)}
            style={{
              background: 'none',
              color: '#3b82f6',
              border: 'none',
              fontSize: '0.875rem',
              cursor: 'pointer',
              textDecoration: 'underline'
            }}
          >
            {showAdvanced ? 'Hide' : 'Show'} Advanced
          </button>
          <button
            onClick={clearFilters}
            style={{
              background: '#f3f4f6',
              color: '#374151',
              padding: '0.25rem 0.75rem',
              borderRadius: '4px',
              border: 'none',
              fontSize: '0.875rem',
              cursor: 'pointer'
            }}
          >
            Clear All
          </button>
        </div>
      </div>

      {/* Basic Filters */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
        gap: '1rem',
        marginBottom: showAdvanced ? '1rem' : '0'
      }}>
        {type === 'logs' ? (
          <>
            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                Log Level
              </label>
              <select
                value={logFilters.level || ''}
                onChange={(e) => updateFilter('level', e.target.value)}
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem'
                }}
              >
                <option value="">All Levels</option>
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="warn">Warning</option>
                <option value="error">Error</option>
              </select>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                User ID
              </label>
              <input
                type="text"
                value={logFilters.user_id || ''}
                onChange={(e) => updateFilter('user_id', e.target.value)}
                placeholder="Filter by user..."
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                Search
              </label>
              <input
                type="text"
                value={logFilters.search || ''}
                onChange={(e) => updateFilter('search', e.target.value)}
                placeholder="Search logs..."
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem'
                }}
              />
            </div>
          </>
        ) : (
          <>
            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                Resource Type
              </label>
              <select
                value={auditFilters.resource_type || ''}
                onChange={(e) => updateFilter('resource_type', e.target.value)}
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem'
                }}
              >
                <option value="">All Resources</option>
                <option value="server">Server</option>
                <option value="virtual-server">Virtual Server</option>
                <option value="user">User</option>
                <option value="session">Session</option>
              </select>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                Action
              </label>
              <select
                value={auditFilters.action || ''}
                onChange={(e) => updateFilter('action', e.target.value)}
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem'
                }}
              >
                <option value="">All Actions</option>
                <option value="create">Create</option>
                <option value="update">Update</option>
                <option value="delete">Delete</option>
                <option value="register">Register</option>
                <option value="unregister">Unregister</option>
              </select>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                Actor ID
              </label>
              <input
                type="text"
                value={auditFilters.actor_id || ''}
                onChange={(e) => updateFilter('actor_id', e.target.value)}
                placeholder="Filter by actor..."
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '4px',
                  fontSize: '0.875rem'
                }}
              />
            </div>
          </>
        )}

        <div>
          <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
            Results Per Page
          </label>
          <select
            value={(filters.limit || 100).toString()}
            onChange={(e) => updateFilter('limit', parseInt(e.target.value))}
            style={{
              width: '100%',
              padding: '0.5rem',
              border: '1px solid #d1d5db',
              borderRadius: '4px',
              fontSize: '0.875rem'
            }}
          >
            <option value="25">25</option>
            <option value="50">50</option>
            <option value="100">100</option>
            <option value="200">200</option>
          </select>
        </div>
      </div>

      {/* Advanced Filters */}
      {showAdvanced && type === 'logs' && (
        <div style={{
          borderTop: '1px solid #e5e7eb',
          paddingTop: '1rem',
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: '1rem'
        }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
              Start Time
            </label>
            <input
              type="datetime-local"
              value={logFilters.start_time || ''}
              onChange={(e) => updateFilter('start_time', e.target.value ? new Date(e.target.value).toISOString() : '')}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '4px',
                fontSize: '0.875rem'
              }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
              End Time
            </label>
            <input
              type="datetime-local"
              value={logFilters.end_time || ''}
              onChange={(e) => updateFilter('end_time', e.target.value ? new Date(e.target.value).toISOString() : '')}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '4px',
                fontSize: '0.875rem'
              }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
              RPC Method
            </label>
            <input
              type="text"
              value={logFilters.method || ''}
              onChange={(e) => updateFilter('method', e.target.value)}
              placeholder="e.g. tools/list"
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '4px',
                fontSize: '0.875rem'
              }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
              Path
            </label>
            <input
              type="text"
              value={logFilters.path || ''}
              onChange={(e) => updateFilter('path', e.target.value)}
              placeholder="API path filter"
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '4px',
                fontSize: '0.875rem'
              }}
            />
          </div>
        </div>
      )}

      {loading && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          marginTop: '1rem',
          color: '#6b7280',
          fontSize: '0.875rem'
        }}>
          <div style={{
            width: '16px',
            height: '16px',
            border: '2px solid #e5e7eb',
            borderTop: '2px solid #3b82f6',
            borderRadius: '50%',
            animation: 'spin 1s linear infinite'
          }}></div>
          Loading...
        </div>
      )}

      <style jsx>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}
