package llwset

import (
	"sync"

	vclock "crdt/internal/vector_clock"
)

type Pair struct {
	key    string
	tmstmp *vclock.VectorClock
}

type LWWSet struct {
	adds    map[string]*vclock.VectorClock
	removes map[string]*vclock.VectorClock
	mutex   sync.Mutex
}

func NewLWWSet() *LWWSet {
	return &LWWSet{
		adds:    make(map[string]*vclock.VectorClock),
		removes: make(map[string]*vclock.VectorClock),
	}
}

func (lww *LWWSet) Add(key string, vc *vclock.VectorClock) {
	lww.mutex.Lock()
	defer lww.mutex.Unlock()
	if existingVC, exists := lww.adds[key]; !exists || existingVC.HappensBefore(vc) {
		lww.adds[key] = vc
	}
}

func (lww *LWWSet) Remove(key string, vc *vclock.VectorClock) {
	lww.mutex.Lock()
	defer lww.mutex.Unlock()
	if existingVC, exists := lww.removes[key]; !exists || existingVC.HappensBefore(vc) {
		lww.removes[key] = vc
	}
}

func (lww *LWWSet) GetState() map[string]bool {
	lww.mutex.Lock()
	defer lww.mutex.Unlock()
	result := make(map[string]bool)
	for key, addVC := range lww.adds {
		if removeVC, removed := lww.removes[key]; !removed || addVC.HappensBefore(removeVC) {
			result[key] = true
		}
	}
	return result
}
