'use client';

import { useState, useCallback } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { Toast } from '@/components/Toast';
import { ConfigurationExportPanel } from '@/components/configuration/ConfigurationExportPanel';
import { ConfigurationImportPanel } from '@/components/configuration/ConfigurationImportPanel';
import { ImportProgressTracker } from '@/components/configuration/ImportProgressTracker';

interface ToastState {
  message: string;
  type: 'success' | 'error';
  show: boolean;
}

interface ImportProgress {
  status: 'pending' | 'running' | 'completed' | 'failed' | 'partial' | 'validating';
  total: number;
  processed: number;
  created: number;
  updated: number;
  skipped: number;
  failed: number;
  currentItem?: string;
}

interface ExportOptions {
  entityTypes: string[];
  includeInactive: boolean;
  includeDependencies: boolean;
  tags: string[];
}

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

export default function ConfigurationPage() {
  const [activeTab, setActiveTab] = useState('export');
  const [isExporting, setIsExporting] = useState(false);
  const [exportProgress, setExportProgress] = useState(0);
  const [isImporting, setIsImporting] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [importProgress, setImportProgress] = useState<ImportProgress>({
    status: 'pending',
    total: 0,
    processed: 0,
    created: 0,
    updated: 0,
    skipped: 0,
    failed: 0
  });
  const [showImportProgress, setShowImportProgress] = useState(false);
  const [toast, setToast] = useState<ToastState | null>(null);

  const showToast = useCallback((message: string, type: 'success' | 'error') => {
    setToast({ message, type, show: true });
  }, []);

  const hideToast = useCallback(() => {
    setToast(null);
  }, []);

  // Mock API calls - replace with actual API calls
  const handleExport = useCallback(async (options: ExportOptions) => {
    setIsExporting(true);
    setExportProgress(0);

    try {
      // Simulate export progress
      for (let i = 0; i <= 100; i += 10) {
        setExportProgress(i);
        await new Promise(resolve => setTimeout(resolve, 200));
      }

      // TODO: Replace with actual API call
      console.log('Exporting configuration:', options);

      // Create and download a mock export file
      const exportData = {
        metadata: {
          exportId: `export-${Date.now()}`,
          timestamp: new Date().toISOString(),
          version: '1.0.0',
          entityTypes: options.entityTypes,
          totalEntities: 0,
          filters: {
            includeInactive: options.includeInactive,
            includeDependencies: options.includeDependencies,
            tags: options.tags
          }
        },
        servers: [],
        virtualServers: [],
        tools: [],
        prompts: [],
        resources: [],
        policies: [],
        rateLimits: []
      };

      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json'
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `mcp-gateway-config-${Date.now()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      showToast('Configuration exported successfully', 'success');
    } catch (error) {
      console.error('Export failed:', error);
      showToast('Export failed: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error');
    } finally {
      setIsExporting(false);
      setExportProgress(0);
    }
  }, [showToast]);

  const handleValidate = useCallback(async (file: File): Promise<ValidationResult> => {
    setIsValidating(true);

    try {
      // Simulate validation
      await new Promise(resolve => setTimeout(resolve, 1500));

      // TODO: Replace with actual API call
      console.log('Validating import file:', file.name);

      // Mock validation result
      const result: ValidationResult = {
        valid: true,
        errors: [],
        warnings: [
          {
            code: 'MOCK_WARNING',
            message: 'This is a mock validation - no actual validation performed'
          }
        ],
        entityCounts: {
          servers: 0,
          tools: 0,
          prompts: 0,
          resources: 0
        },
        conflicts: []
      };

      return result;
    } catch (error) {
      console.error('Validation failed:', error);
      throw error;
    } finally {
      setIsValidating(false);
    }
  }, []);

  const handleImport = useCallback(async (file: File, options: ImportOptions) => {
    setIsImporting(true);
    setShowImportProgress(true);

    const newProgress: ImportProgress = {
      status: 'running',
      total: 10, // Mock total
      processed: 0,
      created: 0,
      updated: 0,
      skipped: 0,
      failed: 0,
      currentItem: 'Starting import...'
    };
    setImportProgress(newProgress);

    try {
      // Simulate import progress
      for (let i = 1; i <= 10; i++) {
        await new Promise(resolve => setTimeout(resolve, 500));

        setImportProgress(prev => ({
          ...prev,
          processed: i,
          created: Math.floor(i * 0.6),
          updated: Math.floor(i * 0.3),
          skipped: Math.floor(i * 0.1),
          currentItem: `Processing item ${i} of 10`
        }));
      }

      // Final status
      setImportProgress(prev => ({
        ...prev,
        status: 'completed',
        currentItem: undefined
      }));

      // TODO: Replace with actual API call
      console.log('Importing configuration:', file.name, options);

      showToast(
        options.dryRun
          ? 'Import validation completed successfully'
          : 'Configuration imported successfully',
        'success'
      );
    } catch (error) {
      console.error('Import failed:', error);
      setImportProgress(prev => ({
        ...prev,
        status: 'failed'
      }));
      showToast('Import failed: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error');
    } finally {
      setIsImporting(false);
    }
  }, [showToast]);

  const tabs = [
    { id: 'export', label: 'Export', icon: '📤' },
    { id: 'import', label: 'Import', icon: '📥' }
  ];

  return (
    <ProtectedRoute>
      <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
        <header style={{ marginBottom: '2rem' }}>
          <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#333', marginBottom: '0.5rem' }}>
            Configuration Management
          </h1>
          <p style={{ fontSize: '1rem', color: '#666' }}>
            Export and import your MCP Gateway configuration
          </p>
        </header>

        {/* Tab Navigation */}
        <div style={{
          borderBottom: '1px solid #e5e7eb',
          marginBottom: '2rem'
        }}>
          <div style={{
            display: 'flex',
            gap: '0',
            marginBottom: '-1px'
          }}>
            {tabs.map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                style={{
                  padding: '0.75rem 1.5rem',
                  border: 'none',
                  backgroundColor: 'transparent',
                  fontSize: '0.875rem',
                  fontWeight: activeTab === tab.id ? '600' : '400',
                  color: activeTab === tab.id ? '#3b82f6' : '#6b7280',
                  cursor: 'pointer',
                  borderBottom: activeTab === tab.id ? '2px solid #3b82f6' : '2px solid transparent',
                  transition: 'all 0.2s ease-in-out'
                }}
                onMouseEnter={(e) => {
                  if (activeTab !== tab.id) {
                    e.currentTarget.style.color = '#374151';
                  }
                }}
                onMouseLeave={(e) => {
                  if (activeTab !== tab.id) {
                    e.currentTarget.style.color = '#6b7280';
                  }
                }}
              >
                {tab.icon} {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Tab Content */}
        {activeTab === 'export' && (
          <ConfigurationExportPanel
            onExport={handleExport}
            isExporting={isExporting}
            exportProgress={exportProgress}
          />
        )}

        {activeTab === 'import' && (
          <>
            <ConfigurationImportPanel
              onImport={handleImport}
              onValidate={handleValidate}
              isImporting={isImporting}
              isValidating={isValidating}
            />

            <ImportProgressTracker
              progress={importProgress}
              visible={showImportProgress}
            />
          </>
        )}

        {toast && (
          <Toast
            message={toast.message}
            type={toast.type}
            onClose={hideToast}
          />
        )}
      </div>
    </ProtectedRoute>
  );
}
