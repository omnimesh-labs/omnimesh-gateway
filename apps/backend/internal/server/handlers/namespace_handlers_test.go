package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"mcp-gateway/apps/backend/internal/types"
)

// MockNamespaceService is a mock implementation of the namespace service
type MockNamespaceService struct {
	mock.Mock
}

func (m *MockNamespaceService) CreateNamespace(ctx context.Context, req types.CreateNamespaceRequest) (*types.Namespace, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Namespace), args.Error(1)
}

func (m *MockNamespaceService) GetNamespace(ctx context.Context, id string) (*types.Namespace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Namespace), args.Error(1)
}

func (m *MockNamespaceService) ListNamespaces(ctx context.Context, orgID string) ([]*types.Namespace, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Namespace), args.Error(1)
}

func (m *MockNamespaceService) UpdateNamespace(ctx context.Context, id string, req types.UpdateNamespaceRequest) (*types.Namespace, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Namespace), args.Error(1)
}

func (m *MockNamespaceService) DeleteNamespace(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNamespaceService) AddServerToNamespace(ctx context.Context, namespaceID string, req types.AddServerToNamespaceRequest) error {
	args := m.Called(ctx, namespaceID, req)
	return args.Error(0)
}

func (m *MockNamespaceService) RemoveServerFromNamespace(ctx context.Context, namespaceID, serverID string) error {
	args := m.Called(ctx, namespaceID, serverID)
	return args.Error(0)
}

func (m *MockNamespaceService) UpdateServerStatus(ctx context.Context, namespaceID, serverID string, req types.UpdateServerStatusRequest) error {
	args := m.Called(ctx, namespaceID, serverID, req)
	return args.Error(0)
}

func (m *MockNamespaceService) AggregateTools(ctx context.Context, namespaceID string) ([]types.NamespaceTool, error) {
	args := m.Called(ctx, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.NamespaceTool), args.Error(1)
}

func (m *MockNamespaceService) UpdateToolStatus(ctx context.Context, namespaceID, serverID, toolName string, req types.UpdateToolStatusRequest) error {
	args := m.Called(ctx, namespaceID, serverID, toolName, req)
	return args.Error(0)
}

func (m *MockNamespaceService) ExecuteTool(ctx context.Context, namespaceID string, req types.ExecuteNamespaceToolRequest) (*types.NamespaceToolResult, error) {
	args := m.Called(ctx, namespaceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.NamespaceToolResult), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNamespaceHandler_CreateNamespace(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := &NamespaceHandler{service: mockService}
	router := setupTestRouter()
	router.POST("/namespaces", handler.CreateNamespace)

	req := types.CreateNamespaceRequest{
		Name:        "test-namespace",
		Description: "Test namespace",
	}

	expectedNamespace := &types.Namespace{
		ID:             "ns-123",
		Name:           "test-namespace",
		Description:    "Test namespace",
		OrganizationID: "00000000-0000-0000-0000-000000000001",
	}

	mockService.On("CreateNamespace", mock.Anything, mock.MatchedBy(func(r types.CreateNamespaceRequest) bool {
		return r.Name == req.Name && r.Description == req.Description && r.OrganizationID == "00000000-0000-0000-0000-000000000001"
	})).Return(expectedNamespace, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/namespaces", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	// Debug: Print response body if not 201
	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response types.Namespace
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedNamespace.ID, response.ID)
	assert.Equal(t, expectedNamespace.Name, response.Name)
	
	mockService.AssertExpectations(t)
}

func TestNamespaceHandler_ListNamespaces(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := &NamespaceHandler{service: mockService}
	router := setupTestRouter()
	router.GET("/namespaces", handler.ListNamespaces)

	expectedNamespaces := []*types.Namespace{
		{
			ID:   "ns-1",
			Name: "namespace-1",
		},
		{
			ID:   "ns-2",
			Name: "namespace-2",
		},
	}

	mockService.On("ListNamespaces", mock.Anything, "00000000-0000-0000-0000-000000000001").
		Return(expectedNamespaces, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/namespaces", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["total"])
	
	mockService.AssertExpectations(t)
}

func TestNamespaceHandler_GetNamespace(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := &NamespaceHandler{service: mockService}
	router := setupTestRouter()
	router.GET("/namespaces/:id", handler.GetNamespace)

	expectedNamespace := &types.Namespace{
		ID:          "ns-123",
		Name:        "test-namespace",
		Description: "Test namespace",
	}

	mockService.On("GetNamespace", mock.Anything, "ns-123").
		Return(expectedNamespace, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/namespaces/ns-123", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response types.Namespace
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedNamespace.ID, response.ID)
	
	mockService.AssertExpectations(t)
}

func TestNamespaceHandler_GetNamespace_NotFound(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := &NamespaceHandler{service: mockService}
	router := setupTestRouter()
	router.GET("/namespaces/:id", handler.GetNamespace)

	mockService.On("GetNamespace", mock.Anything, "ns-999").
		Return(nil, errors.New("namespace not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/namespaces/ns-999", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestNamespaceHandler_UpdateNamespace(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := &NamespaceHandler{service: mockService}
	router := setupTestRouter()
	router.PUT("/namespaces/:id", handler.UpdateNamespace)

	isActive := false
	req := types.UpdateNamespaceRequest{
		Name:        "updated-namespace",
		Description: "Updated description",
		IsActive:    &isActive,
	}

	expectedNamespace := &types.Namespace{
		ID:          "ns-123",
		Name:        "updated-namespace",
		Description: "Updated description",
		IsActive:    false,
	}

	mockService.On("UpdateNamespace", mock.Anything, "ns-123", req).
		Return(expectedNamespace, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/namespaces/ns-123", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response types.Namespace
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedNamespace.Name, response.Name)
	assert.Equal(t, expectedNamespace.IsActive, response.IsActive)
	
	mockService.AssertExpectations(t)
}

func TestNamespaceHandler_DeleteNamespace(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := &NamespaceHandler{service: mockService}
	router := setupTestRouter()
	router.DELETE("/namespaces/:id", handler.DeleteNamespace)

	mockService.On("DeleteNamespace", mock.Anything, "ns-123").
		Return(nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/namespaces/ns-123", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestNamespaceHandler_ExecuteNamespaceTool(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := &NamespaceHandler{service: mockService}
	router := setupTestRouter()
	router.POST("/namespaces/:id/execute", handler.ExecuteNamespaceTool)

	req := types.ExecuteNamespaceToolRequest{
		Tool:      "server1__tool1",
		Arguments: map[string]interface{}{"arg": "value"},
	}

	expectedResult := &types.NamespaceToolResult{
		Success: true,
		Result:  map[string]interface{}{"output": "success"},
	}

	mockService.On("ExecuteTool", mock.Anything, "ns-123", req).
		Return(expectedResult, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/namespaces/ns-123/execute", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response types.NamespaceToolResult
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	
	mockService.AssertExpectations(t)
}