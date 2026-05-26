package eval

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// UserContext is the set of attributes the caller provides for evaluation.
// Key is the attribute name (e.g. "email", "country", "plan").
type UserContext map[string]any

// EvalInput is everything the engine needs to evaluate one flag.
type EvalInput struct {
	FlagKey          string
	UserKey          string
	UserCtx          UserContext
	Enabled          bool
	Variations       int // total number of variations
	DefaultVariation int
	RulesJSON        string // raw JSON from flag_environments.rules
}

// Evaluate returns the variation index to serve.
// If the flag is disabled it immediately returns DefaultVariation.
// Otherwise it walks rules top-to-bottom; the first match wins.
// If no rule matches it returns DefaultVariation.
func Evaluate(in EvalInput) int {
	if !in.Enabled {
		return in.DefaultVariation
	}

	if in.RulesJSON == "" || in.RulesJSON == "[]" {
		return in.DefaultVariation
	}

	var rules []Rule
	if err := json.Unmarshal([]byte(in.RulesJSON), &rules); err != nil {
		return in.DefaultVariation
	}

	for _, rule := range rules {
		if matchesAll(rule.Conditions, in.UserCtx) {
			return serve(rule, in.FlagKey, in.UserKey, in.Variations, in.DefaultVariation)
		}
	}

	return in.DefaultVariation
}

// serve resolves a matched rule to a variation index.
func serve(rule Rule, flagKey, userKey string, numVariations, defaultVariation int) int {
	if rule.Serve != nil {
		idx := *rule.Serve
		if idx >= 0 && idx < numVariations {
			return idx
		}
		return defaultVariation
	}

	if len(rule.Rollout) > 0 {
		return rollout(rule.Rollout, flagKey, userKey, numVariations, defaultVariation)
	}

	return defaultVariation
}

// matchesAll returns true when every condition in the slice is satisfied.
// An empty conditions slice always matches (catch-all rule).
func matchesAll(conditions []Condition, ctx UserContext) bool {
	for _, c := range conditions {
		if !matchCondition(c, ctx) {
			return false
		}
	}
	return true
}

func matchCondition(c Condition, ctx UserContext) bool {
	raw, ok := ctx[c.Attribute]
	if !ok {
		return false
	}
	attr := fmt.Sprintf("%v", raw)

	switch c.Operator {
	case OpEquals:
		return len(c.Values) > 0 && attr == fmt.Sprintf("%v", c.Values[0])

	case OpIn:
		for _, v := range c.Values {
			if attr == fmt.Sprintf("%v", v) {
				return true
			}
		}
		return false

	case OpNotIn:
		for _, v := range c.Values {
			if attr == fmt.Sprintf("%v", v) {
				return false
			}
		}
		return true

	case OpContains:
		return len(c.Values) > 0 && strings.Contains(attr, fmt.Sprintf("%v", c.Values[0]))

	case OpStartsWith:
		return len(c.Values) > 0 && strings.HasPrefix(attr, fmt.Sprintf("%v", c.Values[0]))

	case OpEndsWith:
		return len(c.Values) > 0 && strings.HasSuffix(attr, fmt.Sprintf("%v", c.Values[0]))

	case OpGt, OpGte, OpLt, OpLte:
		return compareNumeric(c.Operator, attr, c.Values)
	}

	return false
}

func compareNumeric(op Operator, attr string, values []any) bool {
	if len(values) == 0 {
		return false
	}
	a, errA := strconv.ParseFloat(attr, 64)
	b, errB := strconv.ParseFloat(fmt.Sprintf("%v", values[0]), 64)
	if errA != nil || errB != nil {
		return false
	}
	switch op {
	case OpGt:
		return a > b
	case OpGte:
		return a >= b
	case OpLt:
		return a < b
	case OpLte:
		return a <= b
	}
	return false
}
