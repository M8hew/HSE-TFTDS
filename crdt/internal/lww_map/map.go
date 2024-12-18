package llwmap

import (
	"sync"

	vclock "crdt/internal/vector_clock"
)

type (
	Pair struct {
		Value               string `json:"value"`
		*vclock.VectorClock `json:"vc"`
	}

	LWWMap struct {
		Adds    map[string]Pair
		Removes map[string]Pair
		mu      sync.Mutex
	}
)

func NewLWWMap() *LWWMap {
	return &LWWMap{
		Adds:    make(map[string]Pair),
		Removes: make(map[string]Pair),
		mu:      sync.Mutex{},
	}
}

func (lww *LWWMap) Add(key, value string, vc *vclock.VectorClock) {
	lww.mu.Lock()
	defer lww.mu.Unlock()

	if existingVC, exists := lww.Adds[key]; !exists || existingVC.HappensBefore(vc) {
		lww.Adds[key] = Pair{Value: value, VectorClock: vc}
	}
}

func (lww *LWWMap) Remove(key, value string, vc *vclock.VectorClock) {
	lww.mu.Lock()
	defer lww.mu.Unlock()

	if existingVC, exists := lww.Removes[key]; !exists || existingVC.HappensBefore(vc) {
		lww.Removes[key] = Pair{Value: value, VectorClock: vc}
	}
}

func (lww *LWWMap) GetState() map[string]string {
	lww.mu.Lock()
	defer lww.mu.Unlock()

	result := make(map[string]string)

	for key, addPr := range lww.Adds {
		if removePr, removed := lww.Removes[key]; !removed || removePr.HappensBefore(addPr.VectorClock) {
			result[key] = addPr.Value
		}
	}
	return result
}
