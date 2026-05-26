package eval

import (
	"crypto/sha256"
	"encoding/binary"
)

// bucket returns a stable number in [0, 99] for a given user+flag pair.
// Using SHA-256 so the distribution is uniform and deterministic.
// Think of it like a stable hash that always puts the same user in the same bucket.
func bucket(flagKey, userKey string) int {
	h := sha256.Sum256([]byte(flagKey + "." + userKey))
	n := binary.BigEndian.Uint32(h[:4])
	return int(n % 100)
}

// rollout maps a user's stable bucket to a variation by walking cumulative weights.
// Example: [{variation:0, weight:10}, {variation:1, weight:90}]
// Users in bucket 0-9 get variation 0; bucket 10-99 get variation 1.
func rollout(steps []RolloutStep, flagKey, userKey string, numVariations, defaultVariation int) int {
	b := bucket(flagKey, userKey)
	cumulative := 0
	for _, step := range steps {
		cumulative += step.Weight
		if b < cumulative {
			if step.Variation >= 0 && step.Variation < numVariations {
				return step.Variation
			}
			return defaultVariation
		}
	}
	return defaultVariation
}
