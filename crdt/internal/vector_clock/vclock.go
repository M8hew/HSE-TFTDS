package vclock

type VectorClock struct {
	Clock map[string]int `json:"Clock"`
}

func NewVectorClock() *VectorClock {
	return &VectorClock{
		Clock: make(map[string]int),
	}
}

func (vc *VectorClock) Increment(nodeID string) {
	vc.Clock[nodeID]++
}

func (vc *VectorClock) Merge(other *VectorClock) {
	for node, timestamp := range other.Clock {
		if current, exists := vc.Clock[node]; !exists || timestamp > current {
			vc.Clock[node] = timestamp
		}
	}
}

func (vc *VectorClock) HappensBefore(other *VectorClock) bool {
	for node, timestamp := range vc.Clock {
		if other.Clock[node] < timestamp {
			return false
		}
	}
	return true
}
