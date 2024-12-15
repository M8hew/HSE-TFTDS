package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalStorage_SetAndGet(t *testing.T) {
	storage := NewLocalStorage()

	err := storage.Set("key1", "value1")
	assert.NoError(t, err)

	value, ok := storage.Get("key1")

	assert.True(t, ok)
	assert.Equal(t, "value1", value)
}

func TestLocalStorage_Del(t *testing.T) {
	storage := NewLocalStorage()

	storage.Set("key2", "value2")
	deleted := storage.Del("key2")
	assert.True(t, deleted)

	_, ok := storage.Get("key2")
	assert.False(t, ok)
}

func TestLocalStorage_Update(t *testing.T) {
	storage := NewLocalStorage()

	updated := storage.Update("key3", "newValue")
	assert.False(t, updated)

	storage.Set("key3", "value3")
	updated = storage.Update("key3", "newValue")
	assert.True(t, updated)

	value, ok := storage.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, "newValue", value)
}

func TestLocalStorage_CAS(t *testing.T) {
	storage := NewLocalStorage()

	cas := storage.CAS("key4", "oldValue", "newValue")
	assert.False(t, cas)

	storage.Set("key4", "value4")
	cas = storage.CAS("key4", "wrongOldValue", "newValue")
	assert.False(t, cas)

	cas = storage.CAS("key4", "value4", "newValue")
	assert.True(t, cas)

	value, ok := storage.Get("key4")
	assert.True(t, ok)
	assert.Equal(t, "newValue", value)
}
