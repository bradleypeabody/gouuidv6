[![Go Report Card](https://goreportcard.com/badge/github.com/kai5263499/gouuidv6)](https://goreportcard.com/report/github.com/kai5263499/gouuidv6)
![lint](https://github.com/kai5263499/gouuidv6/actions/workflows/lint.yml/badge.svg)
![test](https://github.com/kai5263499/gouuidv6/actions/workflows/test.yml/badge.svg)
![race](https://github.com/kai5263499/gouuidv6/actions/workflows/race.yml/badge.svg)
![static](https://github.com/kai5263499/gouuidv6/actions/workflows/static.yml/badge.svg)

Reference implementation of the draft "Version 6" UUID proposal in Go
=======================

## Draft proposal

The draft proposal of the UUIDv6 spec along with examples and rationale is available [here](https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format).

Another informative source which contains reference implementations for generating UUIDv6's in go and converting UUIv1 to UUIDv6 in Python is available [here](http://gh.peabody.io/uuidv6/).

# Byte layout

Bytes 0-7: (each digit shown is hex, 4 bits)
```
    00 00 00 00  00 00 00 00
    |                | ||  |
     ----------------  ||  |
    timestamp          ||  |
    bits 59-12         ||  |
                 version|  |
                         --
                  timestamp
                  bits 11-0
```

Bytes 8-15: (same as RFC 4122)
```
    00 00 00 00  00 00 00 00
    ||  | |                |
    ||  |  ________________
    ||  |       node
    | --
    | clock seq bits 11-0
    2 bits variant, 2 bits
    are 13-12 of clock seq
```

# Usage

Usage is pretty simple, the only variable is the node id used. The most basic usage is to allow the library to use the MAC address as the node id.

```go
id := gouuidv6.New()
id.String()
```

An alternative is to use a randomized node ID
```go
gouuidv6.AlwaysRandomizeNode()
id := gouuidv6.New()
id.String()
```
