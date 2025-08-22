'use client';

import { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { useAuth } from './AuthContext';

export function ProfileDropdown() {
  const { user, logout, isAuthenticated } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // Close dropdown on escape key
  useEffect(() => {
    function handleEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }

    document.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('keydown', handleEscape);
    };
  }, []);

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      await logout();
      setIsOpen(false);
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      setIsLoggingOut(false);
    }
  };

  if (!isAuthenticated || !user) {
    return null;
  }

  return (
    <div style={{ position: 'relative' }} ref={dropdownRef}>
      {/* Profile Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          backgroundColor: 'transparent',
          border: '1px solid #d1d5db',
          borderRadius: '6px',
          padding: '0.5rem 0.75rem',
          cursor: 'pointer',
          transition: 'all 0.2s',
          fontSize: '0.875rem',
          color: '#374151',
        }}
        onMouseOver={(e) => {
          (e.target as HTMLButtonElement).style.backgroundColor = '#f9fafb';
          (e.target as HTMLButtonElement).style.borderColor = '#9ca3af';
        }}
        onMouseOut={(e) => {
          (e.target as HTMLButtonElement).style.backgroundColor = 'transparent';
          (e.target as HTMLButtonElement).style.borderColor = '#d1d5db';
        }}
      >
        {/* Avatar */}
        <div style={{
          width: '24px',
          height: '24px',
          backgroundColor: user?.role === 'admin' ? '#dc2626' : user?.role === 'user' ? '#3b82f6' : '#6b7280',
          borderRadius: '50%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'white',
          fontSize: '0.75rem',
          fontWeight: '600',
        }}>
          {user.email?.charAt(0).toUpperCase()}
        </div>
        
        {/* User Info */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', minWidth: 0 }}>
          <span style={{
            fontWeight: '500',
            fontSize: '0.875rem',
            color: '#111827',
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            maxWidth: '120px',
          }}>
            {user.email}
          </span>
        </div>

        {/* Dropdown Arrow */}
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          style={{
            transform: isOpen ? 'rotate(180deg)' : 'rotate(0deg)',
            transition: 'transform 0.2s',
            color: '#6b7280',
          }}
        >
          <path
            d="M3 4.5L6 7.5L9 4.5"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div style={{
          position: 'absolute',
          top: '100%',
          right: '0',
          marginTop: '0.5rem',
          backgroundColor: 'white',
          border: '1px solid #e5e7eb',
          borderRadius: '8px',
          boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
          minWidth: '200px',
          zIndex: 50,
          animation: 'slideDown 0.15s ease-out',
        }}>
          {/* User Info Header */}
          <div style={{
            padding: '0.75rem 1rem',
            borderBottom: '1px solid #f3f4f6',
          }}>
            <div style={{
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#111827',
              marginBottom: '0.25rem',
            }}>
              {user.email}
            </div>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
            }}>
              <span style={{
                backgroundColor: user?.role === 'admin' ? '#dc2626' : user?.role === 'user' ? '#3b82f6' : '#6b7280',
                color: 'white',
                padding: '0.125rem 0.5rem',
                borderRadius: '12px',
                fontSize: '0.75rem',
                fontWeight: '500',
                textTransform: 'capitalize',
              }}>
                {user.role}
              </span>
              <span style={{
                width: '6px',
                height: '6px',
                backgroundColor: '#10b981',
                borderRadius: '50%',
              }} />
              <span style={{
                fontSize: '0.75rem',
                color: '#6b7280',
              }}>
                Online
              </span>
            </div>
          </div>

          {/* Menu Items */}
          <div style={{ padding: '0.5rem 0' }}>
            <Link
              href="/profile"
              onClick={() => setIsOpen(false)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                padding: '0.5rem 1rem',
                color: '#374151',
                textDecoration: 'none',
                fontSize: '0.875rem',
                transition: 'background-color 0.15s',
                width: '100%',
              }}
              onMouseOver={(e) => {
                (e.target as HTMLAnchorElement).style.backgroundColor = '#f9fafb';
              }}
              onMouseOut={(e) => {
                (e.target as HTMLAnchorElement).style.backgroundColor = 'transparent';
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                <circle cx="12" cy="7" r="4" />
              </svg>
              View Profile
            </Link>

            <Link
              href="/profile/settings"
              onClick={() => setIsOpen(false)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                padding: '0.5rem 1rem',
                color: '#374151',
                textDecoration: 'none',
                fontSize: '0.875rem',
                transition: 'background-color 0.15s',
                width: '100%',
              }}
              onMouseOver={(e) => {
                (e.target as HTMLAnchorElement).style.backgroundColor = '#f9fafb';
              }}
              onMouseOut={(e) => {
                (e.target as HTMLAnchorElement).style.backgroundColor = 'transparent';
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="3" />
                <path d="M12 1v6m0 6v6m11-7h-6m-6 0H1m11-7a4 4 0 1 1-8 0 4 4 0 0 1 8 0z" />
              </svg>
              Account Settings
            </Link>

            <Link
              href="/profile/api-keys"
              onClick={() => setIsOpen(false)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                padding: '0.5rem 1rem',
                color: '#374151',
                textDecoration: 'none',
                fontSize: '0.875rem',
                transition: 'background-color 0.15s',
                width: '100%',
              }}
              onMouseOver={(e) => {
                (e.target as HTMLAnchorElement).style.backgroundColor = '#f9fafb';
              }}
              onMouseOut={(e) => {
                (e.target as HTMLAnchorElement).style.backgroundColor = 'transparent';
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <circle cx="12" cy="16" r="1" />
                <path d="M7 11V7a5 5 0 0 1 10 0v4" />
              </svg>
              API Keys
            </Link>
          </div>

          {/* Divider */}
          <div style={{
            height: '1px',
            backgroundColor: '#f3f4f6',
            margin: '0.5rem 0',
          }} />

          {/* Logout */}
          <div style={{ padding: '0.5rem 0 0.5rem 0' }}>
            <button
              onClick={handleLogout}
              disabled={isLoggingOut}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                padding: '0.5rem 1rem',
                color: '#dc2626',
                backgroundColor: 'transparent',
                border: 'none',
                fontSize: '0.875rem',
                cursor: isLoggingOut ? 'not-allowed' : 'pointer',
                transition: 'background-color 0.15s',
                width: '100%',
                textAlign: 'left',
              }}
              onMouseOver={(e) => {
                if (!isLoggingOut) {
                  (e.target as HTMLButtonElement).style.backgroundColor = '#fef2f2';
                }
              }}
              onMouseOut={(e) => {
                if (!isLoggingOut) {
                  (e.target as HTMLButtonElement).style.backgroundColor = 'transparent';
                }
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16,17 21,12 16,7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
              {isLoggingOut ? 'Signing out...' : 'Sign Out'}
            </button>
          </div>
        </div>
      )}

      {/* CSS Animation for dropdown */}
      <style jsx>{`
        @keyframes slideDown {
          from {
            opacity: 0;
            transform: translateY(-10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
}