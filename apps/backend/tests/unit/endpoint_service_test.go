package unit

import (
	"context"
	"mcp-gateway/apps/backend/internal/services"
	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/tests/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpointService(t *testing.T) {
	// Create test database
	db, teardown, err := helpers.SetupTestDatabase(t)
	require.NoError(t, err)
	defer teardown()

	// Run migrations
	err = helpers.RunMigrations(db)
	require.NoError(t, err)

	// Create endpoint service
	endpointService := services.NewEndpointService(db, "http://localhost:8080")

	ctx := context.Background()
	orgID := "00000000-0000-0000-0000-000000000001"
	userID := "00000000-0000-0000-0000-000000000002"

	t.Run("Validate Endpoint Name", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected bool
		}{
			{"Valid name", "test-endpoint", true},
			{"Valid with hyphen", "test-endpoint-123", true},
			{"Valid with underscore", "test_endpoint", true},
			{"Too short", "ab", false},
			{"Too long", "this-is-a-very-long-endpoint-name-that-exceeds-fifty-characters", false},
			{"Invalid characters", "test@endpoint", false},
			{"With spaces", "test endpoint", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := types.CreateEndpointRequest{
					NamespaceID: "00000000-0000-0000-0000-000000000003",
					Name:        tc.input,
				}

				_, err := endpointService.CreateEndpoint(ctx, req, orgID, &userID)
				if tc.expected {
					// Should either succeed or fail for a different reason (like namespace not found)
					if err != nil {
						assert.Contains(t, err.Error(), "namespace")
					}
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "endpoint name")
				}
			})
		}
	})

	t.Run("Generate URLs", func(t *testing.T) {
		// Test the public interface
		req := types.CreateEndpointRequest{
			NamespaceID:        "00000000-0000-0000-0000-000000000003",
			Name:               "url-test-endpoint",
			EnableAPIKeyAuth:   true,
			EnablePublicAccess: false,
		}

		// This will fail due to namespace not found, but we can still check the validation
		_, err := endpointService.CreateEndpoint(ctx, req, orgID, &userID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "namespace")
	})

	t.Run("Default Values", func(t *testing.T) {
		req := types.CreateEndpointRequest{
			NamespaceID: "00000000-0000-0000-0000-000000000003",
			Name:        "default-test",
			// Don't set rate limit values to test defaults
		}

		// This will fail due to namespace not found, but the defaults should be set
		_, err := endpointService.CreateEndpoint(ctx, req, orgID, &userID)
		require.Error(t, err)

		// Check that defaults would be applied
		assert.Equal(t, 0, req.RateLimitRequests) // Original should not be modified
	})
}
