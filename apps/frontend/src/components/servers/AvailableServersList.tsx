'use client';

import { useState, useEffect, useCallback } from 'react';
import { discoveryApi, type MCPPackage, type MCPServer } from '@/lib/api';
import { RegisterServerModal } from './RegisterServerModal';

interface AvailableServersListProps {
  onRegister: (server: MCPServer) => void;
}

export function AvailableServersList({ onRegister }: AvailableServersListProps) {
  const [packages, setPackages] = useState<MCPPackage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedPackage, setSelectedPackage] = useState<MCPPackage | null>(null);
  const [showRegisterModal, setShowRegisterModal] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);

  const pageSize = 20;

  const loadPackages = useCallback(async (search = '', reset = false) => {
    try {
      setLoading(true);
      const currentOffset = reset ? 0 : offset;
      const response = await discoveryApi.searchPackages(search, currentOffset, pageSize);

      const packageList = Object.values(response.results);

      if (reset) {
        setPackages(packageList);
        setOffset(pageSize);
      } else {
        setPackages(prev => [...prev, ...packageList]);
        setOffset(prev => prev + pageSize);
      }

      setHasMore(response.hasMore);
      setTotal(response.total);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load packages');
    } finally {
      setLoading(false);
    }
  }, [offset]);

  useEffect(() => {
    loadPackages('', true);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setOffset(0);
    loadPackages(query, true);
  };

  const handleLoadMore = () => {
    if (!loading && hasMore) {
      loadPackages(searchQuery);
    }
  };

  const handleRegisterPackage = (pkg: MCPPackage) => {
    setSelectedPackage(pkg);
    setShowRegisterModal(true);
  };

  const handleModalClose = () => {
    setShowRegisterModal(false);
    setSelectedPackage(null);
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  return (
    <div>
      {/* Search Bar */}
      <div style={{ marginBottom: '1.5rem' }}>
        <div style={{ position: 'relative', maxWidth: '400px' }}>
          <input
            type="text"
            placeholder="Search MCP packages..."
            value={searchQuery}
            onChange={(e) => handleSearch(e.target.value)}
            style={{
              width: '100%',
              padding: '0.75rem 1rem',
              border: '1px solid #d1d5db',
              borderRadius: '8px',
              fontSize: '0.875rem',
              outline: 'none',
              transition: 'border-color 0.2s'
            }}
            onFocus={(e) => e.currentTarget.style.borderColor = '#3b82f6'}
            onBlur={(e) => e.currentTarget.style.borderColor = '#d1d5db'}
          />
          <div style={{
            position: 'absolute',
            right: '0.75rem',
            top: '50%',
            transform: 'translateY(-50%)',
            color: '#6b7280',
            fontSize: '0.875rem'
          }}>
            🔍
          </div>
        </div>
        {total > 0 && (
          <div style={{ marginTop: '0.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
            Found {total} package{total !== 1 ? 's' : ''}
          </div>
        )}
      </div>

      {error && (
        <div style={{
          background: '#fef2f2',
          border: '1px solid #fecaca',
          borderRadius: '8px',
          padding: '1rem',
          marginBottom: '1.5rem',
          color: '#dc2626'
        }}>
          {error}
        </div>
      )}

      {/* Package List */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))',
        gap: '1.5rem',
        marginBottom: '2rem'
      }}>
        {packages.map((pkg, index) => (
          <div
            key={`${pkg.name}-${index}`}
            style={{
              background: 'white',
              borderRadius: '8px',
              boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
              padding: '1.5rem',
              border: '1px solid #e5e7eb',
              transition: 'box-shadow 0.2s, border-color 0.2s'
            }}
            onMouseOver={(e) => {
              e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
              e.currentTarget.style.borderColor = '#d1d5db';
            }}
            onMouseOut={(e) => {
              e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.1)';
              e.currentTarget.style.borderColor = '#e5e7eb';
            }}
          >
            <div style={{ marginBottom: '1rem' }}>
              <h3 style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                color: '#111827',
                marginBottom: '0.5rem'
              }}>
                {pkg.name}
              </h3>
              <p style={{
                color: '#6b7280',
                fontSize: '0.875rem',
                lineHeight: '1.5',
                marginBottom: '0.75rem'
              }}>
                {pkg.description}
              </p>
            </div>

            {/* Package Stats */}
            <div style={{
              display: 'flex',
              gap: '1rem',
              marginBottom: '1rem',
              fontSize: '0.75rem',
              color: '#6b7280'
            }}>
              {pkg.github_stars > 0 && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                  ⭐ {formatNumber(pkg.github_stars)}
                </div>
              )}
              {pkg.package_download_count > 0 && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                  📦 {formatNumber(pkg.package_download_count)}
                </div>
              )}
            </div>

            {/* Command Preview */}
            {pkg.command && (
              <div style={{
                background: '#f3f4f6',
                borderRadius: '4px',
                padding: '0.5rem',
                marginBottom: '1rem',
                fontSize: '0.75rem',
                fontFamily: 'monospace',
                color: '#374151',
                overflowX: 'auto'
              }}>
                {pkg.command} {pkg.args.join(' ')}
              </div>
            )}

            {/* Package Registry */}
            {pkg.package_registry && (
              <div style={{
                fontSize: '0.75rem',
                color: '#6b7280',
                marginBottom: '1rem'
              }}>
                Registry: {pkg.package_registry}
              </div>
            )}

            {/* Action Buttons */}
            <div style={{ display: 'flex', gap: '0.5rem', marginTop: 'auto' }}>
              <button
                onClick={() => handleRegisterPackage(pkg)}
                style={{
                  flex: 1,
                  background: '#3b82f6',
                  color: 'white',
                  padding: '0.5rem 1rem',
                  borderRadius: '6px',
                  border: 'none',
                  fontSize: '0.875rem',
                  cursor: 'pointer',
                  transition: 'background-color 0.2s'
                }}
                onMouseOver={(e) => e.currentTarget.style.background = '#2563eb'}
                onMouseOut={(e) => e.currentTarget.style.background = '#3b82f6'}
              >
                Register
              </button>
              {pkg.githubUrl && (
                <a
                  href={pkg.githubUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{
                    background: '#f3f4f6',
                    color: '#374151',
                    padding: '0.5rem 1rem',
                    borderRadius: '6px',
                    border: 'none',
                    fontSize: '0.875rem',
                    textDecoration: 'none',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    transition: 'background-color 0.2s'
                  }}
                  onMouseOver={(e) => e.currentTarget.style.background = '#e5e7eb'}
                  onMouseOut={(e) => e.currentTarget.style.background = '#f3f4f6'}
                >
                  GitHub
                </a>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Loading State */}
      {loading && packages.length === 0 && (
        <div style={{
          background: 'white',
          borderRadius: '8px',
          boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
          padding: '2rem',
          textAlign: 'center'
        }}>
          <div style={{ color: '#6b7280' }}>Loading available packages...</div>
        </div>
      )}

      {/* Empty State */}
      {!loading && packages.length === 0 && !error && (
        <div style={{
          background: 'white',
          borderRadius: '8px',
          boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
          padding: '2rem',
          textAlign: 'center'
        }}>
          <div style={{ fontSize: '1.125rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
            No packages found
          </div>
          <div style={{ color: '#6b7280' }}>
            {searchQuery ? `No packages match "${searchQuery}"` : 'No packages available at the moment'}
          </div>
        </div>
      )}

      {/* Load More Button */}
      {hasMore && !loading && (
        <div style={{ textAlign: 'center', marginTop: '2rem' }}>
          <button
            onClick={handleLoadMore}
            style={{
              background: '#f3f4f6',
              color: '#374151',
              padding: '0.75rem 1.5rem',
              borderRadius: '8px',
              border: 'none',
              fontSize: '0.875rem',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => e.currentTarget.style.background = '#e5e7eb'}
            onMouseOut={(e) => e.currentTarget.style.background = '#f3f4f6'}
          >
            Load More
          </button>
        </div>
      )}

      {/* Loading More */}
      {loading && packages.length > 0 && (
        <div style={{ textAlign: 'center', marginTop: '2rem', color: '#6b7280' }}>
          Loading more packages...
        </div>
      )}

      {/* Register Modal */}
      {showRegisterModal && selectedPackage && (
        <RegisterServerModal
          onClose={handleModalClose}
          onRegister={onRegister}
          prefilledData={{
            name: selectedPackage.name,
            description: selectedPackage.description,
            protocol: 'stdio',
            command: selectedPackage.command,
            args: selectedPackage.args,
            environment: selectedPackage.envs
          }}
        />
      )}
    </div>
  );
}
