// Flaki - Das kleine Generator
// Flaki is a unique id generator inspired by Snowflake (https://github.com/twitter/snowflake).
// It returns unique ids of type uint64.
// The id is composed of: 5-bit component id, 2-bit node id, 15-bit sequence number, and
// 42-bit time's milliseconds since the epoch.
// They will be valid until 139 years 4 months and a few days after the epoch. After that, there
// will be an overflow and the newly generated ids won't be unique anymore.

package flaki

import (
    "fmt"
    "sync"
    "time"
    "github.com/go-kit/kit/log"
)

const(
    componentIdBits = 5
    nodeIdBits = 2
    sequenceBits = 15
    timestampBits = 64 - componentIdBits - nodeIdBits - sequenceBits

    maxComponentId = (1 << componentIdBits) - 1
    maxNodeId = (1 << nodeIdBits) - 1
    sequenceMask = (1 << sequenceBits) -1

    componentIdShift = sequenceBits
    nodeIdShift = sequenceBits + componentIdBits
    timestampLeftShift = sequenceBits + componentIdBits + nodeIdBits
)

type Flaki interface {
    NextId() (uint64, error)
    NextValidId() (uint64)
}

type Generator struct {
    componentId uint64
    nodeId uint64
    lastTimestamp int64
    sequence uint64
    logger log.Logger
    mutex *sync.Mutex

    // startEpoch is the reference time from which we count the elapsed time.
    // The default is 01.01.2017 00:00:00 +0000 UTC.
    startEpoch time.Time

    // timeGen is the function that returns the current time.
    timeGen func() time.Time
}

// option type, takes one argument: the Generator we are operating on.
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

    // Apply options to the Generator.
    for _, opt := range options {
        var err error = opt(flaki)
        if err != nil {
            logger.Log("error", err)
            return nil, err
        }
    }

    logger.Log("msg", "flaki general settings", "component_id_bits", componentIdBits, "node_id_bits", nodeIdBits,
        "sequence_bits", sequenceBits)
    logger.Log("msg", "flaki instance settings", "component_id", flaki.componentId, "node_id", flaki.nodeId,
        "start_sequence", flaki.sequence, "epoch", formatTime(flaki.startEpoch))

    return flaki, nil
}

// Generate a unique id.
// If the clock moves backward, it returns an error.
// If more than 2^15 id are requested in one millisecond, we wait for the next millisecond
// to generate them, otherwise the sequence overflows and this lead to duplicate ids.
func (g *Generator) NextId() (uint64, error) {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    var timestamp int64 = g.currentTimeInUnixMillis()
    var prevTimestamp int64 = g.lastTimestamp

    if timestamp < prevTimestamp {
        var err error = fmt.Errorf("clock moved backwards. Refusing to generate ids for %d [ms]", prevTimestamp - timestamp)
        g.logger.Log("error", err, "prev_timestamp", prevTimestamp, "timestamp", timestamp)
        return 0, err
    }

    if timestamp == prevTimestamp {
        g.sequence = (g.sequence + 1) & sequenceMask
        if g.sequence == 0 {
            timestamp = g.tilNextMillis(prevTimestamp)
        }
    } else {
        g.sequence = 0
    }

    g.lastTimestamp = timestamp
    var id uint64 = (uint64(timestamp - timeToUnixMillis(g.startEpoch)) << timestampLeftShift) |
        (g.nodeId << nodeIdShift) | (g.componentId << componentIdShift) | g.sequence

    g.logger.Log("msg", "id generated", "prev_timestamp", prevTimestamp, "timestamp", timestamp, "sequence", g.sequence, "id", id)
    return id, nil
}

// Generate a valid unique id.
// It always returns a valid identifier, never an error.
// If the clock moves backward, we wait until the situation goes back to normal.
func (g *Generator) NextValidId() (uint64) {
    var id uint64
    var err error = fmt.Errorf("")

    for err != nil {
        id, err = g.NextId()
        if err != nil {
            g.logger.Log("msg", "invalid system clock")
        }
    }

    return id
}

// Wait till the next millisecond.
func (g *Generator) tilNextMillis(prevTimestamp int64) int64 {
    var timestamp int64 = g.currentTimeInUnixMillis()

    for timestamp <= prevTimestamp {
        timestamp = g.currentTimeInUnixMillis()
    }
    return timestamp
}

// Return the date till which the generator can generate valid ids.
func epochValidity(startEpoch time.Time) time.Time {
    var durationMilliseconds int64 = (1 << timestampBits) - 1
    var durationNanoseconds int64 = durationMilliseconds * 1e6

    var validityDuration = time.Duration(durationNanoseconds)
    var validUntil = startEpoch.Add(validityDuration)
    return validUntil
}

// Returns the current time in milliseconds.
func (g *Generator) currentTimeInUnixMillis() int64 {
    return timeToUnixMillis(g.timeGen())
}

func timeToUnixMillis(t time.Time) int64 {
    return t.UnixNano() / 1e6
}

// Set the component id for the generator.
func ComponentId(id uint64) option {
    return func(g *Generator) error {
        return g.setComponentId(id)
    }
}

func (g *Generator) setComponentId(id uint64) error {
    if id > maxComponentId {
        var err error = fmt.Errorf("the component id must be in [%d..%d]", 0, maxComponentId)
        g.logger.Log("error", err, "component_id", id)
        return err
    }
    g.componentId = id
    return nil
}

// Set the node id for the generator.
func NodeId(id uint64) option {
    return func(g *Generator) error {
        return g.setNodeId(id)
    }
}

func (g *Generator) setNodeId(id uint64) error {
    if id > maxNodeId {
        var err error = fmt.Errorf("the node id must be in [%d..%d]", 0, maxNodeId)
        g.logger.Log("error", err, "node_id", id)
        return err
    }
    g.nodeId = id
    return nil
}

// Set the start epoch.
func Epoch(epoch time.Time) option {
    return func(g *Generator) error {
        return g.setStartEpoch(epoch)
    }
}

func (g *Generator) setStartEpoch(epoch time.Time) error {
    // According to time.Time documentation, UnixNano returns the number of nanoseconds elapsed
    // since January 1, 1970 UTC. The result is undefined if the Unix time in nanoseconds cannot
    // be represented by an int64 (i.e. a date after 2262).
    if epoch.Before(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)) ||
        epoch.After(time.Date(2262, 1, 1, 0, 0, 0, 0, time.UTC)) {

        var err error = fmt.Errorf("the epoch must be between 01.01.1970 and 01.01.2262")
        g.logger.Log("error", err, "start_epoch", formatTime(epoch))
        return err
    }
    g.startEpoch = epoch
    return nil
}

// Set the function that returns the current time. It is used in the tests
// to control the time.
func (g *Generator) SetTimeGen(timeGen func() time.Time) {
    g.timeGen = timeGen
}

// Format the time: dd-MM-yyyy hh:mm:ss +hhmm UTC.
func formatTime(t time.Time) string {
    return t.Format("02-01-2006 15:04:05 -0700 MST")
}
