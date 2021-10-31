// Package gouuidv6 implements "Version 6" UUIDs in Go.
// See http://bradleypeabody.github.io/uuidv6/ UUIDs sort
// correctly by time when naively sorted as raw bytes, have a Time()
// function that returns time the UUID was created and have a reasonable
// guarantee of being globally unique (based on the specifications from
// RFC 4122, with a few intentional exceptions.)
package gouuidv6

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

var bigEnd = binary.BigEndian

// UUID represents a "Version 6" UUID.
type UUID [16]byte

// Compare two UUIDs and return true of their lower 8 bytes
func (u UUID) Compare(to UUID) bool {
	return bigEnd.Uint64(u[:8]) <= bigEnd.Uint64(to[:8])
}

// String returns a textual representation per RFC 4122, e.g. "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// ParseBytes parses a slice of bytes into a UUID struct
func ParseBytes(bs []byte) (UUID, error) {
	var ret UUID
	bigEnd.PutUint64(ret[8:], binary.BigEndian.Uint64(bs[8:]))
	bigEnd.PutUint64(ret[:8], binary.BigEndian.Uint64(bs[:8]))
	return ret, nil
}

// Parse text representation into a UUID struct
func Parse(us string) (UUID, error) {
	var ret UUID
	var v1 uint32
	var v2, v3, v4 uint16
	var v5 uint64 // node
	_, err := fmt.Sscanf(us, "%08x-%04x-%04x-%04x-%012x", &v1, &v2, &v3, &v4, &v5)
	if err != nil {
		return ret, err
	}
	bigEnd.PutUint64(ret[8:], v5)
	bigEnd.PutUint16(ret[8:10], uint16(v4))
	bigEnd.PutUint16(ret[6:8], uint16(v3))
	bigEnd.PutUint16(ret[4:6], uint16(v2))
	bigEnd.PutUint32(ret[:4], uint32(v1))
	return ret, nil
}

// MarshalText returns the String representation of a UUID as a slice of bytes
func (u UUID) MarshalText() ([]byte, error) { return []byte(u.String()), nil }

// UnmarshalText updates a UUID struct using a slice of bytes representing a UUID in string format
func (u *UUID) UnmarshalText(text []byte) (err error) { *u, err = Parse(string(text)); return }

// MarshalBinary returns a UUID as a slice of bytes
func (u UUID) MarshalBinary() ([]byte, error) { return u[:], nil }

// UnmarshalBinary updates a UUID struct using a slice of bytes representing a UUID
func (u *UUID) UnmarshalBinary(data []byte) error { copy(u[:], data); return nil }

// MarshalJSON allows the UUID struct to be seamlessly used as a native json type
func (u UUID) MarshalJSON() ([]byte, error) { return []byte(`"` + u.String() + `"`), nil }

// UnmarshalJSON allows the UUID struct to be seamlessly used as a native json type
func (u *UUID) UnmarshalJSON(data []byte) error {
	s := ""
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*u, err = Parse(s)
	return err
}

// Value allows the UUID struct to be seamlessly used as a native SQL type
func (u UUID) Value() (driver.Value, error) {
	return []byte(u[:]), nil
}

// Scan allows the UUID struct to be seamlessly used as a native SQL type
func (u *UUID) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		copy(u[:], v)
		return nil
	}
	// TODO: should we support strings, even though it's not a good way to go?
	return fmt.Errorf("cannot convert from UUID to sql driver type %T", value)
}

// Bytes returns UUID as byte slice
func (u UUID) Bytes() []byte { return u[:] }

// HighBytes returns the first 8 bytes of a UUID
func (u UUID) HighBytes() []byte { return u[:8] }

// LowBytes returns the last 8 bytes of a UUID
func (u UUID) LowBytes() []byte { return u[8:] }

// IsNil returns true if all UUID bytes are zero
func (u UUID) IsNil() bool { return (bigEnd.Uint64(u[0:8]) | bigEnd.Uint64(u[8:16])) == 0 }

