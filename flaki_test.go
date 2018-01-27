package flaki

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFlaki(t *testing.T) {
	var flaki, err = New()

	assert.Nil(t, err)
	assert.NotNil(t, flaki)
}

func TestSetComponentID(t *testing.T) {
	var flaki *Flaki
	var err error

	// Test with invalid component IDs.
	var invalidComponentIDs = []uint64{maxComponentID + 1, maxComponentID + 2}

	for _, invalidID := range invalidComponentIDs {
		flaki, err = New(ComponentID(invalidID))
		assert.NotNil(t, err)
		assert.Nil(t, flaki)
	}

	// Test with valid component IDs.
	var validComponentIDs = []uint64{0, 1, maxComponentID - 1, maxComponentID}

	for _, validID := range validComponentIDs {
		flaki, err = New(ComponentID(validID))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}
}

func TestSetNodeID(t *testing.T) {
	var flaki *Flaki
	var err error

	// Test with invalid node IDs.
	var invalidNodeIDs = []uint64{maxNodeID + 1, maxNodeID + 2}

	for _, invalidID := range invalidNodeIDs {
		flaki, err = New(NodeID(invalidID))
		assert.NotNil(t, err)
		assert.Nil(t, flaki)
	}

	// Test with valid node IDs.
	var validNodeIDs = []uint64{0, 1, maxNodeID - 1, maxNodeID}

	for _, validID := range validNodeIDs {
		flaki, err = New(NodeID(validID))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}
}

func TestSetEpoch(t *testing.T) {
	var flaki *Flaki
	var err error

	// Test with invalid epoch.
	var invalidEpochs = []time.Time{
		time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC),
		time.Date(2262, 1, 1, 0, 0, 1, 0, time.UTC),
	}

	for _, invalidEpoch := range invalidEpochs {
		flaki, err = New(StartEpoch(invalidEpoch))
		assert.NotNil(t, err)
		assert.Nil(t, flaki)
	}

	// Test with valid epoch.
	var validEpochs = []time.Time{
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2262, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	for _, validEpoch := range validEpochs {
		flaki, err = New(StartEpoch(validEpoch))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}
}

func TestGenerateId(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	var id = flaki.NextValidID()
	assert.True(t, id > 0)
}

func TestIncreasingIds(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	var prevID uint64
	for i := 0; i < 1000; i++ {
		id, err := flaki.NextID()
		assert.True(t, id > prevID)
		assert.Nil(t, err)
		prevID = id
	}
}

func TestUniqueIds(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	var ids = make(map[uint64]bool)

	var count int = 1e6
	for i := 0; i < count; i++ {
		var id = flaki.NextValidID()
		// The ID should be unique, i.e. not in the map.
		_, ok := ids[id]
		assert.False(t, ok)
		ids[id] = true
	}
	assert.Equal(t, len(ids), count)
}

func TestNextIDString(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	var id string
	id, err = flaki.NextIDString()
	assert.Nil(t, err)
	assert.NotZero(t, id)
}

func TestNextIDStringError(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	// Simulate clock that goes backward in time. The second time timeGen is called, it returns
	// a time 2 milliseconds in the past.
	var nbrCallToTimeGen = 0
	var simulatedTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	var timeGen = func() time.Time {
		nbrCallToTimeGen++

		// The 3rd time timeGen is called, we simulate the clock going 2 milliseconds backward.
		switch nbrCallToTimeGen {
		case 2:
			simulatedTime = simulatedTime.Add(-2 * time.Millisecond)
		default:
			simulatedTime = simulatedTime.Add(1 * time.Millisecond)
		}
		return simulatedTime
	}
	flaki.setTimeGen(timeGen)

	// Generate IDs.
	var id string
	id, err = flaki.NextIDString()
	assert.Nil(t, err)
	assert.NotZero(t, id)
	id, err = flaki.NextIDString()
	assert.NotNil(t, err)
	assert.Zero(t, id)
}

func TestNextValidIDString(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	var id = flaki.NextValidIDString()
	assert.NotZero(t, id)
}

