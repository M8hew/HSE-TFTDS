package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	llwmap "crdt/internal/lww_map"
	vclock "crdt/internal/vector_clock"
)

type Replica struct {
	id    string
	peers []string

	lwwSet      *llwmap.LWWMap
	vectorClock *vclock.VectorClock
	online      bool

	heartbeatTimer   *time.Timer
	heartbeatTimeout time.Duration

	mu sync.Mutex
}

func NewReplica(id string, peers []string) *Replica {
	return &Replica{
		id:               id,
		lwwSet:           llwmap.NewLWWMap(),
		vectorClock:      vclock.NewVectorClock(),
		peers:            peers,
		online:           false,
		mu:               sync.Mutex{},
		heartbeatTimeout: 10 * time.Second,
	}
}

func (r *Replica) HandleState(w http.ResponseWriter, req *http.Request) {
	fmt.Println("HandleState")
	r.mu.Lock()
	state := NodeState{
		MapState: r.lwwSet.GetState(),
		IsOnline: r.online,
	}

	defer r.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (r *Replica) HandleUpdate(w http.ResponseWriter, req *http.Request) {
	fmt.Println("HandleUpdate")
	r.mu.Lock()
	defer r.mu.Unlock()

	var updateReq UpdateRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	r.vectorClock.Increment(r.id)
	switch updateReq.Operation {
	case "add":
		r.lwwSet.Add(updateReq.Key, updateReq.Value, r.vectorClock)
	case "del":
		r.lwwSet.Remove(updateReq.Key, updateReq.Value, r.vectorClock)
	}

	w.WriteHeader(http.StatusOK)
}

func (r *Replica) HandleSync(w http.ResponseWriter, req *http.Request) {
	fmt.Println("HandleSync")
	if !r.online {
		http.Error(w, "Replica is offline", http.StatusServiceUnavailable)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var syncReq SyncRequest

	if err := json.NewDecoder(req.Body).Decode(&syncReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for key, remotePr := range syncReq.Adds {
		r.lwwSet.Add(key, remotePr.Value, remotePr.VectorClock)
	}
	for key, remotePr := range syncReq.Removes {
		r.lwwSet.Remove(key, remotePr.Value, remotePr.VectorClock)
	}

	w.WriteHeader(http.StatusOK)
}

func (r *Replica) HandleSwitch(w http.ResponseWriter, req *http.Request) {
	fmt.Println("HandleSwitch")
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
	fmt.Println("heartbeat")
	for _, peer := range r.peers {
		go func(peer string) {
			if !r.online {
				return
			}

			req := SyncRequest{
				Adds:    r.lwwSet.Adds,
				Removes: r.lwwSet.Removes,
			}

			for _, peer := range r.peers {
				jsonData, _ := json.Marshal(req)
				http.Post("http://"+peer+"/sync", "application/json", bytes.NewReader(jsonData))
			}
		}(peer)
	}
	r.heartbeatTimer = time.AfterFunc(r.heartbeatTimeout, r.heartbeat)
}
