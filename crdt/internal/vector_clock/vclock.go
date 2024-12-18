package vclock

type VectorClock struct {
	clock map[string]int
}

func NewVectorClock() *VectorClock {
	return &VectorClock{
		clock: make(map[string]int),
	}
}

func (vc *VectorClock) Increment(nodeID string) {
	vc.clock[nodeID]++
}

func (vc *VectorClock) Merge(other *VectorClock) {
	for node, timestamp := range other.clock {
		if current, exists := vc.clock[node]; !exists || timestamp > current {
			vc.clock[node] = timestamp
		}
	}
}

func (vc *VectorClock) HappensBefore(other *VectorClock) bool {
	for node, timestamp := range vc.clock {
		if other.clock[node] < timestamp {
			return false
		}
	}
	return true
}
