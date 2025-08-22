'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from './AuthContext';
import { ProfileDropdown } from './ProfileDropdown';

export function Navigation() {
  const pathname = usePathname();
  const { isAuthenticated } = useAuth();

  const navItems = [
    { href: '/', label: 'Dashboard' },
    { href: '/servers', label: 'Server Management' },
    { href: '/policies', label: 'Policy Management' },
    { href: '/logs', label: 'Logging & Audit' },
  ];


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
              <div style={{ marginLeft: '1rem' }}>
                <ProfileDropdown />
              </div>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
