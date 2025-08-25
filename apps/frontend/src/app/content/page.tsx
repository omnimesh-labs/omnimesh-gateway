'use client';

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { HealthCheck } from '@/components/HealthCheck';
import { useToast } from '@/components/Toast';
import { ResourceTable } from '@/components/content/resources/ResourceTable';
import { ResourceModal } from '@/components/content/resources/ResourceModal';
import { ResourceCard } from '@/components/content/resources/ResourceCard';
import { PromptTable } from '@/components/content/prompts/PromptTable';
import { PromptModal } from '@/components/content/prompts/PromptModal';
import { ToolTable } from '@/components/content/tools/ToolTable';
import { ToolModal } from '@/components/content/tools/ToolModal';
import {
  resourceApi,
  promptApi,
  toolApi,
  type Resource,
  type Prompt,
  type Tool,
  type CreateResourceRequest,
  type UpdateResourceRequest,
  type CreatePromptRequest,
  type UpdatePromptRequest,
  type CreateToolRequest,
  type UpdateToolRequest
} from '@/lib/api';

// Resource Tab Component
const ResourceTab = ({
  resources,
  onRefresh,
  onShowToast,
  onShowCreateModal
}: {
  resources: Resource[];
  onRefresh: () => void;
  onShowToast: (message: string, type: 'success' | 'error') => void;
  onShowCreateModal: () => void;
}) => {
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [viewingResource, setViewingResource] = useState<Resource | null>(null);
  const [showModal, setShowModal] = useState(false);

  const handleEdit = (resource: Resource) => {
    setSelectedResource(resource);
    setShowModal(true);
  };

  const handleView = (resource: Resource) => {
    setViewingResource(resource);
  };

  const handleDelete = async (resource: Resource) => {
    try {
      await resourceApi.deleteResource(resource.id);
      onShowToast(`Resource "${resource.name}" deleted successfully`, 'success');
      onRefresh();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete resource';
      onShowToast(errorMessage, 'error');
      throw error; // Re-throw for button loading state
    }
  };

  const handleSave = async (data: CreateResourceRequest | UpdateResourceRequest) => {
    try {
      if (selectedResource) {
        await resourceApi.updateResource(selectedResource.id, data);
        onShowToast(`Resource "${data.name}" updated successfully`, 'success');
      } else {
        await resourceApi.createResource(data as CreateResourceRequest);
        onShowToast(`Resource "${data.name}" created successfully`, 'success');
      }
      onRefresh();
      setShowModal(false);
      setSelectedResource(null);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to save resource';
      onShowToast(errorMessage, 'error');
      throw error; // Re-throw for modal loading state
    }
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedResource(null);
  };

  const handleCloseCard = () => {
    setViewingResource(null);
  };

  return (
    <>
      <ResourceTable
        resources={resources}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onView={handleView}
      />

      <ResourceModal
        resource={selectedResource || undefined}
        isOpen={showModal}
        onClose={handleCloseModal}
        onSave={handleSave}
      />

      {viewingResource && (
        <ResourceCard
          resource={viewingResource}
          isOpen={!!viewingResource}
          onClose={handleCloseCard}
          onEdit={() => {
            setSelectedResource(viewingResource);
            setViewingResource(null);
            setShowModal(true);
          }}
        />
      )}
    </>
  );
};

