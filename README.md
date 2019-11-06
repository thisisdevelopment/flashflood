- [README IS WIP](#readme-is-wip)
- [FlashFlood ring buffer for golang](#flashflood-ring-buffer-for-golang)
- [Introduction](#introduction)
- [Getting Started](#getting-started)
  - [Simple Usage using a channel](#simple-usage-using-a-channel)
- [Benchmarks](#benchmarks)
- [Full Docs and Examples](#full-docs-and-examples)
- [Contributing](#contributing)
- [License](#license)

# README IS WIP 

# FlashFlood ring buffer for golang
[![go report card](https://goreportcard.com/badge/github.com/thisisdevelopment/flashflood "go report card")](https://goreportcard.com/report/github.com/thisisdevelopment/flashflood)
[![codecov](https://codecov.io/gh/thisisdevelopment/flashflood/branch/master/graph/badge.svg)](https://codecov.io/gh/thisisdevelopment/flashflood)
[![CircleCI](https://circleci.com/gh/thisisdevelopment/flashflood.svg?style=svg)](https://circleci.com/gh/thisisdevelopment/flashflood)
[![GoDoc](https://godoc.org/github.com/thisisdevelopment/flashflood?status.svg)](https://godoc.org/github.com/thisisdevelopment/flashflood)

# Introduction
flashflood is a ringbuffer on steroids for golang

The buffer is a traditional ring buffer but has features to receive the flushed out elements on a channel.
You can set the gate amount in order to group elements flushed out and/or perform transformations on them using FuncStack callbacks

# Getting Started

## Simple Usage using a channel
>
    ff := flashflood.New(&flashflood.Opts{})
    ch, err := ff.GetChan()

    ff.Push(obj1, obj2, ... )

    select {
    case v := <-ch:
        fmt.Printf("FLUSHED: %#v\n", v)
    }

# Benchmarks
>   
    BenchmarkPushChan-16                                     4198971               274 ns/op              64 B/op          1 allocs/op
    BenchmarkPushNoChan-16                                   6114704               225 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanGate-16                                 5161064               239 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBuffer-16                            5757730               209 ns/op              64 B/op          1 allocs/op
    BenchmarkWithGet-16                                      4207569               299 ns/op              48 B/op          2 allocs/op
    BenchmarkPushChanBigBufferPow/pow/1-16                   2316964               541 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/2-16                   1958626               576 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/4-16                   1886179               671 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/8-16                   1729645               697 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/16-16                  1709918               702 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/32-16                  1750644               714 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/64-16                  1739887               694 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/128-16                 1797282               696 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/256-16                 1659885               722 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/512-16                 1688109               672 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPow/pow/1024-16                1746681               686 ns/op              64 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/1-16             3602547               327 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/2-16             3611797               333 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/4-16             3471564               357 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/8-16             3412754               357 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/16-16            3300924               355 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/32-16            3306800               359 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/64-16            3264612               369 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/128-16           3335670               374 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/256-16           3207249               376 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/512-16           3243315               371 ns/op              76 B/op          1 allocs/op
    BenchmarkPushChanBigBufferPowCBFuncWithGate/pow/1024-16          3272562               357 ns/op              76 B/op          1 allocs/op

# Full Docs and Examples
https://godoc.org/github.com/thisisdevelopment/flashflood

# Contributing 
You can help to deliver a better flashflood buffer, check out how you can do things [CONTRIBUTING.md](CONTRIBUTING.md)

# License 
© [This is Development BV](https://www.thisisdevelopment.nl), 2019~time.Now()
Released under the [MIT License](https://github.com/thisisdevelopment/flashflood/blob/master/LICENSE)
