package constants

// AgentType identifies the type of AI agent.
type AgentType string

const (
	AgentPlanner     AgentType = "planner"
	AgentImplementer AgentType = "implementer"
	AgentReviewer    AgentType = "reviewer"
	AgentAuditor     AgentType = "auditor"
	AgentSentinel    AgentType = "sentinel"
	AgentAdvisor     AgentType = "advisor"
)
