package services

import (
	"context"
	"database/sql"
	"fmt"
	"mcp-gateway/apps/backend/internal/database/repositories"
	"mcp-gateway/apps/backend/internal/types"
	"net/http"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
)

// EndpointService handles endpoint operations
type EndpointService struct {
	repo          *repositories.EndpointRepository
	namespaceRepo *repositories.NamespaceRepository
	cache         sync.Map // Simple cache for endpoint lookups
	baseURL       string
}

// NewEndpointService creates a new endpoint service
func NewEndpointService(db *sql.DB, baseURL string) *EndpointService {
	sqlxDB := sqlx.NewDb(db, "postgres")

	return &EndpointService{
		repo:          repositories.NewEndpointRepository(sqlxDB),
		namespaceRepo: repositories.NewNamespaceRepository(sqlxDB),
		baseURL:       baseURL,
	}
}

// CreateEndpoint creates a new endpoint
func (s *EndpointService) CreateEndpoint(ctx context.Context, req types.CreateEndpointRequest, orgID string, userID *string) (*types.Endpoint, error) {
	// Validate endpoint name uniqueness
	if err := s.repo.ValidateNameUniqueness(ctx, req.Name); err != nil {
		return nil, err
	}

	// Validate endpoint name format
	if err := s.validateEndpointName(req.Name); err != nil {
		return nil, err
	}

	// Verify namespace exists and user has access
	namespace, err := s.namespaceRepo.GetByID(ctx, req.NamespaceID)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}

	// Verify namespace belongs to the organization
	if namespace.OrganizationID != orgID {
		return nil, fmt.Errorf("namespace does not belong to organization")
	}

	// Set defaults
	if req.RateLimitRequests == 0 {
		req.RateLimitRequests = 100
	}
	if req.RateLimitWindow == 0 {
		req.RateLimitWindow = 60
	}
	if len(req.AllowedOrigins) == 0 {
		req.AllowedOrigins = []string{"*"}
	}
	if len(req.AllowedMethods) == 0 {
		req.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
	}

	// Create endpoint
	endpoint := &types.Endpoint{
		OrganizationID:     orgID,
		NamespaceID:        req.NamespaceID,
		Name:               req.Name,
		Description:        req.Description,
		EnableAPIKeyAuth:   req.EnableAPIKeyAuth,
		EnableOAuth:        req.EnableOAuth,
		EnablePublicAccess: req.EnablePublicAccess,
		UseQueryParamAuth:  req.UseQueryParamAuth,
		RateLimitRequests:  req.RateLimitRequests,
		RateLimitWindow:    req.RateLimitWindow,
		AllowedOrigins:     req.AllowedOrigins,
		AllowedMethods:     req.AllowedMethods,
		CreatedBy:          userID,
		IsActive:           true,
		Metadata:           req.Metadata,
	}

	if err := s.repo.Create(ctx, endpoint); err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	// Generate URLs
	endpoint.URLs = s.generateURLs(endpoint.Name)

	// Fetch namespace details
	if endpoint.NamespaceID != "" {
		namespace, err := s.namespaceRepo.GetByID(ctx, endpoint.NamespaceID)
		if err == nil {
			endpoint.Namespace = namespace
		}
	}

	// Invalidate cache
	s.clearCache(endpoint.Name)

	return endpoint, nil
}

// GetEndpoint retrieves an endpoint by ID
func (s *EndpointService) GetEndpoint(ctx context.Context, id string) (*types.Endpoint, error) {
	endpoint, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Generate URLs
	endpoint.URLs = s.generateURLs(endpoint.Name)

	// Fetch namespace details
	if endpoint.NamespaceID != "" {
		namespace, err := s.namespaceRepo.GetByID(ctx, endpoint.NamespaceID)
		if err == nil {
			endpoint.Namespace = namespace
		}
	}

	return endpoint, nil
}

// GetEndpointByName retrieves an endpoint by name
func (s *EndpointService) GetEndpointByName(ctx context.Context, name string) (*types.Endpoint, error) {
	endpoint, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Generate URLs
	endpoint.URLs = s.generateURLs(endpoint.Name)

	// Fetch namespace details
	if endpoint.NamespaceID != "" {
		namespace, err := s.namespaceRepo.GetByID(ctx, endpoint.NamespaceID)
		if err == nil {
			endpoint.Namespace = namespace
		}
	}

	return endpoint, nil
}

// ListEndpoints lists all endpoints for an organization
func (s *EndpointService) ListEndpoints(ctx context.Context, orgID string) ([]*types.Endpoint, error) {
	endpoints, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Generate URLs and fetch namespace for each endpoint
	for _, endpoint := range endpoints {
		endpoint.URLs = s.generateURLs(endpoint.Name)

		// Fetch namespace details for each endpoint
		if endpoint.NamespaceID != "" {
			namespace, err := s.namespaceRepo.GetByID(ctx, endpoint.NamespaceID)
			if err == nil {
				endpoint.Namespace = namespace
			}
		}
	}

	return endpoints, nil
}

// ListPublicEndpoints lists all public endpoints
func (s *EndpointService) ListPublicEndpoints(ctx context.Context) ([]*types.Endpoint, error) {
	endpoints, err := s.repo.ListPublic(ctx)
	if err != nil {
		return nil, err
	}

	// Generate URLs and fetch namespace for each endpoint
	for _, endpoint := range endpoints {
		endpoint.URLs = s.generateURLs(endpoint.Name)

		// Fetch namespace details for each endpoint
		if endpoint.NamespaceID != "" {
			namespace, err := s.namespaceRepo.GetByID(ctx, endpoint.NamespaceID)
			if err == nil {
				endpoint.Namespace = namespace
			}
		}
	}

	return endpoints, nil
}

