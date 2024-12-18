package llwmap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	vclock "crdt/internal/vector_clock"
)

func TestNewLWWMap(t *testing.T) {
	lww := NewLWWMap()
	assert.NotNil(t, lww)
	assert.Empty(t, lww.Adds)
	assert.Empty(t, lww.Removes)
}

func TestAdd(t *testing.T) {
	lww := NewLWWMap()

	vc1 := vclock.NewVectorClock()
	vc1.Increment("node1")

	lww.Add("key1", "value1", vc1)

	assert.Equal(t, "value1", lww.Adds["key1"].Value)

	vc2 := vclock.NewVectorClock()
	vc2.Increment("node1")
	vc2.Increment("node1")

	lww.Add("key1", "value2", vc2)
	assert.Equal(t, "value2", lww.Adds["key1"].Value)

	vc3 := vclock.NewVectorClock()
	vc3.Increment("node1")

	lww.Add("key1", "value3", vc3)
	assert.Equal(t, "value2", lww.Adds["key1"].Value)
}

func TestRemove(t *testing.T) {
	lww := NewLWWMap()

	vc1 := vclock.NewVectorClock()
	vc1.Increment("node1")

	lww.Remove("key1", "value1", vc1)

	assert.Equal(t, "value1", lww.Removes["key1"].Value)

	vc2 := vclock.NewVectorClock()
	vc2.Increment("node1")
	vc2.Increment("node1")

	lww.Remove("key1", "value2", vc2)
	assert.Equal(t, "value2", lww.Removes["key1"].Value)

	vc3 := vclock.NewVectorClock()
	vc3.Increment("node1")

	lww.Remove("key1", "value3", vc3)
	assert.Equal(t, "value2", lww.Removes["key1"].Value)
}

func TestGetState(t *testing.T) {
	lww := NewLWWMap()

	vc1 := vclock.NewVectorClock()
	vc1.Increment("node1")

	vc2 := vclock.NewVectorClock()
	vc2.Increment("node1")
	vc2.Increment("node2")

	vc3 := vclock.NewVectorClock()
	vc3.Increment("node1")
	vc3.Increment("node1")
	vc3.Increment("node2")

	lww.Add("key1", "value1", vc1)
	lww.Add("key2", "value2", vc2)

	lww.Remove("key1", "value1", vc3)
	lww.Remove("key3", "value3", vc2)

	state := lww.GetState()
	assert.Equal(t, 1, len(state))
	assert.Equal(t, "value2", state["key2"])
	assert.NotContains(t, state, "key1")
	assert.NotContains(t, state, "key3")
}

func TestConcurrentAddRemove(t *testing.T) {
	lww := NewLWWMap()

	vc1 := vclock.NewVectorClock()
	vc1.Increment("node1")

	vc2 := vclock.NewVectorClock()
	vc2.Increment("node1")
	vc2.Increment("node1")

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrent Add and Remove
	go func() {
		lww.Add("key1", "value1", vc1)
		wg.Done()
	}()
	go func() {
		lww.Remove("key1", "value1", vc2)
		wg.Done()
	}()

	wg.Wait()

	state := lww.GetState()
	assert.Empty(t, state)
}
