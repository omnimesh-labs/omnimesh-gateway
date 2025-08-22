package virtual

import (
	"database/sql"
	"fmt"
	"sync"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// Service manages virtual MCP servers
type Service struct {
	db     *sql.DB
	models *models.VirtualServerModel
	cache  *sync.Map // In-memory cache for performance
	mu     sync.RWMutex
}

// dbWrapper wraps *sql.DB to implement the Database interface
type dbWrapper struct {
	*sql.DB
}

// NewService creates a new virtual server service
func NewService(db *sql.DB) *Service {
	dbWrap := &dbWrapper{db}
	return &Service{
		db:     db,
		models: models.NewVirtualServerModel(dbWrap),
		cache:  &sync.Map{},
	}
}

// Add registers a new virtual server
func (s *Service) Add(spec *types.VirtualServerSpec) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert spec to database model
	vs := &types.VirtualServer{
		ID:             uuid.New(),
		OrganizationID: uuid.MustParse("00000000-0000-0000-0000-000000000000"), // Default org for single-tenant
		Name:           spec.Name,
		Description:    spec.Description,
		AdapterType:    spec.AdapterType,
		ToolsData:      spec.Tools,
		IsActive:       true,
		Metadata:       make(map[string]interface{}),
	}

	// Persist to database
	if err := s.models.Create(vs); err != nil {
		return fmt.Errorf("failed to create virtual server: %w", err)
	}

	// Update cache
	s.cache.Store(spec.ID, spec)

	return nil
}

// Get retrieves a virtual server by ID
func (s *Service) Get(id string) (*types.VirtualServerSpec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache first
	if cached, exists := s.cache.Load(id); exists {
		if spec, ok := cached.(*types.VirtualServerSpec); ok {
			return spec, nil
		}
	}

	// Parse UUID
	serverID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %w", err)
	}

	// Load from database
	vs, err := s.models.GetByID(serverID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("virtual server not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get virtual server: %w", err)
	}

	// Convert to spec
	spec := &types.VirtualServerSpec{
		ID:          vs.ID.String(),
		Name:        vs.Name,
		Description: vs.Description,
		AdapterType: vs.AdapterType,
		Tools:       vs.ToolsData,
		CreatedAt:   vs.CreatedAt,
		UpdatedAt:   vs.UpdatedAt,
	}

	// Update cache
	s.cache.Store(id, spec)

	return spec, nil
}

// GetByName retrieves a virtual server by name
func (s *Service) GetByName(name string) (*types.VirtualServerSpec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Default org

	// Load from database
	vs, err := s.models.GetByName(orgID, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("virtual server not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get virtual server: %w", err)
	}

	// Convert to spec
	spec := &types.VirtualServerSpec{
		ID:          vs.ID.String(),
		Name:        vs.Name,
		Description: vs.Description,
		AdapterType: vs.AdapterType,
		Tools:       vs.ToolsData,
		CreatedAt:   vs.CreatedAt,
		UpdatedAt:   vs.UpdatedAt,
	}

	// Update cache
	s.cache.Store(spec.ID, spec)

	return spec, nil
}

// List retrieves all virtual servers
func (s *Service) List() ([]*types.VirtualServerSpec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Default org

	// Load from database
	servers, err := s.models.List(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual servers: %w", err)
	}

	// Convert to specs and update cache
	var specs []*types.VirtualServerSpec
	for _, vs := range servers {
		spec := &types.VirtualServerSpec{
			ID:          vs.ID.String(),
			Name:        vs.Name,
			Description: vs.Description,
			AdapterType: vs.AdapterType,
			Tools:       vs.ToolsData,
			CreatedAt:   vs.CreatedAt,
			UpdatedAt:   vs.UpdatedAt,
		}
		specs = append(specs, spec)
		s.cache.Store(spec.ID, spec)
	}

	return specs, nil
}

// Delete removes a virtual server
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse UUID
	serverID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid server ID: %w", err)
	}

	// Delete from database
	if err := s.models.Delete(serverID); err != nil {
		return fmt.Errorf("failed to delete virtual server: %w", err)
	}

	// Remove from cache
	s.cache.Delete(id)

	return nil
}

// Update modifies an existing virtual server
func (s *Service) Update(id string, spec *types.VirtualServerSpec) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse UUID
	serverID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid server ID: %w", err)
	}

	// Get existing record
	vs, err := s.models.GetByID(serverID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("virtual server not found: %s", id)
		}
		return fmt.Errorf("failed to get virtual server: %w", err)
	}

	// Update fields
	vs.Name = spec.Name
	vs.Description = spec.Description
	vs.AdapterType = spec.AdapterType
	vs.ToolsData = spec.Tools

	// Persist to database
	if err := s.models.Update(vs); err != nil {
		return fmt.Errorf("failed to update virtual server: %w", err)
	}

	// Update spec with new timestamp
	spec.UpdatedAt = vs.UpdatedAt

	// Update cache
	s.cache.Store(id, spec)

	return nil
}

// ClearCache clears the in-memory cache
func (s *Service) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = &sync.Map{}
}

// LoadCache preloads the cache with all virtual servers
func (s *Service) LoadCache() error {
	specs, err := s.List()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, spec := range specs {
		s.cache.Store(spec.ID, spec)
	}

	return nil
}