// UpdateEndpoint updates an endpoint
func (s *EndpointService) UpdateEndpoint(ctx context.Context, id string, req types.UpdateEndpointRequest) (*types.Endpoint, error) {
	endpoint, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Description != "" {
		endpoint.Description = req.Description
	}
	if req.EnableAPIKeyAuth != nil {
		endpoint.EnableAPIKeyAuth = *req.EnableAPIKeyAuth
	}
	if req.EnableOAuth != nil {
		endpoint.EnableOAuth = *req.EnableOAuth
	}
	if req.EnablePublicAccess != nil {
		endpoint.EnablePublicAccess = *req.EnablePublicAccess
	}
	if req.UseQueryParamAuth != nil {
		endpoint.UseQueryParamAuth = *req.UseQueryParamAuth
	}
	if req.RateLimitRequests != nil {
		endpoint.RateLimitRequests = *req.RateLimitRequests
	}
	if req.RateLimitWindow != nil {
		endpoint.RateLimitWindow = *req.RateLimitWindow
	}
	if len(req.AllowedOrigins) > 0 {
		endpoint.AllowedOrigins = req.AllowedOrigins
	}
	if len(req.AllowedMethods) > 0 {
		endpoint.AllowedMethods = req.AllowedMethods
	}
	if req.IsActive != nil {
		endpoint.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		endpoint.Metadata = req.Metadata
	}

	if err := s.repo.Update(ctx, endpoint); err != nil {
		return nil, err
	}

	// Generate URLs
	endpoint.URLs = s.generateURLs(endpoint.Name)

	// Fetch namespace details
	if endpoint.NamespaceID != "" {
		namespace, err := s.namespaceRepo.GetByID(ctx, endpoint.NamespaceID)
		if err == nil {
			endpoint.Namespace = namespace
		}
	}

	// Invalidate cache
	s.clearCache(endpoint.Name)

	return endpoint, nil
}

// DeleteEndpoint deletes an endpoint
func (s *EndpointService) DeleteEndpoint(ctx context.Context, id string) error {
	// Get endpoint to clear cache
	endpoint, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete endpoint
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	s.clearCache(endpoint.Name)

	return nil
}

// ResolveEndpoint resolves an endpoint by name (used by middleware)
func (s *EndpointService) ResolveEndpoint(ctx context.Context, name string) (*types.EndpointConfig, error) {
	// Check cache
	if cached, ok := s.cache.Load(name); ok {
		return cached.(*types.EndpointConfig), nil
	}

	// Load endpoint and namespace
	endpoint, err := s.repo.GetByNameWithNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	config := &types.EndpointConfig{
		Endpoint:  endpoint,
		Namespace: endpoint.Namespace,
	}

	// Cache the result
	s.cache.Store(name, config)

	return config, nil
}

// ValidateAccess validates access to an endpoint
func (s *EndpointService) ValidateAccess(ctx context.Context, endpoint *types.Endpoint, req *http.Request) error {
	// Check if endpoint is active
	if !endpoint.IsActive {
		return fmt.Errorf("endpoint is not active")
	}

	// If public access is enabled, allow
	if endpoint.EnablePublicAccess {
		return nil
	}

	// Check authentication based on endpoint settings
	authHeader := req.Header.Get("Authorization")
	apiKeyParam := req.URL.Query().Get("api_key")

	// Check API key in header
	if endpoint.EnableAPIKeyAuth && authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			// Validate API key
			// TODO: Implement actual API key validation
			return nil
		}
	}

	// Check API key in query param
	if endpoint.UseQueryParamAuth && apiKeyParam != "" {
		// Validate API key
		// TODO: Implement actual API key validation
		return nil
	}

	// Check OAuth
	if endpoint.EnableOAuth && authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			// Validate OAuth token
			// TODO: Implement actual OAuth validation
			return nil
		}
	}

	return fmt.Errorf("unauthorized: no valid authentication provided")
}

// Private helper methods

func (s *EndpointService) validateEndpointName(name string) error {
	if len(name) < 3 || len(name) > 50 {
		return fmt.Errorf("endpoint name must be between 3 and 50 characters")
	}

	// Name should only contain alphanumeric, underscore, and hyphen
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return fmt.Errorf("endpoint name can only contain alphanumeric characters, underscores, and hyphens")
		}
	}

	return nil
}

func (s *EndpointService) generateURLs(endpointName string) *types.EndpointURLs {
	baseURL := s.baseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &types.EndpointURLs{
		SSE:           fmt.Sprintf("%s/api/public/endpoints/%s/sse", baseURL, endpointName),
		HTTP:          fmt.Sprintf("%s/api/public/endpoints/%s/mcp", baseURL, endpointName),
		WebSocket:     fmt.Sprintf("%s/api/public/endpoints/%s/ws", baseURL, endpointName),
		OpenAPI:       fmt.Sprintf("%s/api/public/endpoints/%s/api/openapi.json", baseURL, endpointName),
		Documentation: fmt.Sprintf("%s/api/public/endpoints/%s/api/docs", baseURL, endpointName),
	}
}

func (s *EndpointService) clearCache(endpointName string) {
	s.cache.Delete(endpointName)
}
