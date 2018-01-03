# Flaki - Das kleine Generator [![Build Status](https://travis-ci.org/cloudtrust/flaki.svg?branch=master)](https://travis-ci.org/cloudtrust/flaki)[![Go Report](https://goreportcard.com/badge/github.com/cloudtrust/flaki)](https://goreportcard.com/github.com/cloudtrust/flaki)

Flaki is an unique id generator inspired by [Snowflake](https://github.com/twitter/snowflake).
It generates 64-bit unique ids of type uint64. Each id is composed of

| 42-bit timestamp | 2-bit node id | 5-bit component id | 15-bit sequence |
---------- | ---------- | ---------- | ---------- |

The node id and component id are configured during the creation of the Flaki generator and do
not change afterwards.
The 42-bit timestamp is the number of millisecond elapsed from the start epoch.
If several ids are requested during the same millisecond, the sequence is incremented to obtain
a unique id every time.
There is a mechanism that does not let the sequence overflow.
If it happen, we wait till the next millisecond to return new ids. This ensure the ids uniqueness.

## Usage

Create a new Flaki generator.

```golang
var logger log.Logger = ...

var flaki, err = NewFlaki(logger)
if err != nil {
    // handle error
}
```

You can configure the Flaki's node id, component id and start epoch by submitting options to the call to NewFlaki.
NewFlaki takes a variable number of options as parameter.
If no option is given, the following default parameters are used:
* 0 for the node id
* 0 for the component id
* 01.01.2017 for the epoch

If you want to modify any of the default parameter, use the corresponding option in the call to NewFlaki.

* ComponentId(uint64)
* NodeId(uint64)
* StartEpoch(time.Time)

```golang
var logger log.Logger = ...
var cId uint64 = ...
var nId uint64 = ..
var e time.Time = ...

var flaki, err = NewFlaki(logger, ComponentId(cId), NodeId(nId), StartEpoch(e))
if err != nil {
    // handle error
}
```

To obtain id, flaki provides two methods: ```NextId() (uint64, error)``` and ```NextValidId() (uint64)``` 

```golang
var id uint64
var err error

id, err = flaki.NextId()

id = flaki.NextValidId()
```

NextId returns either a unique id or an error if the clock moves backward.

Unlike NextId, NexValidId always returns a valid id, never an error.
If the clock moves backward, it wait until the situation goes back to normal before returning new ids.

## Limitations

Flaki won't generate valid ids after the year 2262.
This is due to the fact that the UnixNano function of the ```package time```
returns undefined result if the Unix time in nanoseconds cannot be represented by an int64, i.e. 
a date before the year 1678 or after 2262.

