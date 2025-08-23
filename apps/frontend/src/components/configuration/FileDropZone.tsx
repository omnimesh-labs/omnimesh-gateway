'use client';

import { useState, useRef, DragEvent } from 'react';

interface FileDropZoneProps {
  onFileSelect: (file: File) => void;
  acceptedTypes?: string[];
  maxSize?: number; // in bytes
  disabled?: boolean;
}

export function FileDropZone({
  onFileSelect,
  acceptedTypes = ['.json', 'application/json'],
  maxSize = 10 * 1024 * 1024, // 10MB default
  disabled = false
}: FileDropZoneProps) {
  const [isDragOver, setIsDragOver] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const validateFile = (file: File): string | null => {
    if (maxSize && file.size > maxSize) {
      return `File size must be less than ${(maxSize / 1024 / 1024).toFixed(1)}MB`;
    }
    
    if (acceptedTypes.length > 0) {
      const isValidType = acceptedTypes.some(type => {
        if (type.startsWith('.')) {
          return file.name.toLowerCase().endsWith(type.toLowerCase());
        }
        return file.type === type;
      });
      
      if (!isValidType) {
        return `File must be one of: ${acceptedTypes.join(', ')}`;
      }
    }
    
    return null;
  };

  const handleFile = (file: File) => {
    const validationError = validateFile(file);
    if (validationError) {
      setError(validationError);
      return;
    }
    
    setError(null);
    onFileSelect(file);
  };

  const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    if (!disabled) {
      setIsDragOver(true);
    }
  };

  const handleDragLeave = () => {
    setIsDragOver(false);
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragOver(false);
    
    if (disabled) return;
    
    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
      handleFile(files[0]);
    }
  };

  const handleClick = () => {
    if (!disabled && fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length > 0) {
      handleFile(files[0]);
    }
  };

  return (
    <div style={{ width: '100%' }}>
      <div
        onClick={handleClick}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        style={{
          border: `2px dashed ${isDragOver ? '#3b82f6' : (error ? '#ef4444' : '#d1d5db')}`,
          borderRadius: '8px',
          padding: '3rem 1.5rem',
          textAlign: 'center',
          cursor: disabled ? 'not-allowed' : 'pointer',
          backgroundColor: isDragOver ? '#eff6ff' : (disabled ? '#f9fafb' : '#ffffff'),
          transition: 'all 0.2s ease-in-out',
          opacity: disabled ? 0.5 : 1
        }}
        onMouseEnter={(e) => {
          if (!disabled) {
            e.currentTarget.style.borderColor = '#9ca3af';
          }
        }}
        onMouseLeave={(e) => {
          if (!disabled) {
            e.currentTarget.style.borderColor = error ? '#ef4444' : '#d1d5db';
          }
        }}
      >
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '1rem'
        }}>
          <svg
            style={{
              width: '3rem',
              height: '3rem',
              color: isDragOver ? '#3b82f6' : (error ? '#ef4444' : '#9ca3af')
            }}
            stroke="currentColor"
            fill="none"
            viewBox="0 0 48 48"
          >
            <path
              d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3-3m-3 3l3 3m-3-3V8"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          
          <div>
            <div style={{
              fontSize: '0.875rem',
              color: '#374151',
              marginBottom: '0.25rem'
            }}>
              <span style={{
                fontWeight: '500',
                color: isDragOver ? '#3b82f6' : (error ? '#ef4444' : '#2563eb')
              }}>
                Click to upload
              </span>
              {!disabled && ' or drag and drop'}
            </div>
            
            <p style={{
              fontSize: '0.75rem',
              color: '#6b7280',
              margin: 0
            }}>
              JSON export files only
            </p>
            
            {maxSize && (
              <p style={{
                fontSize: '0.75rem',
                color: '#6b7280',
                margin: '0.25rem 0 0 0'
              }}>
                Max file size: {(maxSize / 1024 / 1024).toFixed(1)}MB
              </p>
            )}
          </div>
        </div>
        
        <input
          ref={fileInputRef}
          type="file"
          accept={acceptedTypes.join(',')}
          onChange={handleFileInputChange}
          disabled={disabled}
          style={{ display: 'none' }}
        />
      </div>
      
      {error && (
        <div style={{
          marginTop: '0.5rem',
          padding: '0.5rem',
          backgroundColor: '#fef2f2',
          border: '1px solid #fecaca',
          borderRadius: '4px',
          fontSize: '0.875rem',
          color: '#dc2626'
        }}>
          {error}
        </div>
      )}
    </div>
  );
}