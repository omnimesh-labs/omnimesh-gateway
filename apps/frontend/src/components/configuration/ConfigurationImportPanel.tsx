'use client';

import { useState } from 'react';
import { FileDropZone } from './FileDropZone';
import { ConflictStrategySelector } from './ConflictStrategySelector';

interface ImportOptions {
  conflictStrategy: string;
  dryRun: boolean;
  rekeySecret?: string;
}

interface ValidationResult {
  valid: boolean;
  errors: Array<{
    code: string;
    message: string;
    entityType?: string;
    entityName?: string;
  }>;
  warnings: Array<{
    code: string;
    message: string;
    entityType?: string;
    entityName?: string;
  }>;
  entityCounts: Record<string, number>;
  conflicts: Array<{
    entityType: string;
    entityName: string;
    conflictType: string;
    suggestion?: string;
  }>;
}

interface ConfigurationImportPanelProps {
  onImport: (file: File, options: ImportOptions) => Promise<void>;
  onValidate: (file: File) => Promise<ValidationResult>;
  isImporting?: boolean;
  isValidating?: boolean;
}

export function ConfigurationImportPanel({
  onImport,
  onValidate,
  isImporting = false,
  isValidating = false
}: ConfigurationImportPanelProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [conflictStrategy, setConflictStrategy] = useState('update');
  const [dryRun, setDryRun] = useState(false);
  const [rekeySecret, setRekeySecret] = useState('');
  const [validation, setValidation] = useState<ValidationResult | null>(null);

  const handleFileSelect = (file: File) => {
    setSelectedFile(file);
    setValidation(null);
  };

  const handleValidate = async () => {
    if (!selectedFile) return;

    try {
      const result = await onValidate(selectedFile);
      setValidation(result);
    } catch (error) {
      console.error('Validation failed:', error);
      setValidation({
        valid: false,
        errors: [{
          code: 'VALIDATION_ERROR',
          message: error instanceof Error ? error.message : 'Validation failed'
        }],
        warnings: [],
        entityCounts: {},
        conflicts: []
      });
    }
  };

  const handleImport = async () => {
    if (!selectedFile) return;

    const options: ImportOptions = {
      conflictStrategy,
      dryRun,
      rekeySecret: rekeySecret.trim() || undefined
    };

    await onImport(selectedFile, options);
  };

  const canImport = selectedFile && !isImporting && !isValidating;
  const canValidate = selectedFile && !isValidating && !isImporting;

  return (
    <div style={{
      backgroundColor: '#ffffff',
      border: '1px solid #e5e7eb',
      borderRadius: '8px',
      padding: '1.5rem'
    }}>
      <h2 style={{
        fontSize: '1.125rem',
        fontWeight: '500',
        color: '#111827',
        margin: '0 0 1.5rem 0'
      }}>
        📥 Import Configuration
      </h2>

      <div style={{
        backgroundColor: '#f9fafb',
        padding: '1rem',
        borderRadius: '8px'
      }}>
        <div style={{ marginBottom: '1.5rem' }}>
          <label style={{
            display: 'block',
            fontSize: '0.875rem',
            fontWeight: '500',
            color: '#374151',
            marginBottom: '0.5rem'
          }}>
            Import File
          </label>

          <FileDropZone
            onFileSelect={handleFileSelect}
            disabled={isImporting || isValidating}
          />

          {selectedFile && (
            <div style={{
              marginTop: '0.5rem',
              padding: '0.5rem',
              backgroundColor: '#eff6ff',
              border: '1px solid #bfdbfe',
              borderRadius: '4px',
              fontSize: '0.875rem',
              color: '#1e40af'
            }}>
              Selected: {selectedFile.name} ({(selectedFile.size / 1024).toFixed(1)} KB)
            </div>
          )}
        </div>

        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: '1.5rem',
          marginBottom: '1.5rem'
        }}>
          <div>
            <ConflictStrategySelector
              value={conflictStrategy}
              onChange={setConflictStrategy}
              disabled={isImporting || isValidating}
            />
          </div>

          <div>
            <label style={{
              display: 'block',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              marginBottom: '0.5rem'
            }}>
              Options
            </label>

            <div style={{ marginBottom: '0.75rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={dryRun}
                  onChange={(e) => setDryRun(e.target.checked)}
                  disabled={isImporting || isValidating}
                  style={{ cursor: isImporting || isValidating ? 'not-allowed' : 'pointer' }}
                />
                <span style={{
                  fontSize: '0.875rem',
                  color: '#6b7280'
                }}>
                  Dry Run (validate only)
                </span>
              </label>
            </div>

            <div>
              <input
                type="password"
                value={rekeySecret}
                onChange={(e) => setRekeySecret(e.target.value)}
                placeholder="New encryption secret (optional)"
                disabled={isImporting || isValidating}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '6px',
                  fontSize: '0.875rem',
                  backgroundColor: isImporting || isValidating ? '#f9fafb' : '#ffffff',
                  color: isImporting || isValidating ? '#9ca3af' : '#111827'
                }}
              />
              <p style={{
                fontSize: '0.75rem',
                color: '#6b7280',
                margin: '0.25rem 0 0 0'
              }}>
                For cross-environment imports
              </p>
            </div>
          </div>
        </div>

        <div style={{
          display: 'flex',
          gap: '1rem',
          flexWrap: 'wrap'
        }}>
          <button
            onClick={handleValidate}
            disabled={!canValidate}
            style={{
              backgroundColor: !canValidate ? '#9ca3af' : '#f59e0b',
              color: '#ffffff',
              padding: '0.5rem 1rem',
              borderRadius: '6px',
              border: 'none',
              fontSize: '0.875rem',
              fontWeight: '500',
              cursor: !canValidate ? 'not-allowed' : 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseEnter={(e) => {
              if (canValidate) {
                e.currentTarget.style.backgroundColor = '#d97706';
              }
            }}
            onMouseLeave={(e) => {
              if (canValidate) {
                e.currentTarget.style.backgroundColor = '#f59e0b';
              }
            }}
          >
            🔍 {isValidating ? 'Validating...' : 'Validate Import'}
          </button>

          <button
            onClick={handleImport}
            disabled={!canImport}
            style={{
              backgroundColor: !canImport ? '#9ca3af' : '#ef4444',
              color: '#ffffff',
              padding: '0.5rem 1rem',
              borderRadius: '6px',
              border: 'none',
              fontSize: '0.875rem',
              fontWeight: '500',
              cursor: !canImport ? 'not-allowed' : 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseEnter={(e) => {
              if (canImport) {
                e.currentTarget.style.backgroundColor = '#dc2626';
              }
            }}
            onMouseLeave={(e) => {
              if (canImport) {
                e.currentTarget.style.backgroundColor = '#ef4444';
              }
            }}
          >
            ⚡ {isImporting ? 'Importing...' : 'Execute Import'}
          </button>
        </div>
      </div>

      {/* Validation Results */}
      {validation && (
        <div style={{
          marginTop: '1.5rem',
          padding: '1rem',
          backgroundColor: validation.valid ? '#f0fdf4' : '#fef2f2',
          border: `1px solid ${validation.valid ? '#bbf7d0' : '#fecaca'}`,
          borderRadius: '6px'
        }}>
          <h4 style={{
            fontSize: '0.875rem',
            fontWeight: '500',
            color: validation.valid ? '#166534' : '#dc2626',
            margin: '0 0 0.75rem 0'
          }}>
            {validation.valid ? '✅ Validation Passed' : '❌ Validation Failed'}
          </h4>

          {validation.errors.length > 0 && (
            <div style={{ marginBottom: '0.75rem' }}>
              <h5 style={{
                fontSize: '0.8125rem',
                fontWeight: '500',
                color: '#dc2626',
                margin: '0 0 0.5rem 0'
              }}>
                Errors:
              </h5>
              {validation.errors.map((error, index) => (
                <div key={index} style={{
                  fontSize: '0.8125rem',
                  color: '#dc2626',
                  marginBottom: '0.25rem'
                }}>
                  • {error.message}
                  {error.entityType && ` (${error.entityType}${error.entityName ? `: ${error.entityName}` : ''})`}
                </div>
              ))}
            </div>
          )}

          {validation.warnings.length > 0 && (
            <div style={{ marginBottom: '0.75rem' }}>
              <h5 style={{
                fontSize: '0.8125rem',
                fontWeight: '500',
                color: '#f59e0b',
                margin: '0 0 0.5rem 0'
              }}>
                Warnings:
              </h5>
              {validation.warnings.map((warning, index) => (
                <div key={index} style={{
                  fontSize: '0.8125rem',
                  color: '#f59e0b',
                  marginBottom: '0.25rem'
                }}>
                  • {warning.message}
                  {warning.entityType && ` (${warning.entityType}${warning.entityName ? `: ${warning.entityName}` : ''})`}
                </div>
              ))}
            </div>
          )}

          {validation.conflicts.length > 0 && (
            <div style={{ marginBottom: '0.75rem' }}>
              <h5 style={{
                fontSize: '0.8125rem',
                fontWeight: '500',
                color: '#f59e0b',
                margin: '0 0 0.5rem 0'
              }}>
                Conflicts:
              </h5>
              {validation.conflicts.map((conflict, index) => (
                <div key={index} style={{
                  fontSize: '0.8125rem',
                  color: '#f59e0b',
                  marginBottom: '0.25rem'
                }}>
                  • {conflict.entityType}: {conflict.entityName} ({conflict.conflictType})
                  {conflict.suggestion && ` - ${conflict.suggestion}`}
                </div>
              ))}
            </div>
          )}

          <div style={{
            fontSize: '0.8125rem',
            color: '#6b7280'
          }}>
            Entity counts: {Object.entries(validation.entityCounts)
              .map(([type, count]) => `${type}: ${count}`)
              .join(', ')}
          </div>
        </div>
      )}
    </div>
  );
}
