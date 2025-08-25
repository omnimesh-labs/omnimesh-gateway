'use client';

interface ConflictStrategySelectorProps {
  value: string;
  onChange: (strategy: string) => void;
  disabled?: boolean;
}

const CONFLICT_STRATEGIES = [
  {
    value: 'update',
    label: 'Update existing items',
    description: 'Overwrite existing items with imported data'
  },
  {
    value: 'skip',
    label: 'Skip conflicting items',
    description: 'Keep existing items and skip importing conflicts'
  },
  {
    value: 'rename',
    label: 'Rename conflicting items',
    description: 'Import with modified names to avoid conflicts'
  },
  {
    value: 'fail',
    label: 'Fail on conflicts',
    description: 'Stop import process if any conflicts are found'
  }
];

export function ConflictStrategySelector({
  value,
  onChange,
  disabled = false
}: ConflictStrategySelectorProps) {
  const selectedStrategy = CONFLICT_STRATEGIES.find(s => s.value === value);

  return (
    <div>
      <label style={{
        display: 'block',
        fontSize: '0.875rem',
        fontWeight: '500',
        color: '#374151',
        marginBottom: '0.5rem'
      }}>
        Conflict Strategy
      </label>

      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        style={{
          width: '100%',
          padding: '0.5rem 0.75rem',
          border: '1px solid #d1d5db',
          borderRadius: '6px',
          fontSize: '0.875rem',
          backgroundColor: disabled ? '#f9fafb' : '#ffffff',
          color: disabled ? '#9ca3af' : '#111827',
          cursor: disabled ? 'not-allowed' : 'pointer'
        }}
      >
        {CONFLICT_STRATEGIES.map(strategy => (
          <option key={strategy.value} value={strategy.value}>
            {strategy.label}
          </option>
        ))}
      </select>

      {selectedStrategy && (
        <p style={{
          fontSize: '0.75rem',
          color: '#6b7280',
          margin: '0.25rem 0 0 0',
          fontStyle: 'italic'
        }}>
          {selectedStrategy.description}
        </p>
      )}
    </div>
  );
}
