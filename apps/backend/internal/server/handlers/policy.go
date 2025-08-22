package handlers

import (
	"database/sql"
	"encoding/json"
	"mcp-gateway/apps/backend/internal/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PolicyHandler handles policy-related requests
type PolicyHandler struct {
	db *sql.DB
}

// NewPolicyHandler creates a new policy handler
func NewPolicyHandler(db *sql.DB) *PolicyHandler {
	return &PolicyHandler{
		db: db,
	}
}

// CreatePolicy handles POST /api/admin/policies
func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
	var req types.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Get user context from middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Convert conditions and actions to JSON
	conditionsJSON, err := json.Marshal(req.Conditions)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conditions format"})
		return
	}

	actionsJSON, err := json.Marshal(req.Actions)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid actions format"})
		return
	}

	// Create policy
	policy := types.Policy{
		ID:             uuid.New().String(),
		OrganizationID: orgID.(string),
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		Priority:       req.Priority,
		IsActive:       true,
	}

	query := `
		INSERT INTO policies (id, organization_id, name, description, type, priority, conditions, actions, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`

	_, err = h.db.Exec(query, policy.ID, policy.OrganizationID, policy.Name, policy.Description, 
		policy.Type, policy.Priority, conditionsJSON, actionsJSON, policy.IsActive, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create policy", "details": err.Error()})
		return
	}

	// Parse back the JSON for response
	json.Unmarshal(conditionsJSON, &policy.Conditions)
	json.Unmarshal(actionsJSON, &policy.Actions)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Policy created successfully",
		"policy":  policy,
		"user_id": userID,
	})
}

// ListPolicies handles GET /api/admin/policies
func (h *PolicyHandler) ListPolicies(c *gin.Context) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Query parameters for filtering and pagination
	policyType := c.Query("type")
	isActiveStr := c.Query("is_active")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Build query with optional filters
	query := `
		SELECT id, organization_id, name, description, type, priority, conditions, actions, is_active, created_at, updated_at
		FROM policies 
		WHERE organization_id = $1
	`
	args := []interface{}{orgID.(string)}
	argIndex := 2

	if policyType != "" {
		query += " AND type = $" + strconv.Itoa(argIndex)
		args = append(args, policyType)
		argIndex++
	}

	if isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			query += " AND is_active = $" + strconv.Itoa(argIndex)
			args = append(args, isActive)
			argIndex++
		}
	}

	query += " ORDER BY priority DESC, created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch policies", "details": err.Error()})
		return
	}
	defer rows.Close()

	var policies []types.Policy
	for rows.Next() {
		var policy types.Policy
		var conditionsJSON, actionsJSON []byte

		err := rows.Scan(
			&policy.ID, &policy.OrganizationID, &policy.Name, &policy.Description,
			&policy.Type, &policy.Priority, &conditionsJSON, &actionsJSON,
			&policy.IsActive, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan policy", "details": err.Error()})
			return
		}

		// Parse JSON fields
		json.Unmarshal(conditionsJSON, &policy.Conditions)
		json.Unmarshal(actionsJSON, &policy.Actions)

		policies = append(policies, policy)
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
		"total":    len(policies),
		"limit":    limit,
		"offset":   offset,
	})
}

// GetPolicy handles GET /api/admin/policies/:id
func (h *PolicyHandler) GetPolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	query := `
		SELECT id, organization_id, name, description, type, priority, conditions, actions, is_active, created_at, updated_at
		FROM policies 
		WHERE id = $1 AND organization_id = $2
	`

	var policy types.Policy
	var conditionsJSON, actionsJSON []byte

	err := h.db.QueryRow(query, policyID, orgID.(string)).Scan(
		&policy.ID, &policy.OrganizationID, &policy.Name, &policy.Description,
		&policy.Type, &policy.Priority, &conditionsJSON, &actionsJSON,
		&policy.IsActive, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch policy", "details": err.Error()})
		return
	}

	// Parse JSON fields
	json.Unmarshal(conditionsJSON, &policy.Conditions)
	json.Unmarshal(actionsJSON, &policy.Actions)

	c.JSON(http.StatusOK, gin.H{"policy": policy})
}

// UpdatePolicy handles PUT /api/admin/policies/:id
func (h *PolicyHandler) UpdatePolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}

	var req types.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Check if policy exists and belongs to organization
	var existingPolicy types.Policy
	checkQuery := "SELECT id FROM policies WHERE id = $1 AND organization_id = $2"
	err := h.db.QueryRow(checkQuery, policyID, orgID.(string)).Scan(&existingPolicy.ID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check policy", "details": err.Error()})
		return
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		updateFields = append(updateFields, "name = $"+strconv.Itoa(argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.Description != "" {
		updateFields = append(updateFields, "description = $"+strconv.Itoa(argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.Priority != 0 {
		updateFields = append(updateFields, "priority = $"+strconv.Itoa(argIndex))
		args = append(args, req.Priority)
		argIndex++
	}

	if req.Conditions != nil {
		conditionsJSON, err := json.Marshal(req.Conditions)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conditions format"})
			return
		}
		updateFields = append(updateFields, "conditions = $"+strconv.Itoa(argIndex))
		args = append(args, conditionsJSON)
		argIndex++
	}

	if req.Actions != nil {
		actionsJSON, err := json.Marshal(req.Actions)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid actions format"})
			return
		}
		updateFields = append(updateFields, "actions = $"+strconv.Itoa(argIndex))
		args = append(args, actionsJSON)
		argIndex++
	}

	if req.IsActive != nil {
		updateFields = append(updateFields, "is_active = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add updated_at field
	updateFields = append(updateFields, "updated_at = NOW()")

	// Add WHERE conditions
	args = append(args, policyID, orgID.(string))
	whereClause := " WHERE id = $" + strconv.Itoa(argIndex) + " AND organization_id = $" + strconv.Itoa(argIndex+1)

	query := "UPDATE policies SET " + updateFields[0]
	for i := 1; i < len(updateFields); i++ {
		query += ", " + updateFields[i]
	}
	query += whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update policy", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Policy updated successfully",
		"policy_id": policyID,
	})
}

// DeletePolicy handles DELETE /api/admin/policies/:id
func (h *PolicyHandler) DeletePolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Check if policy exists and belongs to organization
	var existingPolicy types.Policy
	checkQuery := "SELECT id FROM policies WHERE id = $1 AND organization_id = $2"
	err := h.db.QueryRow(checkQuery, policyID, orgID.(string)).Scan(&existingPolicy.ID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check policy", "details": err.Error()})
		return
	}

	// Delete the policy
	deleteQuery := "DELETE FROM policies WHERE id = $1 AND organization_id = $2"
	_, err = h.db.Exec(deleteQuery, policyID, orgID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete policy", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Policy deleted successfully",
		"policy_id": policyID,
	})
}