package models

import (
	"encoding/json"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// PolicyModel handles policy database operations
type PolicyModel struct {
	BaseModel
}

// NewPolicyModel creates a new policy model
func NewPolicyModel(db Database) *PolicyModel {
	return &PolicyModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new policy into the database
func (m *PolicyModel) Create(policy *types.Policy) error {
	query := `
		INSERT INTO policies (id, organization_id, name, description, type, priority,
			conditions, actions, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	conditionsJSON, _ := json.Marshal(policy.Conditions)
	actionsJSON, _ := json.Marshal(policy.Actions)

	_, err := m.db.Exec(query, policy.ID, policy.OrganizationID, policy.Name,
		policy.Description, policy.Type, policy.Priority, string(conditionsJSON),
		string(actionsJSON), policy.IsActive, policy.CreatedAt, policy.UpdatedAt)

	return err
}

// GetByID retrieves a policy by ID
func (m *PolicyModel) GetByID(id string) (*types.Policy, error) {
	query := `
		SELECT id, organization_id, name, description, type, priority,
			conditions, actions, is_active, created_at, updated_at
		FROM policies
		WHERE id = $1
	`

	policy := &types.Policy{}
	var conditionsJSON, actionsJSON string

	err := m.db.QueryRow(query, id).Scan(
		&policy.ID, &policy.OrganizationID, &policy.Name, &policy.Description,
		&policy.Type, &policy.Priority, &conditionsJSON, &actionsJSON,
		&policy.IsActive, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal([]byte(conditionsJSON), &policy.Conditions)
	json.Unmarshal([]byte(actionsJSON), &policy.Actions)

	return policy, nil
}

// ListByOrganization lists policies for an organization
func (m *PolicyModel) ListByOrganization(orgID string, policyType string, activeOnly bool) ([]*types.Policy, error) {
	query := `
		SELECT id, organization_id, name, description, type, priority,
			conditions, actions, is_active, created_at, updated_at
		FROM policies
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	argIndex := 2

	if policyType != "" {
		query += " AND type = $" + string(rune(argIndex))
		args = append(args, policyType)
		argIndex++
	}

	if activeOnly {
		query += " AND is_active = true"
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*types.Policy
	for rows.Next() {
		policy := &types.Policy{}
		var conditionsJSON, actionsJSON string

		err := rows.Scan(
			&policy.ID, &policy.OrganizationID, &policy.Name, &policy.Description,
			&policy.Type, &policy.Priority, &conditionsJSON, &actionsJSON,
			&policy.IsActive, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal([]byte(conditionsJSON), &policy.Conditions)
		json.Unmarshal([]byte(actionsJSON), &policy.Actions)

		policies = append(policies, policy)
	}

	return policies, nil
}

// Update updates a policy in the database
func (m *PolicyModel) Update(policy *types.Policy) error {
	query := `
		UPDATE policies
		SET name = $1, description = $2, priority = $3, conditions = $4,
			actions = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`

	policy.UpdatedAt = time.Now()

	conditionsJSON, _ := json.Marshal(policy.Conditions)
	actionsJSON, _ := json.Marshal(policy.Actions)

	_, err := m.db.Exec(query, policy.Name, policy.Description, policy.Priority,
		string(conditionsJSON), string(actionsJSON), policy.IsActive,
		policy.UpdatedAt, policy.ID)

	return err
}

// Delete deletes a policy from the database
func (m *PolicyModel) Delete(id string) error {
	query := `DELETE FROM policies WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// RateLimitRuleModel handles rate limit rule database operations
type RateLimitRuleModel struct {
	BaseModel
}

// NewRateLimitRuleModel creates a new rate limit rule model
func NewRateLimitRuleModel(db Database) *RateLimitRuleModel {
	return &RateLimitRuleModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new rate limit rule into the database
func (m *RateLimitRuleModel) Create(rule *types.RateLimitRule) error {
	query := `
		INSERT INTO rate_limits (id, organization_id, name, description, type, limit_value,
			window_duration, algorithm, conditions, priority, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	conditionsJSON, _ := json.Marshal(rule.Conditions)

	_, err := m.db.Exec(query, rule.ID, rule.OrganizationID, rule.Name,
		rule.Description, rule.Type, rule.Limit, rule.Window.Nanoseconds(),
		rule.Algorithm, string(conditionsJSON), rule.Priority, rule.IsActive,
		rule.CreatedAt, rule.UpdatedAt)

	return err
}

// GetByID retrieves a rate limit rule by ID
func (m *RateLimitRuleModel) GetByID(id string) (*types.RateLimitRule, error) {
	query := `
		SELECT id, organization_id, name, description, type, limit_value,
			window_duration, algorithm, conditions, priority, is_active, created_at, updated_at
		FROM rate_limits
		WHERE id = $1
	`

	rule := &types.RateLimitRule{}
	var conditionsJSON string
	var windowNanos int64

	err := m.db.QueryRow(query, id).Scan(
		&rule.ID, &rule.OrganizationID, &rule.Name, &rule.Description,
		&rule.Type, &rule.Limit, &windowNanos, &rule.Algorithm, &conditionsJSON,
		&rule.Priority, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON and duration fields
	json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
	rule.Window = time.Duration(windowNanos)

	return rule, nil
}

// ListByOrganization lists rate limit rules for an organization
func (m *RateLimitRuleModel) ListByOrganization(orgID string, ruleType string, activeOnly bool) ([]*types.RateLimitRule, error) {
	query := `
		SELECT id, organization_id, name, description, type, limit_value,
			window_duration, algorithm, conditions, priority, is_active, created_at, updated_at
		FROM rate_limits
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	argIndex := 2

	if ruleType != "" {
		query += " AND type = $" + string(rune(argIndex))
		args = append(args, ruleType)
		argIndex++
	}

	if activeOnly {
		query += " AND is_active = true"
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*types.RateLimitRule
	for rows.Next() {
		rule := &types.RateLimitRule{}
		var conditionsJSON string
		var windowNanos int64

		err := rows.Scan(
			&rule.ID, &rule.OrganizationID, &rule.Name, &rule.Description,
			&rule.Type, &rule.Limit, &windowNanos, &rule.Algorithm, &conditionsJSON,
			&rule.Priority, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON and duration fields
		json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
		rule.Window = time.Duration(windowNanos)

		rules = append(rules, rule)
	}

	return rules, nil
}

// Update updates a rate limit rule in the database
func (m *RateLimitRuleModel) Update(rule *types.RateLimitRule) error {
	query := `
		UPDATE rate_limits
		SET name = $1, description = $2, limit_value = $3, window_duration = $4,
			algorithm = $5, conditions = $6, priority = $7, is_active = $8, updated_at = $9
		WHERE id = $10
	`

	rule.UpdatedAt = time.Now()
	conditionsJSON, _ := json.Marshal(rule.Conditions)

	_, err := m.db.Exec(query, rule.Name, rule.Description, rule.Limit,
		rule.Window.Nanoseconds(), rule.Algorithm, string(conditionsJSON),
		rule.Priority, rule.IsActive, rule.UpdatedAt, rule.ID)

	return err
}

// Delete deletes a rate limit rule from the database
func (m *RateLimitRuleModel) Delete(id string) error {
	query := `DELETE FROM rate_limits WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}
