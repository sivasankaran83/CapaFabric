package models

import "time"

// AgentDecisionAction is the structured action an agent decides to take.
type AgentDecisionAction string

const (
	AgentDecisionInvoke    AgentDecisionAction = "invoke"
	AgentDecisionComplete  AgentDecisionAction = "complete"
	AgentDecisionEscalate  AgentDecisionAction = "escalate"
	AgentDecisionRetry     AgentDecisionAction = "retry"
)

// AgentDecision is the structured output of one deliberation iteration.
type AgentDecision struct {
	Action       AgentDecisionAction `json:"action"`
	CapabilityID string              `json:"capability_id,omitempty"`
	Arguments    map[string]any      `json:"arguments,omitempty"`
	Reasoning    string              `json:"reasoning,omitempty"`
	Confidence   float64             `json:"confidence"`
	FinalAnswer  string              `json:"final_answer,omitempty"`
}

// AgentStateStatus represents the lifecycle state of a goal.
type AgentStateStatus string

const (
	AgentStateRunning   AgentStateStatus = "running"
	AgentStateCompleted AgentStateStatus = "completed"
	AgentStateFailed    AgentStateStatus = "failed"
	AgentStateEscalated AgentStateStatus = "escalated"
)

// AgentState tracks the persistent state of a goal pursuit.
type AgentState struct {
	GoalID        string           `json:"goal_id"`
	AgentID       string           `json:"agent_id"`
	TenantID      string           `json:"tenant_id,omitempty"`
	Goal          string           `json:"goal"`
	Status        AgentStateStatus `json:"status"`
	Iteration     int              `json:"iteration"`
	MaxIterations int              `json:"max_iterations"`
	FinalAnswer   string           `json:"final_answer,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
}
