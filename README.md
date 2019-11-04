- [README IS WIP](#readme-is-wip)
- [FlashFlood ring buffer for golang](#flashflood-ring-buffer-for-golang)
- [Introduction](#introduction)
- [Getting Started](#getting-started)
  - [Simple Usage using a channel](#simple-usage-using-a-channel)
- [Contributing](#contributing)
- [License](#license)

# README IS WIP 

# FlashFlood ring buffer for golang
[![go report card](https://goreportcard.com/badge/github.com/thisisdevelopment/flashflood "go report card")](https://goreportcard.com/badge/github.com/thisisdevelopment/flashflood)
[![codecov](https://codecov.io/gh/thisisdevelopment/flashflood/branch/master/graph/badge.svg)](https://codecov.io/gh/thisisdevelopment/flashflood)
[![CircleCI](https://circleci.com/gh/thisisdevelopment/flashflood.svg?style=svg)](https://circleci.com/gh/thisisdevelopment/flashflood)



# Introduction
flashflood is a ringbuffer on steroids for golang

The buffer is a traditional ring buffer but has features to receive the flushed out elements on a channel.
You can set the gate amount in order to group elements flushed out and/or perform transformations on them using FuncStack callbacks

# Getting Started

## Simple Usage using a channel

>
    	ff := flashflood.New(&flashflood.Opts{})
        ch, err := ff.GetChan()

        select {
        case v := <-ch:
            fmt.Printf("FLUSHED: %#v\n", v)
        }

# Contributing 
You can help to deliver a better flashflood buffer, check out things you can do [CoC.md](CoC.md)

# License 
© [This is Development BV](https://www.thisisdevelopment.nl), 2019~time.Now
Released under the [MIT License](https://github.com/thisisdevelopment/flashflood/blob/master/LICENSE)
