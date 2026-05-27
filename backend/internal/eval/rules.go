package eval

// Rule and Condition types mirror what the frontend stores in flag_environments.rules.
// Stored as JSON in SQLite, decoded here before evaluation.

type Operator string

const (
	OpIn         Operator = "in"
	OpNotIn      Operator = "notIn"
	OpContains   Operator = "contains"
	OpStartsWith Operator = "startsWith"
	OpEndsWith   Operator = "endsWith"
	OpEquals     Operator = "equals"
	OpGt         Operator = "gt"
	OpGte        Operator = "gte"
	OpLt         Operator = "lt"
	OpLte        Operator = "lte"
)

type Condition struct {
	Attribute string   `json:"attribute"`
	Operator  Operator `json:"operator"`
	Values    []any    `json:"values"`
	Segment   string   `json:"segment,omitempty"` // segment key — replaces Values for in/notIn when set
}

// Rule serves a fixed variation index OR a percentage rollout (non-nil Rollout).
// First-match wins — rules are evaluated in order.
type Rule struct {
	Conditions []Condition   `json:"conditions"`
	Serve      *int          `json:"serve"`   // variation index; nil when using rollout
	Rollout    []RolloutStep `json:"rollout"` // percentage split; nil/empty when using Serve
}

// RolloutStep assigns Weight percent of traffic to variation Variation.
// Weights across all steps should sum to 100.
type RolloutStep struct {
	Variation int `json:"variation"`
	Weight    int `json:"weight"`
}
