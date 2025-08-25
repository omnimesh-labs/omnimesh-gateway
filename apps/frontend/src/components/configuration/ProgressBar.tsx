'use client';

interface ProgressBarProps {
  value: number; // 0-100
  max?: number;
  className?: string;
  showPercentage?: boolean;
  color?: 'blue' | 'green' | 'red' | 'yellow';
}

const COLOR_STYLES = {
  blue: '#3b82f6',
  green: '#10b981',
  red: '#ef4444',
  yellow: '#f59e0b'
};

export function ProgressBar({
  value,
  max = 100,
  className = '',
  showPercentage = true,
  color = 'blue'
}: ProgressBarProps) {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100);

  return (
    <div className={className}>
      {showPercentage && (
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
            Progress
          </span>
          <span style={{
            fontSize: '0.875rem',
            color: '#6b7280'
          }}>
            {Math.round(percentage)}%
          </span>
        </div>
      )}

      <div style={{
        width: '100%',
        height: '8px',
        backgroundColor: '#e5e7eb',
        borderRadius: '4px',
        overflow: 'hidden'
      }}>
        <div
          style={{
            height: '100%',
            backgroundColor: COLOR_STYLES[color],
            borderRadius: '4px',
            transition: 'width 0.3s ease-in-out',
            width: `${percentage}%`
          }}
        />
      </div>
    </div>
  );
}
