package vclock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVectorClock(t *testing.T) {
	vc := NewVectorClock()

	assert.NotNil(t, vc)
	assert.Empty(t, vc.Clock)
}

func TestIncrement(t *testing.T) {
	vc := NewVectorClock()
	vc.Increment("node1")

	assert.Equal(t, 1, vc.Clock["node1"])

	vc.Increment("node1")

	assert.Equal(t, 2, vc.Clock["node1"])

	vc.Increment("node2")

	assert.Equal(t, 1, vc.Clock["node2"])
}

func TestMerge(t *testing.T) {
	vc1 := NewVectorClock()

	vc1.Increment("node1")
	vc1.Increment("node1")
	vc1.Increment("node2")

	vc2 := NewVectorClock()

	vc2.Increment("node1")
	vc2.Increment("node3")

	vc1.Merge(vc2)

	assert.Equal(t, 2, vc1.Clock["node1"])
	assert.Equal(t, 1, vc1.Clock["node2"])
	assert.Equal(t, 1, vc1.Clock["node3"])
}

func TestHappensBefore(t *testing.T) {
	vc1 := NewVectorClock()
	vc1.Increment("node1")
	vc1.Increment("node2")

	vc2 := NewVectorClock()
	vc2.Increment("node1")
	vc2.Increment("node1")
	vc2.Increment("node2")

	vc3 := NewVectorClock()
	vc3.Increment("node1")
	vc3.Increment("node2")
	vc3.Increment("node3")

	assert.True(t, vc1.HappensBefore(vc2))
	assert.False(t, vc2.HappensBefore(vc3))
	assert.False(t, vc3.HappensBefore(vc1))
}