func TestConstantTimeStamp(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	// Simulate IDs generation with the same timestamp. The date returned must be after the epoch.
	var constantTimeGen = func() time.Time {
		return time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	flaki.setTimeGen(constantTimeGen)

	var prevID uint64
	prevID, err = flaki.NextID()
	assert.Nil(t, err)

	// When the timestamp is the same, the sequence is incremented to generate the next ID.
	for i := 0; i < 1000; i++ {
		id, err := flaki.NextID()
		assert.Nil(t, err)
		assert.True(t, id == prevID+1)
		prevID = id
	}
}

func TestBackwardTimeShift(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	// Simulate clock that goes backward in time. The 3rd time timeGen is called, it returns
	// a time 2 milliseconds in the past.
	var nbrCallToTimeGen = 0
	var simulatedTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	var timeGen = func() time.Time {
		nbrCallToTimeGen++

		// The 3rd time timeGen is called, we simulate the clock going 2 milliseconds backward.
		switch nbrCallToTimeGen {
		case 3:
			simulatedTime = simulatedTime.Add(-2 * time.Millisecond)
		default:
			simulatedTime = simulatedTime.Add(1 * time.Millisecond)
		}
		return simulatedTime
	}

	flaki.setTimeGen(timeGen)

	// Generate IDs.
	var id uint64
	id, err = flaki.NextID()
	assert.Nil(t, err)
	assert.True(t, id > 0)
	id = flaki.NextValidID()
	assert.True(t, id > 0)

	// Generate new ID. This must returns an error.
	id, err = flaki.NextID()
	assert.NotNil(t, err)
	assert.True(t, id == 0)
	// Here the ID should be valid. If the clock goes backward (timestamp < prevTimestamp),
	// we wait until the situation goes back to normal (timestamp >= prevTimestamp). In this
	// test we wait approximately 2 milliseconds.
	id = flaki.NextValidID()
	assert.True(t, id > 0)
}

func TestTilNextMillis(t *testing.T) {
	var flaki, err = New()
	assert.Nil(t, err)

	// Simulate sequence overflow. We return a constant timestamp and generate more than 2^15 (sequenceMask) new IDs.
	// The expected behavior is that flaki will wait until the next millisecond to return a new unique ID.
	var nbrCallToTimeGen = 0
	var simulatedTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	var timeGen = func() time.Time {
		switch nbrCallToTimeGen {
		case sequenceMask + 3:
			// We must eventually increment the time, otherwise we are stuck in an infinite loop.
			simulatedTime = simulatedTime.Add(1 * time.Second)
		default:
		}
		nbrCallToTimeGen++
		return simulatedTime
	}

	flaki.setTimeGen(timeGen)

	// Generate IDs.
	var prevID uint64
	for i := 0; i < sequenceMask+3; i++ {
		var id, err = flaki.NextID()
		assert.Nil(t, err)
		assert.True(t, id > prevID)
		prevID = id
	}
}

func TestEpochValidity(t *testing.T) {
	var startEpoch = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	var expected = time.Date(2157, 5, 15, 7, 0, 0, 0, time.UTC)

	var result = epochValidity(startEpoch)
	// We set the minutes, seconds and nanoseconds to zero for the comparison.
	result = result.Truncate(1 * time.Hour)

	assert.Equal(t, expected, result)
}

func TestEpochOverflow(t *testing.T) {
	var startEpoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	var maxTimestampValue uint64 = (1 << timestampBits) - 1

	var flaki *Flaki
	{
		var err error
		flaki, err = New(StartEpoch(startEpoch), ComponentID(0), NodeID(0))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}

	// Simulate jump in time to 1 millisecond before the end of epoch validity.
	var simulatedTime = epochValidity(startEpoch).Add(-1 * time.Millisecond)
	var timeGen = func() time.Time {
		simulatedTime = simulatedTime.Add(1 * time.Millisecond)
		return simulatedTime
	}

	flaki.setTimeGen(timeGen)

	// The timestamp part of the ID is about to overflow.
	var id, err = flaki.NextID()
	assert.Nil(t, err)
	assert.Equal(t, (id >> timestampLeftShift), maxTimestampValue)

	// The timestamp part of the ID overflows.
	id, err = flaki.NextID()
	assert.Nil(t, err)
	assert.Equal(t, id, uint64(0))
}

func BenchmarkNextID(b *testing.B) {
	var flaki, err = New()
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		flaki.NextID()
	}
}

func BenchmarkNextValidID(b *testing.B) {
	var flaki, err = New()
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		flaki.NextValidID()
	}
}
