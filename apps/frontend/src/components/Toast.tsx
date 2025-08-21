'use client';

import { useState, useEffect } from 'react';

export interface ToastProps {
  message: string;
  type?: 'success' | 'error' | 'info' | 'warning';
  duration?: number;
  onClose: () => void;
}

export function Toast({ message, type = 'info', duration = 5000, onClose }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose();
    }, duration);

    return () => clearTimeout(timer);
  }, [duration, onClose]);

  const getToastStyles = () => {
    const baseStyles = {
      position: 'fixed' as const,
      top: '1rem',
      right: '1rem',
      background: 'white',
      border: '1px solid',
      borderRadius: '8px',
      padding: '1rem 1.5rem',
      boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
      zIndex: 1000,
      display: 'flex',
      alignItems: 'center',
      gap: '0.75rem',
      maxWidth: '400px',
      animation: 'slideIn 0.3s ease-out',
    };

    const typeStyles = {
      success: {
        borderColor: '#10b981',
        background: '#f0fdf4',
        color: '#065f46',
      },
      error: {
        borderColor: '#ef4444',
        background: '#fef2f2',
        color: '#dc2626',
      },
      warning: {
        borderColor: '#f59e0b',
        background: '#fffbeb',
        color: '#d97706',
      },
      info: {
        borderColor: '#3b82f6',
        background: '#eff6ff',
        color: '#1d4ed8',
      },
    };

    return { ...baseStyles, ...typeStyles[type] };
  };

  const getIcon = () => {
    switch (type) {
      case 'success':
        return '✅';
      case 'error':
        return '❌';
      case 'warning':
        return '⚠️';
      case 'info':
        return 'ℹ️';
      default:
        return 'ℹ️';
    }
  };

  return (
    <>
      <style jsx>{`
        @keyframes slideIn {
          from {
            transform: translateX(100%);
            opacity: 0;
          }
          to {
            transform: translateX(0);
            opacity: 1;
          }
        }
      `}</style>
      <div style={getToastStyles()}>
        <span style={{ fontSize: '1.25rem' }}>{getIcon()}</span>
        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: '500', fontSize: '0.875rem' }}>
            {message}
          </div>
        </div>
        <button
          onClick={onClose}
          style={{
            background: 'none',
            border: 'none',
            color: 'inherit',
            cursor: 'pointer',
            fontSize: '1.25rem',
            padding: '0.25rem',
            borderRadius: '4px',
            opacity: 0.7,
            transition: 'opacity 0.2s',
          }}
          onMouseOver={(e) => e.currentTarget.style.opacity = '1'}
          onMouseOut={(e) => e.currentTarget.style.opacity = '0.7'}
        >
          ×
        </button>
      </div>
    </>
  );
}

// Toast hook for managing multiple toasts
export function useToast() {
  const [toasts, setToasts] = useState<Array<{ id: string; props: Omit<ToastProps, 'onClose'> }>>([]);

  const showToast = (props: Omit<ToastProps, 'onClose'>) => {
    const id = Math.random().toString(36).substr(2, 9);
    setToasts(prev => [...prev, { id, props }]);
  };

  const removeToast = (id: string) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  };

  const ToastContainer = () => (
    <>
      {toasts.map(({ id, props }) => (
        <Toast
          key={id}
          {...props}
          onClose={() => removeToast(id)}
        />
      ))}
    </>
  );

  return {
    showToast,
    ToastContainer,
    success: (message: string) => showToast({ message, type: 'success' }),
    error: (message: string) => showToast({ message, type: 'error' }),
    warning: (message: string) => showToast({ message, type: 'warning' }),
    info: (message: string) => showToast({ message, type: 'info' }),
  };
}
