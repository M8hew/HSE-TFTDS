package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"kvstorage/api/oapi"
	"kvstorage/internal/raft"
	"kvstorage/internal/storage"
)

// static check
var _ oapi.ServerInterface = &Handler{}

type Handler struct {
	raftServer *raft.RaftServer
	storage    *storage.LocalStorage
	logger     *zap.Logger
}

func NewHandler(
	raftServer *raft.RaftServer,
	storage *storage.LocalStorage,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		raftServer: raftServer,
		storage:    storage,
		logger:     logger,
	}
}

// Compare-And-Swap (CAS) operation
// (POST /cas)
func (h *Handler) PostCas(ctx echo.Context) error {
	h.logger.Info("PostCas")

	if !h.raftServer.IsLeader() {
		return ctx.JSON(http.StatusSeeOther, "Cannot do write operation, not leader")
	}

	reqBody := oapi.PostCasJSONRequestBody{}
	if err := ctx.Bind(&reqBody); err != nil {
		h.logger.Info("Cannot parse request body")
		return ctx.JSON(http.StatusBadRequest, "Cannot parse request body")
	}

	if ok := h.raftServer.ReplicateEntry(raft.NewCASEntry(reqBody.Key, reqBody.OldValue, reqBody.NewValue)); !ok {
		return ctx.JSON(http.StatusConflict, "Was unable to replicate entry")
	}
	return ctx.JSON(http.StatusOK, "OK")
}

// Create a new key-value pair
// (POST /keys)
func (h *Handler) PostKeys(ctx echo.Context) error {
	h.logger.Info("PostKeys")

	if !h.raftServer.IsLeader() {
		return ctx.JSON(http.StatusSeeOther, "Cannot do write operation, not leader")
	}

	reqBody := oapi.PostKeysJSONRequestBody{}
	if err := ctx.Bind(&reqBody); err != nil {
		h.logger.Info("Cannot parse request body")
		return ctx.JSON(http.StatusBadRequest, "Cannot parse request body")
	}

	if ok := h.raftServer.ReplicateEntry(raft.NewSetEntry(reqBody.Key, reqBody.Value)); !ok {
		return ctx.JSON(http.StatusConflict, "Was unable to replicate entry")
	}
	return ctx.JSON(http.StatusOK, "OK")
}

// Delete a key and its value
// (DELETE /keys/{key})
func (h *Handler) DeleteKeysKey(ctx echo.Context, key string) error {
	h.logger.Info("DeleteKeysKey")

	if !h.raftServer.IsLeader() {
		return ctx.JSON(http.StatusSeeOther, "Cannot do write operation, not leader")
	}

	if ok := h.raftServer.ReplicateEntry(raft.NewDeleteEntry(key)); !ok {
		return ctx.JSON(http.StatusConflict, "Was unable to replicate entry")
	}
	return ctx.JSON(http.StatusOK, "OK")
}

// Update the value for an existing key
// (PUT /keys/{key})
func (h *Handler) PutKeysKey(ctx echo.Context, key string) error {
	h.logger.Info("PutKeysKey")

	if !h.raftServer.IsLeader() {
		return ctx.JSON(http.StatusSeeOther, "Cannot do write operation, not leader")
	}

	reqBody := oapi.PutKeysKeyJSONRequestBody{}
	if err := ctx.Bind(&reqBody); err != nil {
		h.logger.Info("Cannot parse request body")
		return ctx.JSON(http.StatusBadRequest, "Cannot parse request body")
	}

	if ok := h.raftServer.ReplicateEntry(raft.NewUpdateEntry(key, reqBody.Value)); !ok {
		return ctx.JSON(http.StatusConflict, "Was unable to replicate entry")
	}
	return ctx.JSON(http.StatusOK, "OK")
}

// Retrieve the value of a key
// (GET /keys/{key})
func (h *Handler) GetKeysKey(ctx echo.Context, key string) error {
	h.logger.Info("GetKeysKey")

	if h.raftServer.IsLeader() {
		return ctx.JSON(http.StatusSeeOther, "Read from master may be inconsistent, try another node")
	}

	val, ok := h.storage.Get(key)
	if !ok {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "value not found in storage",
		})
	}
	return ctx.JSON(http.StatusOK, map[string]string{
		"key: ": key,
		"value": val,
	})
}