// Time extracts and return the time from the UUID
func (u UUID) Time() time.Time {

	// verify version and variant fields
	if !((u[6]&0xF0) == 0x60 && (u[8]&0xC0) == 0x80) {
		return time.Time{} // return zero time if not a version 6 UUID
	}

	hi := uint64(bigEnd.Uint64(u[:8]))

	// chop the version data out and form the number we want
	t := ((hi >> 4) & 0xFFFFFFFFFFFFF000) | (0x0FFF & hi)

	// convert to nanoseconds
	ut := int64(t-tsoff) * 100

	return time.Unix(ut/int64(time.Second), ut%int64(time.Second))
}

// Node extracts and return the node from the UUID
func (u UUID) Node() uint64 {

	var b []byte = make([]byte, 16)
	copy(b[2:], u[10:])
	i := uint64(bigEnd.Uint64(b))

	return uint64(i)
}

// NewFromTime returns a new UUID set to the given time
func NewFromTime(t time.Time) UUID {

	// NOTE: We intentionally ignore RFC 4122 section 4.2.1.2. and in the case
	// that UUIDs are requested within the same 100-nanosecond time interval,
	// we just increment the clock sequence - the same thing the RFC advises
	// in the case of the clock moving backward (section 4.1.5).

	// get current timestamp
	tsval := tstime(t)

	newlock.Lock()
	// if clock is the same as last time or moved backward, increment clockseq
	if lastts >= tsval {
		clockseq++
	}
	lastts = tsval
	cs := clockseq
	newlock.Unlock()

	var ret UUID

	// shift up 4 bits, mask back in the relevant lower part and set the version
	hi := uint64(((tsval << 4) & 0xFFFFFFFFFFFF0000) | (tsval & 0x0FFF) | 0x6000)

	// 2 bit variant, 14 bits clock sequence, 48 bits node
	lo := (uint64(0x8000) << 48) | (uint64(cs&0x3fff) << 48)
	if alwaysRandomizeNode || node == 0 {
		lo = lo | getRandomNode()
	} else {
		lo = lo | node
	}

	bigEnd.PutUint64(ret[:8], hi)
	bigEnd.PutUint64(ret[8:], lo)

	return ret
}

// New returns a new UUID initialized to a proper value according to "Version 6" rules.
func New() UUID { return NewFromTime(time.Now()) }

func tstime(t time.Time) uint64 { return tsoff + uint64(t.UnixNano()/100) }

// UUID static time offset (see https://play.golang.org/p/pPJd86iZMW)
const tsoff = uint64(122192928000000000)

// lock we use when creating new UUIDs
var newlock sync.Mutex

// last timestamp used
var lastts uint64

// clock sequence value (32-bit so we can use sync/atomic)
var clockseq uint32

// the node part - based on interface MAC address or random
var node uint64

// do we generate new node each time we generate new UUID
var alwaysRandomizeNode bool

func init() {
	b := make([]byte, 8)

	// start with random clock sequence
	if _, err := rand.Read(b); err == nil {
		clockseq = bigEnd.Uint32(b[:4])
	}

	mn := getMacNode()
	if mn != 0 {
		node = mn
	}

	// no node yet, make it random
	RandomizeNode()
}

func getMacNode() uint64 {
	// try to get first interface MAC and use that for node
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if len(i.HardwareAddr) >= 6 {
			return uint64(bigEnd.Uint16(i.HardwareAddr[:2]))<<32 | uint64(bigEnd.Uint32(i.HardwareAddr[2:6]))
		}
	}

	return 0
}

// RandomizeNode sets the 'node' part of the UUID to a random value, instead of using one
// of the MAC addresses from the system.  Use this if you are concerned about
// the privacy aspect of using a MAC address.
func RandomizeNode() {
	node = getRandomNode()
}

func getRandomNode() uint64 {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return 0
	}

	// mask out high 2 bytes and set the multicast bit
	randNode := (bigEnd.Uint64(b[:8]) & 0x0000FFFFFFFFFFFF) | 0x0000010000000000

	return randNode
}

// AlwaysRandomizeNode sets the uuid generation in such way that each uuid has a random
func AlwaysRandomizeNode() {
	alwaysRandomizeNode = true
}

// GetNode returns the node id this instance is using
func GetNode() uint64 {
	return node
}

// SetNode sets the node used for uuidv6's
func SetNode(nodeID uint64) {
	alwaysRandomizeNode = false
	node = nodeID
}
