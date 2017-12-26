# flaki - Das kleine Generator [![Build Status](https://travis-ci.org/cloudtrust/flaki.svg?branch=master)](https://travis-ci.org/cloudtrust/flaki)

Flaki is an unique id generator inspired by [Snowflake](https://github.com/twitter/snowflake).

## Installation

## Configuration


g.logger.Log("warning", "flaki won't generate valid ids after 01.01.2262")

// It returns unique positive ids of type int64.
// The id is composed of: 5-bit component id, 2-bit node id, 15-bit sequence number, and
// 42-bit time's milliseconds since the epoch.