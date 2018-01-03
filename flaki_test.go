package flaki

import (
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestNewFlaki(t *testing.T) {
	var flaki Flaki
	var err error
	var logger log.Logger

	// Test with nil logger (should return an error).
	flaki, err = NewFlaki(logger)
	assert.NotNil(t, err)
	assert.Nil(t, flaki)

	// Test with valid logger.
	f, err := os.Open(os.DevNull)
	assert.Nil(t, err)
	logger = log.NewLogfmtLogger(f)
	flaki, err = NewFlaki(logger)
	assert.Nil(t, err)
	assert.NotNil(t, flaki)
}

func TestSetComponentID(t *testing.T) {
	var flaki Flaki
	var err error

	// Test with invalid component ids.
	var invalidComponentIDs = []uint64{maxComponentID + 1, maxComponentID + 2}

	for _, invalidID := range invalidComponentIDs {
		flaki, err = NewFlaki(getLogger(), ComponentID(invalidID))
		assert.NotNil(t, err)
		assert.Nil(t, flaki)
	}

	// Test with valid component ids.
	var validComponentIDs = []uint64{0, 1, maxComponentID - 1, maxComponentID}

	for _, validID := range validComponentIDs {
		flaki, err = NewFlaki(getLogger(), ComponentID(validID))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}
}

func TestSetNodeID(t *testing.T) {
	var flaki Flaki
	var err error

	// Test with invalid node ids.
	var invalidNodeIDs = []uint64{maxNodeID + 1, maxNodeID + 2}

	for _, invalidID := range invalidNodeIDs {
		flaki, err = NewFlaki(getLogger(), NodeID(invalidID))
		assert.NotNil(t, err)
		assert.Nil(t, flaki)
	}

	// Test with valid node ids.
	var validNodeIDs = []uint64{0, 1, maxNodeID - 1, maxNodeID}

	for _, validID := range validNodeIDs {
		flaki, err = NewFlaki(getLogger(), NodeID(validID))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}
}

func TestSetEpoch(t *testing.T) {
	var flaki Flaki
	var err error

	// Test with invalid epoch.
	var invalidEpochs = []time.Time{
		time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC),
		time.Date(2262, 1, 1, 0, 0, 1, 0, time.UTC),
	}

	for _, invalidEpoch := range invalidEpochs {
		flaki, err = NewFlaki(getLogger(), StartEpoch(invalidEpoch))
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
		flaki, err = NewFlaki(getLogger(), StartEpoch(validEpoch))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}
}

func TestGenerateId(t *testing.T) {
	var flaki = getFlaki(t)
	var id = flaki.NextValidID()
	assert.True(t, id > 0)
}

func TestIncreasingIds(t *testing.T) {
	var flaki = getFlaki(t)

	var prevID uint64
	for i := 0; i < 1000; i++ {
		id, err := flaki.NextID()
		assert.True(t, id > prevID)
		assert.Nil(t, err)
		prevID = id
	}
}

func TestUniqueIds(t *testing.T) {
	var flaki = getFlaki(t)
	var ids = make(map[uint64]bool)

	var count int = 1e6
	for i := 0; i < count; i++ {
		var id = flaki.NextValidID()
		// The id should be unique, i.e. not in the map.
		_, ok := ids[id]
		assert.False(t, ok)
		ids[id] = true
	}
	assert.Equal(t, len(ids), count)
}

func TestFormatTime(t *testing.T) {
	var time = time.Date(2000, 1, 23, 15, 16, 17, 0, time.UTC)
	var expectedString = "23-01-2000 15:16:17 +0000 UTC"
	var actual = formatTime(time)

	assert.Equal(t, expectedString, actual)
}

// We use this interface to access SetTimeGen to control the time for the tests.
type flakiTime interface {
	Flaki
	SetTimeGen(func() time.Time)
}

