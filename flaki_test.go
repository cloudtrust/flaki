package flaki

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/go-kit/kit/log"
    "os"
    "fmt"
    "time"
)

func TestNewFlaki(t *testing.T) {
    var flaki Flaki
    var err error
    var logger log.Logger = nil

    // Test with nil logger
    flaki, err = NewFlaki(logger)
    assert.NotNil(t, err)
    assert.Nil(t, flaki)

    // Test with valid logger
    logger = log.NewLogfmtLogger(os.Stdout)
    flaki, err = NewFlaki(logger)
    assert.Nil(t, err)
    assert.NotNil(t, flaki)
}

func TestSetComponentId(t *testing.T) {
    var flaki Flaki
    var err error

    // Test invalid component ids
    var invalidComponentIds []int64 = []int64{-1, max_component_id + 1}

    for _, invalidId := range invalidComponentIds {
        flaki, err = NewFlaki(getLogger(), ComponentId(invalidId))
        assert.NotNil(t, err)
        assert.Nil(t, flaki)
    }

    // Test valid component ids
    var validComponentIds []int64 = []int64{0, 1, max_component_id - 1, max_component_id}

    for _, validId := range validComponentIds {
        flaki, err = NewFlaki(getLogger(), ComponentId(validId))
        assert.Nil(t, err)
        assert.NotNil(t, flaki)
    }
}

func TestSetNodeId(t *testing.T) {
    var flaki Flaki
    var err error

    // Test invalid node ids
    var invalidNodeIds []int64 = []int64{-1, max_node_id + 1}

    for _, invalidId := range invalidNodeIds {
        flaki, err = NewFlaki(getLogger(), NodeId(invalidId))
        assert.NotNil(t, err)
        assert.Nil(t, flaki)
    }

    // Test valid node ids
    var validNodeId []int64 = []int64{0, 1, max_node_id - 1, max_node_id}

    for _, validId := range validNodeId {
        flaki, err = NewFlaki(getLogger(), NodeId(validId))
        assert.Nil(t, err)
        assert.NotNil(t, flaki)
    }
}

func TestSetEpoch(t *testing.T) {
    var flaki Flaki
    var err error

    var invalidEpochs []time.Time = []time.Time{
        time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC),
        time.Date(2262, 1, 1, 0, 0, 1, 0, time.UTC),
    }

    for _, invalidEpoch := range invalidEpochs {
        flaki, err = NewFlaki(getLogger(), Epoch(invalidEpoch))
        assert.NotNil(t, err)
        assert.Nil(t, flaki)
    }

    var validEpochs []time.Time = []time.Time{
        time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Date(2262, 1, 1, 0, 0, 0, 0, time.UTC),
    }

    for _, validEpoch := range validEpochs {
        flaki, err = NewFlaki(getLogger(), Epoch(validEpoch))
        assert.Nil(t, err)
        assert.NotNil(t, flaki)
    }
}

func TestGenerateId(t *testing.T) {
    var flaki Flaki = getFlaki(t)
    var id int64 = flaki.NextValidId()
    var str = fmt.Sprintf("%s",flaki)
    assert.NotNil(t, str)
    assert.True(t, id > 0)
}

func TestIncreasingIds(t *testing.T) {
    var flaki Flaki = getFlaki(t)

    var prevId int64 = 0
    for i := 0; i < 1000; i++ {
        id , err := flaki.NextId()
        assert.True(t, id > prevId)
        assert.Nil(t, err)
    }
}

func TestUniqueIds(t *testing.T) {
    var flaki Flaki = getFlaki(t)
    var ids = make(map[int64]bool)

    var count = 1000000
    for i := 0; i < count; i++ {
        var id int64 = flaki.NextValidId()
        if _, ok := ids[id]; ok {
            fmt.Printf("%d", id)
        } else {
            ids[id] = true
        }
    }
    assert.Equal(t, len(ids), count)
}

func TestFormatTime(t *testing.T) {
    var time time.Time = time.Date(2000, 1, 23, 15, 16, 17, 0, time.UTC)
    var expectedString = "23-01-2000 15:16:17 +0000 UTC"
    var actual = formatTime(time)

    assert.Equal(t, expectedString, actual)
}

type flakiTime interface {
    Flaki
    SetTimeGen(func() time.Time)
}

