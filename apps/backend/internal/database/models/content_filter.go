package models

import (
	"time"
)

// ContentFilter represents a content filter configuration in the database
type ContentFilter struct {
	ID             string                 `db:"id" json:"id"`
	OrganizationID string                 `db:"organization_id" json:"organization_id"`
	Name           string                 `db:"name" json:"name"`
	Description    string                 `db:"description" json:"description"`
	Type           string                 `db:"type" json:"type"`
	Enabled        bool                   `db:"enabled" json:"enabled"`
	Priority       int                    `db:"priority" json:"priority"`
	Config         map[string]interface{} `db:"config" json:"config"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `db:"updated_at" json:"updated_at"`
	CreatedBy      *string                `db:"created_by" json:"created_by,omitempty"`
}

// FilterViolation represents a filter violation record in the database
type FilterViolation struct {
	ID               string                 `db:"id" json:"id"`
	OrganizationID   string                 `db:"organization_id" json:"organization_id"`
	FilterID         string                 `db:"filter_id" json:"filter_id"`
	RequestID        string                 `db:"request_id" json:"request_id"`
	SessionID        *string                `db:"session_id" json:"session_id,omitempty"`
	ServerID         *string                `db:"server_id" json:"server_id,omitempty"`
	ViolationType    string                 `db:"violation_type" json:"violation_type"`
	ActionTaken      string                 `db:"action_taken" json:"action_taken"`
	ContentSnippet   *string                `db:"content_snippet" json:"content_snippet,omitempty"`
	PatternMatched   *string                `db:"pattern_matched" json:"pattern_matched,omitempty"`
	Severity         string                 `db:"severity" json:"severity"`
	UserID           string                 `db:"user_id" json:"user_id"`
	RemoteIP         *string                `db:"remote_ip" json:"remote_ip,omitempty"`
	UserAgent        *string                `db:"user_agent" json:"user_agent,omitempty"`
	Direction        *string                `db:"direction" json:"direction,omitempty"`
	Metadata         map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt        time.Time              `db:"created_at" json:"created_at"`
}

// ProxyRoute represents a proxy routing rule in the database
type ProxyRoute struct {
	ID                  string                 `db:"id" json:"id"`
	OrganizationID      string                 `db:"organization_id" json:"organization_id"`
	Name                string                 `db:"name" json:"name"`
	Description         string                 `db:"description" json:"description"`
	PathPattern         string                 `db:"path_pattern" json:"path_pattern"`
	MethodPattern       string                 `db:"method_pattern" json:"method_pattern"`
	HostPattern         string                 `db:"host_pattern" json:"host_pattern"`
	TargetType          string                 `db:"target_type" json:"target_type"`
	TargetConfig        map[string]interface{} `db:"target_config" json:"target_config"`
	Enabled             bool                   `db:"enabled" json:"enabled"`
	Priority            int                    `db:"priority" json:"priority"`
	TimeoutSeconds      int                    `db:"timeout_seconds" json:"timeout_seconds"`
	MaxRetries          int                    `db:"max_retries" json:"max_retries"`
	LoadBalancerType    string                 `db:"load_balancer_type" json:"load_balancer_type"`
	HealthCheckEnabled  bool                   `db:"health_check_enabled" json:"health_check_enabled"`
	CreatedAt           time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time              `db:"updated_at" json:"updated_at"`
}

// RequestRoutingLog represents a request routing log entry in the database
type RequestRoutingLog struct {
	ID                     string                 `db:"id" json:"id"`
	OrganizationID         string                 `db:"organization_id" json:"organization_id"`
	RequestID              string                 `db:"request_id" json:"request_id"`
	Method                 string                 `db:"method" json:"method"`
	Path                   string                 `db:"path" json:"path"`
	Host                   *string                `db:"host" json:"host,omitempty"`
	UserAgent              *string                `db:"user_agent" json:"user_agent,omitempty"`
	RemoteIP               *string                `db:"remote_ip" json:"remote_ip,omitempty"`
	MatchedRouteID         *string                `db:"matched_route_id" json:"matched_route_id,omitempty"`
	TargetServerID         *string                `db:"target_server_id" json:"target_server_id,omitempty"`
	RoutingDecision        string                 `db:"routing_decision" json:"routing_decision"`
	RouteResolutionTimeMs  *int                   `db:"route_resolution_time_ms" json:"route_resolution_time_ms,omitempty"`
	TotalRequestTimeMs     *int                   `db:"total_request_time_ms" json:"total_request_time_ms,omitempty"`
	StatusCode             *int                   `db:"status_code" json:"status_code,omitempty"`
	ErrorMessage           *string                `db:"error_message" json:"error_message,omitempty"`
	Metadata               map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt              time.Time              `db:"created_at" json:"created_at"`
}

// NewContentFilter creates a new ContentFilter
func NewContentFilter(orgID, name, description, filterType string, enabled bool, priority int, config map[string]interface{}, createdBy *string) *ContentFilter {
	return &ContentFilter{
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		Type:           filterType,
		Enabled:        enabled,
		Priority:       priority,
		Config:         config,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// IsValid validates the ContentFilter model
func (cf *ContentFilter) IsValid() bool {
	validTypes := []string{"pii", "resource", "deny", "regex"}
	typeValid := false
	for _, t := range validTypes {
		if cf.Type == t {
			typeValid = true
			break
		}
	}
	
	return cf.Name != "" &&
		cf.OrganizationID != "" &&
		typeValid &&
		cf.Priority >= 1 && cf.Priority <= 1000 &&
		cf.Config != nil
}

// IsValid validates the FilterViolation model
func (fv *FilterViolation) IsValid() bool {
	validActions := []string{"block", "warn", "audit", "allow"}
	actionValid := false
	for _, a := range validActions {
		if fv.ActionTaken == a {
			actionValid = true
			break
		}
	}
	
	return fv.OrganizationID != "" &&
		fv.FilterID != "" &&
		fv.RequestID != "" &&
		fv.ViolationType != "" &&
		actionValid &&
		fv.UserID != ""
}

// IsValid validates the ProxyRoute model
func (pr *ProxyRoute) IsValid() bool {
	return pr.Name != "" &&
		pr.OrganizationID != "" &&
		pr.PathPattern != "" &&
		pr.TargetType != "" &&
		pr.Priority >= 1 && pr.Priority <= 1000 &&
		pr.TimeoutSeconds > 0 &&
		pr.MaxRetries >= 0
}

// IsValid validates the RequestRoutingLog model
func (rrl *RequestRoutingLog) IsValid() bool {
	return rrl.OrganizationID != "" &&
		rrl.RequestID != "" &&
		rrl.Method != "" &&
		rrl.Path != "" &&
		rrl.RoutingDecision != ""
}

// TableName returns the database table name for ContentFilter
func (cf *ContentFilter) TableName() string {
	return "content_filters"
}

// TableName returns the database table name for FilterViolation
func (fv *FilterViolation) TableName() string {
	return "filter_violations"
}

// TableName returns the database table name for ProxyRoute
func (pr *ProxyRoute) TableName() string {
	return "proxy_routes"
}

// TableName returns the database table name for RequestRoutingLog
func (rrl *RequestRoutingLog) TableName() string {
	return "request_routing_log"
}