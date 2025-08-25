'use client';

import { ProgressBar } from './ProgressBar';

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

interface ImportProgressTrackerProps {
  progress: ImportProgress;
  visible: boolean;
}

export function ImportProgressTracker({ progress, visible }: ImportProgressTrackerProps) {
  if (!visible) return null;

  const percentage = progress.total > 0 ? (progress.processed / progress.total) * 100 : 0;
  const isComplete = progress.status === 'completed' || progress.status === 'failed';

  const getStatusColor = () => {
    switch (progress.status) {
      case 'completed':
        return 'green';
      case 'failed':
        return 'red';
      case 'partial':
        return 'yellow';
      default:
        return 'blue';
    }
  };

  const getStatusLabel = () => {
    switch (progress.status) {
      case 'pending':
        return 'Pending';
      case 'validating':
        return 'Validating';
      case 'running':
        return 'Importing';
      case 'completed':
        return 'Completed';
      case 'failed':
        return 'Failed';
      case 'partial':
        return 'Completed with Issues';
      default:
        return 'Processing';
    }
  };

  return (
    <div style={{
      backgroundColor: '#ffffff',
      border: '1px solid #e5e7eb',
      borderRadius: '8px',
      padding: '1.5rem',
      marginTop: '1rem'
    }}>
      <h3 style={{
        fontSize: '1rem',
        fontWeight: '500',
        color: '#111827',
        margin: '0 0 1rem 0'
      }}>
        📊 Import Status
      </h3>

      <div style={{ marginBottom: '1rem' }}>
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '0.5rem'
        }}>
          <span style={{
            fontSize: '0.875rem',
            color: '#374151',
            fontWeight: '500'
          }}>
            Status: {getStatusLabel()}
          </span>
          <span style={{
            fontSize: '0.875rem',
            color: '#6b7280'
          }}>
            {progress.processed}/{progress.total} items
          </span>
        </div>

        <ProgressBar
          value={percentage}
          color={getStatusColor()}
          showPercentage={false}
        />
      </div>

      {progress.currentItem && !isComplete && (
        <div style={{
          fontSize: '0.875rem',
          color: '#6b7280',
          marginBottom: '1rem',
          fontStyle: 'italic'
        }}>
          Currently processing: {progress.currentItem}
        </div>
      )}

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))',
        gap: '1rem',
        marginBottom: '1rem'
      }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{
            fontSize: '1.25rem',
            fontWeight: '600',
            color: '#111827',
            marginBottom: '0.25rem'
          }}>
            {progress.total}
          </div>
          <div style={{
            fontSize: '0.875rem',
            color: '#6b7280'
          }}>
            Total
          </div>
        </div>

        <div style={{ textAlign: 'center' }}>
          <div style={{
            fontSize: '1.25rem',
            fontWeight: '600',
            color: '#10b981',
            marginBottom: '0.25rem'
          }}>
            {progress.created}
          </div>
          <div style={{
            fontSize: '0.875rem',
            color: '#6b7280'
          }}>
            Created
          </div>
        </div>

        <div style={{ textAlign: 'center' }}>
          <div style={{
            fontSize: '1.25rem',
            fontWeight: '600',
            color: '#3b82f6',
            marginBottom: '0.25rem'
          }}>
            {progress.updated}
          </div>
          <div style={{
            fontSize: '0.875rem',
            color: '#6b7280'
          }}>
            Updated
          </div>
        </div>

        <div style={{ textAlign: 'center' }}>
          <div style={{
            fontSize: '1.25rem',
            fontWeight: '600',
            color: '#f59e0b',
            marginBottom: '0.25rem'
          }}>
            {progress.skipped}
          </div>
          <div style={{
            fontSize: '0.875rem',
            color: '#6b7280'
          }}>
            Skipped
          </div>
        </div>

        <div style={{ textAlign: 'center' }}>
          <div style={{
            fontSize: '1.25rem',
            fontWeight: '600',
            color: '#ef4444',
            marginBottom: '0.25rem'
          }}>
            {progress.failed}
          </div>
          <div style={{
            fontSize: '0.875rem',
            color: '#6b7280'
          }}>
            Failed
          </div>
        </div>
      </div>

      {isComplete && (
        <div style={{
          padding: '0.75rem',
          backgroundColor: progress.status === 'failed' ? '#fef2f2' : '#f0fdf4',
          border: `1px solid ${progress.status === 'failed' ? '#fecaca' : '#bbf7d0'}`,
          borderRadius: '6px',
          fontSize: '0.875rem',
          color: progress.status === 'failed' ? '#dc2626' : '#166534',
          textAlign: 'center'
        }}>
          {progress.status === 'failed'
            ? '❌ Import failed with errors'
            : progress.failed > 0
            ? '⚠️ Import completed with some failures'
            : '✅ Import completed successfully'
          }
        </div>
      )}
    </div>
  );
}
