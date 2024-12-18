package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	llwset "crdt/internal/lww_set"
	vclock "crdt/internal/vector_clock"
)

type Replica struct {
	id    string
	peers []string

	lwwSet      *llwset.LWWSet
	vectorClock *vclock.VectorClock
	online      bool

	heartbeatTimer   *time.Timer
	heartbeatTimeout time.Duration

	mu sync.Mutex
}

func NewReplica(id string, peers []string) *Replica {
	return &Replica{
		id:               id,
		lwwSet:           llwset.NewLWWSet(),
		vectorClock:      vclock.NewVectorClock(),
		peers:            peers,
		online:           false,
		mu:               sync.Mutex{},
		heartbeatTimeout: 5 * time.Second,
	}
}

func (r *Replica) HandleState(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := map[string]interface{}{
		"state":  r.lwwSet.GetState(),
		"online": r.online,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (r *Replica) HandleUpdate(w http.ResponseWriter, req *http.Request) {
	if !r.online {
		http.Error(w, "Replica is offline", http.StatusServiceUnavailable)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var updates map[string]string
	if err := json.NewDecoder(req.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for key, operation := range updates {
		r.vectorClock.Increment(r.id)
		if operation == "add" {
			r.lwwSet.Add(key, r.vectorClock)
		} else if operation == "remove" {
			r.lwwSet.Remove(key, r.vectorClock)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (r *Replica) HandleSync(w http.ResponseWriter, req *http.Request) {
	if !r.online {
		http.Error(w, "Replica is offline", http.StatusServiceUnavailable)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var remoteState struct {
		Adds    map[string]*vclock.VectorClock `json:"adds"`
		Removes map[string]*vclock.VectorClock `json:"removes"`
	}

	if err := json.NewDecoder(req.Body).Decode(&remoteState); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for key, remoteVC := range remoteState.Adds {
		r.lwwSet.Add(key, remoteVC)
	}
	for key, remoteVC := range remoteState.Removes {
		r.lwwSet.Remove(key, remoteVC)
	}

	w.WriteHeader(http.StatusOK)
}

func (r *Replica) HandleSwitch(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.online {
		r.heartbeatTimer.Stop()
	} else {
		r.heartbeat()
	}
	r.online = !r.online

	w.WriteHeader(http.StatusOK)
}

func (r *Replica) heartbeat() {
	for _, peer := range r.peers {
		go func(peer string) {

		}(peer)
	}
	r.heartbeatTimer = time.AfterFunc(r.heartbeatTimeout, r.heartbeat)
}