// Prompt Tab Component
const PromptTab = ({
  prompts,
  onRefresh,
  onShowToast,
  onShowCreateModal
}: {
  prompts: Prompt[];
  onRefresh: () => void;
  onShowToast: (message: string, type: 'success' | 'error') => void;
  onShowCreateModal: () => void;
}) => {
  const [selectedPrompt, setSelectedPrompt] = useState<Prompt | null>(null);
  const [viewingPrompt, setViewingPrompt] = useState<Prompt | null>(null);
  const [showModal, setShowModal] = useState(false);

  const handleEdit = (prompt: Prompt) => {
    setSelectedPrompt(prompt);
    setShowModal(true);
  };

  const handleView = (prompt: Prompt) => {
    setViewingPrompt(prompt);
  };

  const handleDelete = async (prompt: Prompt) => {
    try {
      await promptApi.deletePrompt(prompt.id);
      onShowToast(`Prompt "${prompt.name}" deleted successfully`, 'success');
      onRefresh();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete prompt';
      onShowToast(errorMessage, 'error');
      throw error; // Re-throw for button loading state
    }
  };

  const handleSave = async (data: CreatePromptRequest | UpdatePromptRequest) => {
    try {
      if (selectedPrompt) {
        await promptApi.updatePrompt(selectedPrompt.id, data);
        onShowToast(`Prompt "${data.name}" updated successfully`, 'success');
      } else {
        await promptApi.createPrompt(data as CreatePromptRequest);
        onShowToast(`Prompt "${data.name}" created successfully`, 'success');
      }
      onRefresh();
      setShowModal(false);
      setSelectedPrompt(null);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to save prompt';
      onShowToast(errorMessage, 'error');
      throw error; // Re-throw for modal loading state
    }
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedPrompt(null);
  };

  const handleCloseView = () => {
    setViewingPrompt(null);
  };

  return (
    <>
      <PromptTable
        prompts={prompts}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onView={handleView}
      />

      <PromptModal
        prompt={selectedPrompt || undefined}
        isOpen={showModal}
        onClose={handleCloseModal}
        onSave={handleSave}
      />

      {viewingPrompt && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '1rem'
        }}>
          <div style={{
            backgroundColor: 'white',
            borderRadius: '8px',
            width: '100%',
            maxWidth: '600px',
            maxHeight: '90vh',
            overflow: 'hidden',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
          }}>
            <div style={{
              padding: '1.5rem',
              borderBottom: '1px solid #e5e7eb',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}>
              <h2 style={{ margin: 0, fontSize: '1.25rem', fontWeight: '600', color: '#111827' }}>
                {viewingPrompt.name}
              </h2>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  onClick={() => {
                    setSelectedPrompt(viewingPrompt);
                    setViewingPrompt(null);
                    setShowModal(true);
                  }}
                  style={{
                    padding: '0.25rem 0.5rem',
                    fontSize: '0.75rem',
                    color: '#3b82f6',
                    backgroundColor: 'transparent',
                    border: '1px solid #3b82f6',
                    borderRadius: '0.375rem',
                    cursor: 'pointer'
                  }}
                >
                  Edit
                </button>
                <button
                  onClick={handleCloseView}
                  style={{
                    padding: '0.25rem 0.5rem',
                    fontSize: '0.75rem',
                    color: '#6b7280',
                    backgroundColor: 'transparent',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    cursor: 'pointer'
                  }}
                >
                  ×
                </button>
              </div>
            </div>
            <div style={{ padding: '1.5rem', maxHeight: 'calc(90vh - 120px)', overflowY: 'auto' }}>
              <div style={{ marginBottom: '1rem' }}>
                <strong>Category:</strong> {viewingPrompt.category}
              </div>
              {viewingPrompt.description && (
                <div style={{ marginBottom: '1rem' }}>
                  <strong>Description:</strong> {viewingPrompt.description}
                </div>
              )}
              <div style={{ marginBottom: '1rem' }}>
                <strong>Template:</strong>
                <pre style={{
                  backgroundColor: '#f9fafb',
                  padding: '1rem',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem',
                  fontFamily: 'monospace',
                  whiteSpace: 'pre-wrap',
                  maxHeight: '200px',
                  overflow: 'auto'
                }}>
                  {viewingPrompt.prompt_template}
                </pre>
              </div>
              <div style={{ marginBottom: '1rem' }}>
                <strong>Usage Count:</strong> {viewingPrompt.usage_count}
              </div>
              {viewingPrompt.tags && viewingPrompt.tags.length > 0 && (
                <div style={{ marginBottom: '1rem' }}>
                  <strong>Tags:</strong> {viewingPrompt.tags.join(', ')}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
};

// Tool Tab Component
const ToolTab = ({
  tools,
  onRefresh,
  onShowToast,
  onShowCreateModal
}: {
  tools: Tool[];
  onRefresh: () => void;
  onShowToast: (message: string, type: 'success' | 'error') => void;
  onShowCreateModal: () => void;
}) => {
  const [selectedTool, setSelectedTool] = useState<Tool | null>(null);
  const [viewingTool, setViewingTool] = useState<Tool | null>(null);
  const [showModal, setShowModal] = useState(false);

  const handleEdit = (tool: Tool) => {
    setSelectedTool(tool);
    setShowModal(true);
  };

  const handleView = (tool: Tool) => {
    setViewingTool(tool);
  };

  const handleDelete = async (tool: Tool) => {
    try {
      await toolApi.deleteTool(tool.id);
      onShowToast(`Tool "${tool.name}" deleted successfully`, 'success');
      onRefresh();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete tool';
      onShowToast(errorMessage, 'error');
      throw error; // Re-throw for button loading state
    }
  };

  const handleSave = async (data: CreateToolRequest | UpdateToolRequest) => {
    try {
      if (selectedTool) {
        await toolApi.updateTool(selectedTool.id, data);
        onShowToast(`Tool "${data.name}" updated successfully`, 'success');
      } else {
        await toolApi.createTool(data as CreateToolRequest);
        onShowToast(`Tool "${data.name}" created successfully`, 'success');
      }
      onRefresh();
      setShowModal(false);
      setSelectedTool(null);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to save tool';
      onShowToast(errorMessage, 'error');
      throw error; // Re-throw for modal loading state
    }
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedTool(null);
  };

  const handleCloseView = () => {
    setViewingTool(null);
  };

  return (
    <>
      <ToolTable
        tools={tools}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onView={handleView}
      />

      <ToolModal
        tool={selectedTool || undefined}
        isOpen={showModal}
        onClose={handleCloseModal}
        onSave={handleSave}
      />

      {viewingTool && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '1rem'
        }}>
          <div style={{
            backgroundColor: 'white',
            borderRadius: '8px',
            width: '100%',
            maxWidth: '700px',
            maxHeight: '90vh',
            overflow: 'hidden',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
          }}>
            <div style={{
              padding: '1.5rem',
              borderBottom: '1px solid #e5e7eb',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}>
              <h2 style={{ margin: 0, fontSize: '1.25rem', fontWeight: '600', color: '#111827' }}>
                {viewingTool.name}
              </h2>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  onClick={() => {
                    setSelectedTool(viewingTool);
                    setViewingTool(null);
                    setShowModal(true);
                  }}
                  style={{
                    padding: '0.25rem 0.5rem',
                    fontSize: '0.75rem',
                    color: '#3b82f6',
                    backgroundColor: 'transparent',
                    border: '1px solid #3b82f6',
                    borderRadius: '0.375rem',
                    cursor: 'pointer'
                  }}
                >
                  Edit
                </button>
                <button
                  onClick={handleCloseView}
                  style={{
                    padding: '0.25rem 0.5rem',
                    fontSize: '0.75rem',
                    color: '#6b7280',
                    backgroundColor: 'transparent',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    cursor: 'pointer'
                  }}
                >
                  ×
                </button>
              </div>
            </div>
            <div style={{ padding: '1.5rem', maxHeight: 'calc(90vh - 120px)', overflowY: 'auto' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
                <div><strong>Function Name:</strong> {viewingTool.function_name}</div>
                <div><strong>Category:</strong> {viewingTool.category}</div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
                <div><strong>Implementation:</strong> {viewingTool.implementation_type}</div>
                <div><strong>Usage Count:</strong> {viewingTool.usage_count}</div>
              </div>
              {viewingTool.description && (
                <div style={{ marginBottom: '1rem' }}>
                  <strong>Description:</strong> {viewingTool.description}
                </div>
              )}
              {viewingTool.endpoint_url && (
                <div style={{ marginBottom: '1rem' }}>
                  <strong>Endpoint:</strong> {viewingTool.endpoint_url}
                </div>
              )}
              <div style={{ marginBottom: '1rem' }}>
                <strong>Schema:</strong>
                <pre style={{
                  backgroundColor: '#f9fafb',
                  padding: '1rem',
                  borderRadius: '0.375rem',
                  fontSize: '0.75rem',
                  fontFamily: 'monospace',
                  whiteSpace: 'pre-wrap',
                  maxHeight: '200px',
                  overflow: 'auto'
                }}>
                  {JSON.stringify(viewingTool.schema, null, 2)}
                </pre>
              </div>
              {viewingTool.tags && viewingTool.tags.length > 0 && (
                <div style={{ marginBottom: '1rem' }}>
                  <strong>Tags:</strong> {viewingTool.tags.join(', ')}
                </div>
              )}
              {viewingTool.documentation && (
                <div style={{ marginBottom: '1rem' }}>
                  <strong>Documentation:</strong>
                  <div style={{ marginTop: '0.5rem', whiteSpace: 'pre-wrap' }}>{viewingTool.documentation}</div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default function ContentPage() {
  const [resources, setResources] = useState<Resource[]>([]);
  const [prompts, setPrompts] = useState<Prompt[]>([]);
  const [tools, setTools] = useState<Tool[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'resources' | 'prompts' | 'tools'>('resources');
  const [showCreateResourceModal, setShowCreateResourceModal] = useState(false);
  const [showCreatePromptModal, setShowCreatePromptModal] = useState(false);
  const [showCreateToolModal, setShowCreateToolModal] = useState(false);
  const { success, error: showError, ToastContainer } = useToast();

  // Load all content data
  const loadContent = async () => {
    try {
      setLoading(true);
      const [resourcesResponse, promptsResponse, toolsResponse] = await Promise.all([
        resourceApi.listResources({ limit: 50 }),
        promptApi.listPrompts({ limit: 50 }),
        toolApi.listTools({ limit: 50 })
      ]);

      setResources(resourcesResponse.data);
      setPrompts(promptsResponse.data);
      setTools(toolsResponse.data);
      setError(null);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load content';
      setError(errorMessage);
      showError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadContent();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleRefresh = () => {
    loadContent();
  };

  const handleShowToast = (message: string, type: 'success' | 'error') => {
    if (type === 'success') {
      success(message);
    } else {
      showError(message);
    }
  };

  if (loading && resources.length === 0 && prompts.length === 0 && tools.length === 0) {
    return (
      <ProtectedRoute>
        <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <div style={{ fontSize: '1.125rem', color: '#666' }}>Loading content...</div>
          </div>
        </div>
      </ProtectedRoute>
    );
  }

  const tabs = [
    { id: 'resources', label: 'Resources', count: resources.length },
    { id: 'prompts', label: 'Prompts', count: prompts.length },
    { id: 'tools', label: 'Tools', count: tools.length },
  ] as const;

  return (
    <ProtectedRoute>
      <ToastContainer />
      <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
        {/* Header */}
        <header style={{ marginBottom: '2rem' }}>
          <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#333', marginBottom: '0.5rem' }}>
            Content Management
          </h1>
          <p style={{ fontSize: '1rem', color: '#666' }}>
            Manage your organization&#39;s resources, prompts, and tools
          </p>
        </header>

        <HealthCheck />

        {error && (
          <div style={{
            background: '#fef2f2',
            border: '1px solid #fecaca',
            borderRadius: '8px',
            padding: '1rem',
            marginBottom: '1.5rem',
            color: '#dc2626'
          }}>
            {error}
          </div>
        )}

        {/* Tab Navigation */}
        <div style={{ marginBottom: '2rem' }}>
          <div style={{
            borderBottom: '1px solid #e5e7eb',
            display: 'flex',
            gap: '2rem'
          }}>
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                style={{
                  padding: '0.75rem 0',
                  fontSize: '1rem',
                  fontWeight: activeTab === tab.id ? '600' : '400',
                  color: activeTab === tab.id ? '#3b82f6' : '#6b7280',
                  background: 'none',
                  border: 'none',
                  borderBottom: activeTab === tab.id ? '2px solid #3b82f6' : '2px solid transparent',
                  cursor: 'pointer',
                  transition: 'all 0.2s'
                }}
              >
                {tab.label} ({tab.count})
              </button>
            ))}
          </div>
        </div>

        {/* Action Bar */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '1.5rem'
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#333' }}>
            {activeTab === 'resources' && 'Your Resources'}
            {activeTab === 'prompts' && 'Your Prompts'}
            {activeTab === 'tools' && 'Your Tools'}
          </h2>
          <div style={{ display: 'flex', gap: '0.75rem' }}>
            <button
              onClick={handleRefresh}
              disabled={loading}
              style={{
                background: '#f3f4f6',
                color: '#374151',
                padding: '0.5rem 1rem',
                borderRadius: '6px',
                border: 'none',
                fontSize: '0.875rem',
                cursor: loading ? 'not-allowed' : 'pointer',
                opacity: loading ? 0.5 : 1,
                transition: 'background-color 0.2s'
              }}
              onMouseOver={(e) => !loading && (e.currentTarget.style.background = '#e5e7eb')}
              onMouseOut={(e) => !loading && (e.currentTarget.style.background = '#f3f4f6')}
            >
              {loading ? 'Loading...' : 'Refresh'}
            </button>
            <button
              onClick={() => {
                if (activeTab === 'resources') {
                  setShowCreateResourceModal(true);
                } else if (activeTab === 'prompts') {
                  setShowCreatePromptModal(true);
                } else if (activeTab === 'tools') {
                  setShowCreateToolModal(true);
                }
              }}
              style={{
                background: '#3b82f6',
                color: 'white',
                padding: '0.5rem 1rem',
                borderRadius: '6px',
                border: 'none',
                fontSize: '0.875rem',
                cursor: 'pointer',
                transition: 'background-color 0.2s'
              }}
              onMouseOver={(e) => e.currentTarget.style.background = '#2563eb'}
              onMouseOut={(e) => e.currentTarget.style.background = '#3b82f6'}
            >
              {activeTab === 'resources' && 'Add Resource'}
              {activeTab === 'prompts' && 'Add Prompt'}
              {activeTab === 'tools' && 'Add Tool'}
            </button>
          </div>
        </div>

        {/* Tab Content */}
        <div style={{
          background: 'white',
          borderRadius: '8px',
          border: '1px solid #e5e7eb',
          minHeight: '400px'
        }}>
          {activeTab === 'resources' && (
            <ResourceTab
              resources={resources}
              onRefresh={handleRefresh}
              onShowToast={handleShowToast}
              onShowCreateModal={() => setShowCreateResourceModal(true)}
            />
          )}
          {activeTab === 'prompts' && (
            <PromptTab
              prompts={prompts}
              onRefresh={handleRefresh}
              onShowToast={handleShowToast}
              onShowCreateModal={() => setShowCreatePromptModal(true)}
            />
          )}
          {activeTab === 'tools' && (
            <ToolTab
              tools={tools}
              onRefresh={handleRefresh}
              onShowToast={handleShowToast}
              onShowCreateModal={() => setShowCreateToolModal(true)}
            />
          )}
        </div>

        {/* Global Create Modals */}
        <ResourceModal
          isOpen={showCreateResourceModal}
          onClose={() => setShowCreateResourceModal(false)}
          onSave={async (data) => {
            try {
              await resourceApi.createResource(data as CreateResourceRequest);
              handleShowToast(`Resource "${data.name}" created successfully`, 'success');
              loadContent();
              setShowCreateResourceModal(false);
            } catch (error) {
              const errorMessage = error instanceof Error ? error.message : 'Failed to create resource';
              handleShowToast(errorMessage, 'error');
              throw error;
            }
          }}
        />

        <PromptModal
          isOpen={showCreatePromptModal}
          onClose={() => setShowCreatePromptModal(false)}
          onSave={async (data) => {
            try {
              await promptApi.createPrompt(data as CreatePromptRequest);
              handleShowToast(`Prompt "${data.name}" created successfully`, 'success');
              loadContent();
              setShowCreatePromptModal(false);
            } catch (error) {
              const errorMessage = error instanceof Error ? error.message : 'Failed to create prompt';
              handleShowToast(errorMessage, 'error');
              throw error;
            }
          }}
        />

        <ToolModal
          isOpen={showCreateToolModal}
          onClose={() => setShowCreateToolModal(false)}
          onSave={async (data) => {
            try {
              await toolApi.createTool(data as CreateToolRequest);
              handleShowToast(`Tool "${data.name}" created successfully`, 'success');
              loadContent();
              setShowCreateToolModal(false);
            } catch (error) {
              const errorMessage = error instanceof Error ? error.message : 'Failed to create tool';
              handleShowToast(errorMessage, 'error');
              throw error;
            }
          }}
        />
      </div>
    </ProtectedRoute>
  );
}
