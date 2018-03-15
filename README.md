# Flaki - Das kleine Generator [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoDoc][godoc-img]][godoc] [![Go Report Card][report-img]][report]

Flaki is an unique ID generator inspired by [Snowflake](https://github.com/twitter/snowflake).
It generates 64-bit unique IDs of type uint64 or string. The string IDs are simply the uint64 IDs represented as a string. Each ID is composed of

| 42-bit timestamp | 2-bit node ID | 5-bit component ID | 15-bit sequence |
---------- | ---------- | ---------- | ---------- |

The node ID and component ID are configured during the creation of the Flaki generator and do
not change afterwards.
The 42-bit timestamp is the number of millisecond elapsed from the start epoch.
If several IDs are requested during the same millisecond, the sequence is incremented to obtain a unique ID every time.
There is a mechanism that does not let the sequence overflow.
If it happen, we wait till the next millisecond to return new IDs. This ensure the IDs uniqueness.

## Usage

Create a new Flaki generator.

```golang
var flaki, err = New()
if err != nil {
    // handle error
}
```

You can configure the Flaki's node ID, component ID and start epoch by submitting options to the call to New.
New takes a variable number of options as parameter.
If no option is given, the following default parameters are used:

* 0 for the node ID
* 0 for the component ID
* 01.01.2017 for the epoch

If you want to modify any of the default parameter, use the corresponding option in the call to New.

* ComponentID(uint64)
* NodeID(uint64)
* StartEpoch(time.Time)

```golang
var cID uint64 = ...
var nID uint64 = ..
var e time.Time = ...

var flaki, err = New(ComponentID(cID), NodeID(nID), StartEpoch(e))
if err != nil {
    // handle error
}
```

To obtain IDs, flaki provides two methods: ```NextID() (uint64, error)```, ```NextIDString() (string, error)```, ```NextValidID() uint64``` and ```NextValidIDString() string```.

```golang
var id uint64
var err error

id, err = flaki.NextID()

id = flaki.NextValidID()

var idStr string

idStr, err = flaki.NextIDString()

idStr = flaki.NextValidIDString()
```

NextID returns either a unique ID or an error if the clock moves backward.

Unlike NextID, NexValidID always returns a valid ID, never an error.
If the clock moves backward, it wait until the situation goes back to normal before returning new IDs.

## Limitations

Flaki won't generate valid IDs after the year 2262.
This is due to the fact that the UnixNano function of the ```package time```
returns undefined result if the Unix time in nanoseconds cannot be represented by an int64, i.e.
a date before the year 1678 or after 2262.

[ci-img]: https://travis-ci.org/cloudtrust/flaki.svg?branch=master
[ci]: https://travis-ci.org/cloudtrust/flaki
[cov-img]: https://coveralls.io/repos/github/cloudtrust/flaki/badge.svg?branch=master
[cov]: https://coveralls.io/github/cloudtrust/flaki?branch=master
[godoc-img]: https://godoc.org/github.com/cloudtrust/flaki?status.svg
[godoc]: https://godoc.org/github.com/cloudtrust/flaki
[report-img]: https://goreportcard.com/badge/github.com/cloudtrust/flaki
[report]: https://goreportcard.com/report/github.com/cloudtrust/flaki