func TestConstantTimeStamp(t *testing.T) {
    var flaki Flaki = getFlaki(t)

    var flakiTime, ok = flaki.(flakiTime)
    assert.True(t, ok)

    // Simulate ids generation with same timestamp. /!\ the date returned must be after the epoch.
    var constantTimeGen = func() time.Time {
        return time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
    }
    flakiTime.SetTimeGen(constantTimeGen)

    var prevId, err = flaki.NextId()
    assert.Nil(t, err)

    // When the timestamp is the same, the sequence is incremented to generate the next id
    for i := 0; i < 1000; i++ {
        id , err := flaki.NextId()
        assert.Nil(t, err)
        assert.True(t, id == prevId+1)
        prevId = id
    }
}

func TestBackwardTimeShift(t *testing.T) {
    var flaki Flaki = getFlaki(t)

    var flakiTime, ok = flaki.(flakiTime)
    assert.True(t, ok)

    // Simulate clock that goes backward in time. The 3rd time timeGen is called, it returns
    // a time in the past.
    var nbrCallToTimeGen = 0
    var simulatedTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
    var timeGen = func() time.Time {
        nbrCallToTimeGen++

        // The 3rd time timeGen is called, we simulate the clock going 2 seconds backward
        switch nbrCallToTimeGen {
        case 3:
            simulatedTime = simulatedTime.Add(-2 * time.Second)
        default:
            simulatedTime = simulatedTime.Add(1 * time.Second)
        }
        return simulatedTime
    }
    flakiTime.SetTimeGen(timeGen)

    // Generate ids
    var id, err = flaki.NextId()
    assert.Nil(t, err)
    assert.True(t, id > 0)
    id = flaki.NextValidId()
    assert.True(t, id > 0)

    // Generate new id. This must returns an error
    id, err = flaki.NextId()
    assert.NotNil(t, err)
    assert.True(t, id == -1)
    // Here the id should be valid. If the clock goes backward (timestamp < prevTimestamp),
    // we wait until the situation goes back to normal (timestamp >= prevTimestamp). In this
    // test we wait approximately 2 seconds.
    id = flaki.NextValidId()
    assert.True(t, id > 0)
}

func TestTilNextMillis(t *testing.T) {
    var flaki Flaki = getFlaki(t)

    var flakiTime, ok = flaki.(flakiTime)
    assert.True(t, ok)

    // Simulate sequence overflow. We return a constant timestamp and generate more than sequence_mask new ids.
    // The expected behavior is that flaki will wait until the timestamp increment to return a new id. We increment
    // it the (sequence_mask + 2) time timeGen is called, so we are not stuck in an infinite loop.
    var nbrCallToTimeGen = 0
    var simulatedTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
    var timeGen = func() time.Time {
        switch nbrCallToTimeGen {
        case sequence_mask + 3:
            simulatedTime = simulatedTime.Add(1 * time.Second)
        default:
        }
        nbrCallToTimeGen++
        return simulatedTime
    }
    flakiTime.SetTimeGen(timeGen)

    // Generate ids
    var prevId int64 = -1
    for i := 0; i < sequence_mask + 3; i++ {
        var id , err = flaki.NextId()
        assert.Nil(t, err)
        assert.True(t, id > prevId)
        prevId = id
    }
}

func TestEpochValidity(t *testing.T) {
    var startEpoch = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
    var expected = time.Date(2087, 9, 7, 15, 0, 0, 0, time.UTC)

    var result = epochValidity(startEpoch)
    // We set the minutes, seconds and nanoseconds to zero for the comparison
    result = result.Truncate(1 * time.Hour)

    assert.Equal(t, expected, result)
}

func BenchmarkNextId(b *testing.B) {
    var flaki Flaki
    {
        var err error
        flaki, err = NewFlaki(getLogger())
        assert.Nil(b, err)
    }

    for n := 0; n < b.N; n++ {
        flaki.NextId()
    }

}

func BenchmarkNextValidId(b *testing.B) {
    var flaki Flaki
    {
        var err error
        flaki, err = NewFlaki(getLogger())
        assert.Nil(b, err)
    }

    for n := 0; n < b.N; n++ {
        flaki.NextValidId()
    }
}

func getLogger() log.Logger {
    f, err := os.Open(os.DevNull)
    if err != nil {
        return nil
    }
    var logger= log.NewLogfmtLogger(os.Stdout)
    logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
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
