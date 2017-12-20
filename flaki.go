package flaki

import (
    "github.com/go-kit/kit/log"
    "time"
    "fmt"
    "sync"
    "github.com/go-kit/kit/log/level"
)

// The ids generated by flaki are composed by
//      - 5 bits for the component id
//      - 2 bits for the node id
//      - 15 bits for the sequence number
//      - 41 bits for the timestamp
const(
    component_id_bits = 5
    node_id_bits = 2
    sequence_bits = 15

    max_component_id = (1 << component_id_bits) - 1
    max_node_id = (1 << node_id_bits) - 1
    sequence_mask = (1 << sequence_bits) -1

    component_id_shift = sequence_bits
    node_id_shift = sequence_bits + component_id_bits
    timestamp_left_shift = sequence_bits + component_id_bits + node_id_bits
)

type Flaki interface {
    NextId() (int64, error)
    NextValidId() (int64)
}

type Generator struct {
    componentId int64
    nodeId int64
    startEpoch time.Time
    lastTimestamp int64
    sequence int64
    timeGen func() time.Time
    logger log.Logger
    mutex *sync.Mutex
}

type option func(*Generator) error

func NewFlaki(logger log.Logger, options ...option) (Flaki, error) {

    if logger == nil {
        return nil, fmt.Errorf("logger must not be nil")
    }

    var flaki = &Generator{
        componentId: 0,
        nodeId: 0,
        startEpoch: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
        lastTimestamp: -1,
        sequence: 0,
        timeGen: time.Now,
        logger: logger,
        mutex: &sync.Mutex{},
    }

    // Apply options to the Generator
    for _, opt := range options {
        var err error = opt(flaki)
        if err != nil {
            return nil, err
        }
    }

    logger.Log("flaki", "general_settings", "component_id_bits", component_id_bits, "node_id_bits", node_id_bits,
        "sequence_bits", sequence_bits)
    logger.Log("flaki", "instance_settings", "component_id", flaki.componentId, "node_id", flaki.nodeId,
        "start_sequence", flaki.sequence, "epoch", formatTime(flaki.startEpoch))

    return flaki, nil
}

func (g *Generator) NextId() (int64, error) {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    var timestamp int64 = g.currentTimeInUnixMillis()
    var prevTimestamp int64 = g.lastTimestamp

    if timestamp < prevTimestamp {
        var err error = fmt.Errorf("clock moved backwards. Refusing to generate ids for %d [ms]", prevTimestamp - timestamp)
        level.Error(g.logger).Log("error", err, "prev_timestamp", prevTimestamp, "timestamp", timestamp)
        return 0, err
    }

    if timestamp == prevTimestamp {
        g.sequence = (g.sequence + 1) & sequence_mask
        if g.sequence == 0 {
            timestamp = g.tilNextMillis(prevTimestamp)
        }
    } else {
        g.sequence = 0
    }

    g.lastTimestamp = timestamp
    var id int64 = ((timestamp - timeToUnixMillis(g.startEpoch)) << timestamp_left_shift) |
        (g.nodeId << node_id_shift) | (g.componentId << component_id_shift) | g.sequence

    level.Debug(g.logger).Log("prev_timestamp", prevTimestamp, "timestamp", timestamp, "sequence", g.sequence, "id", id)
    return id, nil
}

func (g *Generator) NextValidId() (int64) {
    var id int64 = -1
    var err error

    for id == -1 {
        id, err = g.NextId()
        if err != nil {
            level.Warn(g.logger).Log("flaki", "Invalid system clock")
        }
    }

    return id
}

func (g *Generator) tilNextMillis(prevTimestamp int64) int64 {
    var timestamp int64 = g.currentTimeInUnixMillis()

    for timestamp <= prevTimestamp {
        timestamp = g.currentTimeInUnixMillis()
    }
    return timestamp
}

// Return the date till which the generator can generate valid ids
func epochValidity(startEpoch time.Time) time.Time {
    // The number of available bits for the timestamp is 63 (int64 with sign bit always 0) minus
    // the bits used for the node id, component id and sequence.
    var timestampBits uint = 63 - node_id_bits - component_id_bits - sequence_bits
    var durationMilliseconds int64 = (1 << timestampBits) - 1
    var durationNanoseconds int64 = durationMilliseconds * 1e6

    var validityDuration = time.Duration(durationNanoseconds)
    var validUntil = startEpoch.Add(validityDuration)
    return validUntil
}

func (g *Generator) currentTimeInUnixMillis() int64 {
    return timeToUnixMillis(g.timeGen())
}

func timeToUnixMillis(t time.Time) int64 {
    return t.UnixNano() / 1e6
}

// Set the component id for the generator
func ComponentId(id int64) option {
    return func(g *Generator) error {
        return g.setComponentId(id)
    }
}

func (g *Generator) setComponentId(id int64) error {
    // Input validation
    if id < 0 || id > max_component_id {
        var err error = fmt.Errorf("the component id must be in [%d..%d]", 0, max_component_id)
        level.Error(g.logger).Log("error", err, "component_id", id)
        return err
    }
    g.componentId = id
    return nil
}

// Set the node id for the generator
func NodeId(id int64) option {
    return func(g *Generator) error {
        return g.setNodeId(id)
    }
}

func (g *Generator) setNodeId(id int64) error {
    // Input validation
    if id < 0 || id > max_node_id {
        var err error = fmt.Errorf("the node id must be in [%d..%d]", 0, max_node_id)
        level.Error(g.logger).Log("error", err, "node_id", id)
        return err
    }
    g.nodeId = id
    return nil
}

// Set the start epoch
func Epoch(epoch time.Time) option {
    return func(g *Generator) error {
        return g.setStartEpoch(epoch)
    }
}

func (g *Generator) setStartEpoch(epoch time.Time) error {
    // Input validation
    // According to time.Time documentation, UnixNano returns the number of nanoseconds elapsed
    // since January 1, 1970 UTC. The result is undefined if the Unix time in nanoseconds cannot
    // be represented by an int64 (i.e. a date after 2262)
    if epoch.Before(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)) ||
        epoch.After(time.Date(2262, 1, 1, 0, 0, 0, 0, time.UTC)) {

        var err error = fmt.Errorf("the epoch must be between 01.01.1970 and 01.01.2262")
        level.Error(g.logger).Log("error", err, "start_epoch", formatTime(epoch))
        return err
    }

    g.startEpoch = epoch
    return nil
}


// Use in the tests to control the time
func (g *Generator) SetTimeGen(timeGen func() time.Time) {
    g.timeGen = timeGen
}

func (g *Generator) String() string {
    return fmt.Sprintf("Flaki general settings: component id bits = %d, node id bits = %d, sequence bits = %d\n" +
        "Flaki instance settings: component id = %d, node id = %d, start sequence = %d, epoch = %s",
        component_id_bits, node_id_bits, sequence_bits, g.componentId, g.nodeId, g.sequence, formatTime(g.startEpoch))
}

// Format the time: dd-MM-yyyy hh:mm:ss +hhmm UTC
func formatTime(t time.Time) string {
    return t.Format("02-01-2006 15:04:05 -0700 MST")
}