func TestConstantTimeStamp(t *testing.T) {
	var flaki = getFlaki(t)

	var flakiTime, ok = flaki.(flakiTime)
	assert.True(t, ok)

	// Simulate ids generation with the same timestamp. The date returned must be after the epoch.
	var constantTimeGen = func() time.Time {
		return time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	flakiTime.SetTimeGen(constantTimeGen)

	var prevID, err = flaki.NextID()
	assert.Nil(t, err)

	// When the timestamp is the same, the sequence is incremented to generate the next id.
	for i := 0; i < 1000; i++ {
		id, err := flaki.NextID()
		assert.Nil(t, err)
		assert.True(t, id == prevID+1)
		prevID = id
	}
}

func TestBackwardTimeShift(t *testing.T) {
	var flaki = getFlaki(t)

	var flakiTime, ok = flaki.(flakiTime)
	assert.True(t, ok)

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
	flakiTime.SetTimeGen(timeGen)

	// Generate ids.
	var id, err = flaki.NextID()
	assert.Nil(t, err)
	assert.True(t, id > 0)
	id = flaki.NextValidID()
	assert.True(t, id > 0)

	// Generate new id. This must returns an error.
	id, err = flaki.NextID()
	assert.NotNil(t, err)
	assert.True(t, id == 0)
	// Here the id should be valid. If the clock goes backward (timestamp < prevTimestamp),
	// we wait until the situation goes back to normal (timestamp >= prevTimestamp). In this
	// test we wait approximately 2 milliseconds.
	id = flaki.NextValidID()
	assert.True(t, id > 0)
}

func TestTilNextMillis(t *testing.T) {
	var flaki = getFlaki(t)

	var flakiTime, ok = flaki.(flakiTime)
	assert.True(t, ok)

	// Simulate sequence overflow. We return a constant timestamp and generate more than 2^15 (sequenceMask) new ids.
	// The expected behavior is that flaki will wait until the next millisecond to return a new unique id.
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
	flakiTime.SetTimeGen(timeGen)

	// Generate ids.
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

	var flaki Flaki
	{
		var err error
		flaki, err = NewFlaki(getLogger(), StartEpoch(startEpoch), ComponentID(0), NodeID(0))
		assert.Nil(t, err)
		assert.NotNil(t, flaki)
	}

	var flakiTime, ok = flaki.(flakiTime)
	assert.True(t, ok)

	// Simulate jump in time to 1 millisecond before the end of epoch validity.
	var simulatedTime = epochValidity(startEpoch).Add(-1 * time.Millisecond)
	var timeGen = func() time.Time {
		simulatedTime = simulatedTime.Add(1 * time.Millisecond)
		return simulatedTime
	}
	flakiTime.SetTimeGen(timeGen)

	// The timestamp part of the id is about to overflow.
	var id, err = flaki.NextID()
	assert.Nil(t, err)
	assert.Equal(t, (id >> timestampLeftShift), maxTimestampValue)

	// The timestamp part of the id overflows.
	id, err = flaki.NextID()
	assert.Nil(t, err)
	assert.Equal(t, id, uint64(0))
}

func BenchmarkNextID(b *testing.B) {
	var flaki Flaki
	{
		var err error
		flaki, err = NewFlaki(getLogger())
		assert.Nil(b, err)
	}

	for n := 0; n < b.N; n++ {
		flaki.NextID()
	}
}

func BenchmarkNextValidID(b *testing.B) {
	var flaki Flaki
	{
		var err error
		flaki, err = NewFlaki(getLogger())
		assert.Nil(b, err)
	}

	for n := 0; n < b.N; n++ {
		flaki.NextValidID()
	}
}

func getLogger() log.Logger {
	f, err := os.Open(os.DevNull)
	if err != nil {
		return nil
	}
	return log.NewLogfmtLogger(f)
}

func getFlaki(t *testing.T) Flaki {
	var flaki Flaki
	{
		var err error

		flaki, err = NewFlaki(getLogger())
		assert.Nil(t, err)
	}
	return flaki
}
