package handler

import (
	llwmap "crdt/internal/lww_map"
)

const (
	Add string = "add"
	Del string = "del"
)

type (
	NodeState struct {
		MapState map[string]string `json:"map_state"`
		IsOnline bool              `json:"is_online"`
	}

	UpdateRequest struct {
		Operation string `json:"operation"`
		Key       string `json:"key"`
		Value     string `json:"value"`
	}

	SyncRequest struct {
		Adds    map[string]llwmap.Pair `json:"adds"`
		Removes map[string]llwmap.Pair `json:"removes"`
	}
)
