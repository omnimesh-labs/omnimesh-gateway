'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from './AuthContext';
import { useState } from 'react';

export function Navigation() {
  const pathname = usePathname();
  const { isAuthenticated, user, logout } = useAuth();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const navItems = [
    { href: '/', label: 'Dashboard' },
    { href: '/servers', label: 'Server Management' },
    { href: '/policies', label: 'Policy Management' },
    { href: '/logs', label: 'Logging & Audit' },
  ];

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <nav style={{
      background: 'white',
      borderBottom: '1px solid #e5e7eb',
      padding: '1rem 0',
      marginBottom: '2rem'
    }}>
      <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '0 2rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Link 
            href="/"
            style={{
              fontSize: '1.5rem',
              fontWeight: 'bold',
              color: '#111827',
              textDecoration: 'none'
            }}
          >
            MCP Gateway
          </Link>
          
          <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
            {navItems.map(item => (
              <Link
                key={item.href}
                href={item.href}
                style={{
                  color: pathname === item.href ? '#3b82f6' : '#6b7280',
                  textDecoration: 'none',
                  fontSize: '0.875rem',
                  fontWeight: pathname === item.href ? '600' : '400',
                  padding: '0.5rem 0',
                  borderBottom: pathname === item.href ? '2px solid #3b82f6' : '2px solid transparent',
                  transition: 'color 0.2s'
                }}
              >
                {item.label}
              </Link>
            ))}
            
            {isAuthenticated && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginLeft: '1rem' }}>
                <div style={{
                  fontSize: '0.875rem',
                  color: '#6b7280',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                }}>
                  <span style={{
                    width: '8px',
                    height: '8px',
                    backgroundColor: '#10b981',
                    borderRadius: '50%',
                  }} />
                  {user?.email}
                  <span style={{
                    backgroundColor: user?.role === 'admin' ? '#dc2626' : user?.role === 'user' ? '#3b82f6' : '#6b7280',
                    color: 'white',
                    padding: '0.125rem 0.5rem',
                    borderRadius: '12px',
                    fontSize: '0.75rem',
                    fontWeight: '500',
                    textTransform: 'capitalize',
                  }}>
                    {user?.role}
                  </span>
                </div>
                <button
                  onClick={handleLogout}
                  disabled={isLoggingOut}
                  style={{
                    backgroundColor: 'transparent',
                    border: '1px solid #d1d5db',
                    color: '#6b7280',
                    padding: '0.5rem 1rem',
                    borderRadius: '4px',
                    fontSize: '0.875rem',
                    cursor: isLoggingOut ? 'not-allowed' : 'pointer',
                    transition: 'all 0.2s',
                  }}
                  onMouseOver={(e) => {
                    if (!isLoggingOut) {
                      (e.target as HTMLButtonElement).style.backgroundColor = '#f9fafb';
                      (e.target as HTMLButtonElement).style.borderColor = '#9ca3af';
                    }
                  }}
                  onMouseOut={(e) => {
                    if (!isLoggingOut) {
                      (e.target as HTMLButtonElement).style.backgroundColor = 'transparent';
                      (e.target as HTMLButtonElement).style.borderColor = '#d1d5db';
                    }
                  }}
                >
                  {isLoggingOut ? 'Signing out...' : 'Sign out'}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
