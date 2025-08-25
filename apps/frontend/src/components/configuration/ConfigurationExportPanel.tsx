'use client';

import { useState } from 'react';
import { EntityTypeSelector } from './EntityTypeSelector';
import { ProgressBar } from './ProgressBar';

interface ExportOptions {
  entityTypes: string[];
  includeInactive: boolean;
  includeDependencies: boolean;
  tags: string[];
}

interface ConfigurationExportPanelProps {
  onExport: (options: ExportOptions) => Promise<void>;
  isExporting?: boolean;
  exportProgress?: number;
}

const DEFAULT_ENTITY_TYPES = [
  'servers',
  'virtual_servers',
  'tools',
  'prompts',
  'resources',
  'policies',
  'rate_limits'
];

export function ConfigurationExportPanel({
  onExport,
  isExporting = false,
  exportProgress = 0
}: ConfigurationExportPanelProps) {
  const [selectedTypes, setSelectedTypes] = useState<string[]>(DEFAULT_ENTITY_TYPES);
  const [includeInactive, setIncludeInactive] = useState(false);
  const [includeDependencies, setIncludeDependencies] = useState(true);
  const [tags, setTags] = useState('');

  const handleExportAll = async () => {
    const options: ExportOptions = {
      entityTypes: DEFAULT_ENTITY_TYPES,
      includeInactive,
      includeDependencies,
      tags: tags.split(',').map(t => t.trim()).filter(Boolean)
    };

    await onExport(options);
  };

  const handleExportSelected = async () => {
    if (selectedTypes.length === 0) {
      alert('Please select at least one entity type to export.');
      return;
    }

    const options: ExportOptions = {
      entityTypes: selectedTypes,
      includeInactive,
      includeDependencies,
      tags: tags.split(',').map(t => t.trim()).filter(Boolean)
    };

    await onExport(options);
  };

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
        📤 Export Configuration
      </h2>

      <div style={{
        backgroundColor: '#f9fafb',
        padding: '1rem',
        borderRadius: '8px',
        marginBottom: '2rem'
      }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
          gap: '1.5rem',
          marginBottom: '1rem'
        }}>
          <div>
            <EntityTypeSelector
              entityTypes={DEFAULT_ENTITY_TYPES}
              selectedTypes={selectedTypes}
              onChange={setSelectedTypes}
            />
          </div>

          <div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.5rem'
              }}>
                Filter Options
              </label>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <input
                    type="checkbox"
                    checked={includeInactive}
                    onChange={(e) => setIncludeInactive(e.target.checked)}
                    disabled={isExporting}
                    style={{ cursor: isExporting ? 'not-allowed' : 'pointer' }}
                  />
                  <span style={{
                    fontSize: '0.875rem',
                    color: '#6b7280'
                  }}>
                    Include Inactive
                  </span>
                </label>

                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <input
                    type="checkbox"
                    checked={includeDependencies}
                    onChange={(e) => setIncludeDependencies(e.target.checked)}
                    disabled={isExporting}
                    style={{ cursor: isExporting ? 'not-allowed' : 'pointer' }}
                  />
                  <span style={{
                    fontSize: '0.875rem',
                    color: '#6b7280'
                  }}>
                    Include Dependencies
                  </span>
                </label>
              </div>
            </div>

            <div>
              <label style={{
                display: 'block',
                fontSize: '0.875rem',
                fontWeight: '500',
                color: '#374151',
                marginBottom: '0.25rem'
              }}>
                Filter by Tags
              </label>
              <input
                type="text"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
                placeholder="production, api, staging"
                disabled={isExporting}
                style={{
                  width: '100%',
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '6px',
                  fontSize: '0.875rem',
                  backgroundColor: isExporting ? '#f9fafb' : '#ffffff',
                  color: isExporting ? '#9ca3af' : '#111827'
                }}
              />
              <p style={{
                fontSize: '0.75rem',
                color: '#6b7280',
                margin: '0.25rem 0 0 0'
              }}>
                Comma-separated tags
              </p>
            </div>
          </div>

          <div>
            <label style={{
              display: 'block',
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151',
              marginBottom: '0.5rem'
            }}>
              Export Actions
            </label>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              <button
                onClick={handleExportAll}
                disabled={isExporting}
                style={{
                  width: '100%',
                  backgroundColor: isExporting ? '#9ca3af' : '#3b82f6',
                  color: '#ffffff',
                  padding: '0.5rem 1rem',
                  borderRadius: '6px',
                  border: 'none',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  cursor: isExporting ? 'not-allowed' : 'pointer',
                  transition: 'background-color 0.2s'
                }}
                onMouseEnter={(e) => {
                  if (!isExporting) {
                    e.currentTarget.style.backgroundColor = '#2563eb';
                  }
                }}
                onMouseLeave={(e) => {
                  if (!isExporting) {
                    e.currentTarget.style.backgroundColor = '#3b82f6';
                  }
                }}
              >
                📥 Export All Configuration
              </button>

              <button
                onClick={handleExportSelected}
                disabled={isExporting || selectedTypes.length === 0}
                style={{
                  width: '100%',
                  backgroundColor: isExporting || selectedTypes.length === 0 ? '#9ca3af' : '#10b981',
                  color: '#ffffff',
                  padding: '0.5rem 1rem',
                  borderRadius: '6px',
                  border: 'none',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  cursor: isExporting || selectedTypes.length === 0 ? 'not-allowed' : 'pointer',
                  transition: 'background-color 0.2s'
                }}
                onMouseEnter={(e) => {
                  if (!isExporting && selectedTypes.length > 0) {
                    e.currentTarget.style.backgroundColor = '#059669';
                  }
                }}
                onMouseLeave={(e) => {
                  if (!isExporting && selectedTypes.length > 0) {
                    e.currentTarget.style.backgroundColor = '#10b981';
                  }
                }}
              >
                📋 Export Selected Types
              </button>
            </div>

            {isExporting && (
              <div style={{ marginTop: '1rem' }}>
                <div style={{
                  fontSize: '0.875rem',
                  color: '#374151',
                  marginBottom: '0.5rem'
                }}>
                  Exporting...
                </div>
                <ProgressBar
                  value={exportProgress}
                  color="blue"
                  showPercentage={false}
                />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
