// Package flaki provides the implementation of Flaki - Das kleine Generator
// Flaki is a unique ID generator inspired by Snowflake (https://github.com/twitter/snowflake).
// It returns unique IDs of type uint64. The ID is composed of: 5-bit component ID, 2-bit node ID,
// 15-bit sequence number, and 42-bit time's milliseconds since the epoch.
// Unique IDs will be generated until 139 years 4 months and a few days after the epoch. After that, there
// will be an overflow and the newly generated IDs won't be unique anymore.
package flaki

import (
	"sync"

	"fmt"
	"strconv"
	"time"
)

const (
	componentIDBits  = 5
	nodeIDNodeIDBits = 2
	sequenceBits     = 15
	timestampBits    = 64 - componentIDBits - nodeIDNodeIDBits - sequenceBits

	maxComponentID = (1 << componentIDBits) - 1
	maxNodeID      = (1 << nodeIDNodeIDBits) - 1
	sequenceMask   = (1 << sequenceBits) - 1

	componentIDShift   = sequenceBits
	nodeIDNodeIDShift  = sequenceBits + componentIDBits
	timestampLeftShift = sequenceBits + componentIDBits + nodeIDNodeIDBits
)

// Flaki is the interface of the unique ID generator.
type Flaki interface {
	NextID() (uint64, error)
	NextIDString() (string, error)
	NextValidID() uint64
	NextValidIDString() string
}

// Generator is the unique ID generator. Create a generator with NewFlaki.
type Generator struct {
	componentID   uint64
	nodeIDNodeID  uint64
	lastTimestamp int64
	sequence      uint64
	mutex         *sync.Mutex

	// startEpoch is the reference time from which we count the elapsed time.
	// The default is 01.01.2017 00:00:00 +0000 UTC.
	startEpoch time.Time

	// timeGen is the function that returns the current time.
	timeGen func() time.Time
}

// option type is use to configure the Flaki generator. It takes one argument: the Generator we are operating on.
type option func(*Generator) error

// NewFlaki returns a new unique ID generator.
//
// If you do not specify options, Flaki will use the following
// default parameters: 0 for the node ID, 0 for the component ID,
// and 01.01.2017 as start epoch.
//
// To change the default settings, use the options in the call
// to NewFlaki, i.e. NewFlaki(logger, ComponentID(1), NodeID(2), StartEpoch(startEpoch))
func NewFlaki(options ...option) (Flaki, error) {

	var flaki = &Generator{
		componentID:   0,
		nodeIDNodeID:  0,
		startEpoch:    time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
		lastTimestamp: -1,
		sequence:      0,
		timeGen:       time.Now,
		mutex:         &sync.Mutex{},
	}

	// Apply options to the Generator.
	for _, opt := range options {
		var err = opt(flaki)
		if err != nil {
			return nil, err
		}
	}

	return flaki, nil
}

// NextID returns a new unique ID, or an error if the clock moves backward.
func (g *Generator) NextID() (uint64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	var timestamp = g.currentTimeInUnixMillis()
	var prevTimestamp = g.lastTimestamp

	if timestamp < prevTimestamp {
		return 0, fmt.Errorf("clock moved backwards. Refusing to generate IDs for %d [ms]", prevTimestamp-timestamp)
	}

	// If too many IDs (more than 2^sequenceBits) are requested in a given time unit (millisecond),
	// the sequence overflows. If it happens, we wait till the next time unit to generate new IDs,
	// otherwise we end up with duplicates.
	if timestamp == prevTimestamp {
		g.sequence = (g.sequence + 1) & sequenceMask
		if g.sequence == 0 {
			timestamp = g.tilNextMillis(prevTimestamp)
		}
	} else {
		g.sequence = 0
	}

	g.lastTimestamp = timestamp
	var id = (uint64(timestamp-timeToUnixMillis(g.startEpoch)) << timestampLeftShift) |
		(g.nodeIDNodeID << nodeIDNodeIDShift) | (g.componentID << componentIDShift) | g.sequence

	return id, nil
}

// NextIDString returns the NextID as a string.
func (g *Generator) NextIDString() (string, error) {
	var id, err = g.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

// NextValidID always returns a new unique ID, it never returns an error.
// If the clock moves backward, it waits until the situation goes back to normal
// and then returns the valid ID.
func (g *Generator) NextValidID() uint64 {
	var id uint64
	var err = fmt.Errorf("")

	for err != nil {
		id, err = g.NextID()
		if err != nil {
			// Do nothing: we wait until we get a valid ID
		}
	}

	return id
}

// NextValidIDString returns the NextValidID as a string.
func (g *Generator) NextValidIDString() string {
	var id = g.NextValidID()
	return strconv.FormatUint(id, 10)
}

// tilNextMillis waits until the next millisecond.
func (g *Generator) tilNextMillis(prevTimestamp int64) int64 {
	var timestamp = g.currentTimeInUnixMillis()

	for timestamp <= prevTimestamp {
		timestamp = g.currentTimeInUnixMillis()
	}
	return timestamp
}

// epochValidity returns the date till which Flaki can generate valid IDs.
func epochValidity(startEpoch time.Time) time.Time {
	var durationMilliseconds int64 = (1 << timestampBits) - 1
	var durationNanoseconds = durationMilliseconds * 1e6

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

// ComponentID is the option used to set the component ID in the call
// to NewFlaki.
func ComponentID(id uint64) option {
	return func(g *Generator) error {
		return g.setComponentID(id)
	}
}

func (g *Generator) setComponentID(id uint64) error {
	if id > maxComponentID {
		return fmt.Errorf("the component id must be in [%d..%d]", 0, maxComponentID)
	}
	g.componentID = id
	return nil
}

// NodeID is the option used to set the node ID in the call to NewFlaki.
func NodeID(id uint64) option {
	return func(g *Generator) error {
		return g.setNodeID(id)
	}
}

func (g *Generator) setNodeID(id uint64) error {
	if id > maxNodeID {
		return fmt.Errorf("the node id must be in [%d..%d]", 0, maxNodeID)
	}
	g.nodeIDNodeID = id
	return nil
}

// StartEpoch is the option used to set the epoch in the call to NewFlaki.
func StartEpoch(epoch time.Time) option {
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
		return fmt.Errorf("the epoch must be between 01.01.1970 and 01.01.2262")
	}
	g.startEpoch = epoch
	return nil
}

// SetTimeGen set the function that returns the current time. It is used in the tests
// to control the time.
func (g *Generator) SetTimeGen(timeGen func() time.Time) {
	g.timeGen = timeGen
}

// Format the time: dd-MM-yyyy hh:mm:ss +hhmm UTC.
func formatTime(t time.Time) string {
	return t.Format("02-01-2006 15:04:05 -0700 MST")
}
