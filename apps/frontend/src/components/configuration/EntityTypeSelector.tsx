'use client';

import { useState } from 'react';

interface EntityTypeSelectorProps {
  entityTypes: string[];
  selectedTypes: string[];
  onChange: (selectedTypes: string[]) => void;
  label?: string;
}

const ENTITY_TYPE_LABELS: Record<string, string> = {
  servers: 'Servers',
  virtual_servers: 'Virtual Servers',
  tools: 'Tools', 
  prompts: 'Prompts',
  resources: 'Resources',
  policies: 'Policies',
  rate_limits: 'Rate Limits',
  users: 'Users',
  api_keys: 'API Keys',
  roots: 'Roots',
};

export function EntityTypeSelector({
  entityTypes,
  selectedTypes,
  onChange,
  label = 'Entity Types'
}: EntityTypeSelectorProps) {
  const [selectAll, setSelectAll] = useState(selectedTypes.length === entityTypes.length);

  const handleSelectAll = () => {
    if (selectAll) {
      onChange([]);
      setSelectAll(false);
    } else {
      onChange([...entityTypes]);
      setSelectAll(true);
    }
  };

  const handleTypeToggle = (type: string) => {
    const newSelection = selectedTypes.includes(type)
      ? selectedTypes.filter(t => t !== type)
      : [...selectedTypes, type];
    
    onChange(newSelection);
    setSelectAll(newSelection.length === entityTypes.length);
  };

  return (
    <div>
      <label style={{
        display: 'block',
        fontSize: '0.875rem',
        fontWeight: '500',
        color: '#374151',
        marginBottom: '0.5rem'
      }}>
        {label}
      </label>
      
      <div style={{
        border: '1px solid #d1d5db',
        borderRadius: '6px',
        padding: '0.75rem',
        backgroundColor: '#f9fafb',
        maxHeight: '200px',
        overflowY: 'auto'
      }}>
        <div style={{ marginBottom: '0.5rem' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <input
              type="checkbox"
              checked={selectAll}
              onChange={handleSelectAll}
              style={{ cursor: 'pointer' }}
            />
            <span style={{
              fontSize: '0.875rem',
              fontWeight: '500',
              color: '#374151'
            }}>
              Select All
            </span>
          </label>
        </div>
        
        <div style={{
          borderTop: '1px solid #e5e7eb',
          paddingTop: '0.5rem',
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))',
          gap: '0.5rem'
        }}>
          {entityTypes.map(type => (
            <label
              key={type}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                cursor: 'pointer',
                padding: '0.25rem',
                borderRadius: '4px'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.backgroundColor = '#f3f4f6';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.backgroundColor = 'transparent';
              }}
            >
              <input
                type="checkbox"
                checked={selectedTypes.includes(type)}
                onChange={() => handleTypeToggle(type)}
                style={{ cursor: 'pointer' }}
              />
              <span style={{
                fontSize: '0.875rem',
                color: '#6b7280'
              }}>
                {ENTITY_TYPE_LABELS[type] || type}
              </span>
            </label>
          ))}
        </div>
      </div>
    </div>
  );
}