package flaki

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/go-kit/kit/log"
    "os"
    "fmt"
    "time"
)

func TestX(t *testing.T) {
    var flaki Flaki
    var err error

    flaki, err = NewFlaki(getLogger(), ComponentId(max_component_id), NodeId(max_node_id))
    assert.Nil(t, err)

    flaki.NextId()
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
    var validComponentIds []int64 = []int64{0, 1, max_component_id-1, max_component_id}

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
    var validNodeId []int64 = []int64{0, 1, max_node_id-1, max_node_id}

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

    // Simulate clock that goes backward in time: we set a starting time in the future, then
    // we switch back to the actual time.
    var timeInFuture = time.Now().Add(1 * time.Second)
    var futureTimeGen = func() time.Time {
        return timeInFuture
    }
    flakiTime.SetTimeGen(futureTimeGen)

    // Generate id
    var id, err = flaki.NextId()
    assert.Nil(t, err)
    assert.NotZero(t, id)

    // Go backward in time
    flakiTime.SetTimeGen(time.Now)

    // Generate new id. This must returns an error
    id, err = flaki.NextId()
    assert.NotNil(t, err)
    assert.Zero(t, id)
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