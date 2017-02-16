// Implements "Version 6" UUIDs in Go.  See http://bradleypeabody.github.io/uuidv6/
// UUIDs sort correctly by time when naively sorted as raw bytes, have a Time()
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

// "Version 6" UUID.
type UUID [16]byte

// Slice of UUIDs, sorts using first 64bits (where the time is)
type UUIDSlice []UUID

func (s UUIDSlice) Len() int { return len(s) }

// FIXME: Not sure this is smart - why wouldn't we just sort by the entire binary value, to ensure consistency...
func (s UUIDSlice) Less(i, j int) bool { return bigEnd.Uint64(s[i][:8]) < bigEnd.Uint64(s[j][:8]) }

func (s UUIDSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Textual representation per RFC 4122, e.g. "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// Parse text representation
func Parse(us string) (UUID, error) {
	var ret UUID
	var v1, v2, v3, v4, v5 uint64
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

func (u UUID) MarshalText() ([]byte, error)           { return []byte(u.String()), nil }
func (u *UUID) UnmarshalText(text []byte) (err error) { *u, err = Parse(string(text)); return }

func (u UUID) MarshalBinary() ([]byte, error)     { return u[:], nil }
func (u *UUID) UnmarshalBinary(data []byte) error { copy(u[:], data); return nil }

func (u UUID) MarshalJSON() ([]byte, error) { return []byte(`"` + u.String() + `"`), nil }
func (u *UUID) UnmarshalJSON(data []byte) error {
	s := ""
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*u, err = Parse(s)
	return err
}

func (u UUID) Value() (driver.Value, error) {
	return []byte(u[:]), nil
}

func (u *UUID) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		copy(u[:], v)
		return nil
	}
	// TODO: should we support strings, even though it's not a good way to go?
	return fmt.Errorf("cannot convert from UUID to sql driver type %T", value)
}

// Return as byte slice.
func (u UUID) Bytes() []byte { return u[:] }

// Return true if all UUID bytes are zero.
func (u UUID) IsNil() bool { return (bigEnd.Uint64(u[0:8]) | bigEnd.Uint64(u[8:16])) == 0 }

// Extract and return the time from the UUID.
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
	lo := (uint64(0x8000) << 48) | (uint64(cs&0x3fff) << 48) | node

	bigEnd.PutUint64(ret[:8], hi)
	bigEnd.PutUint64(ret[8:], lo)

	return ret

}

// Return a new UUID initialized to a proper value according to "Version 6" rules.
func New() UUID { return NewFromTime(time.Now()) }

// Returns a timestamp appropriate for UUID time
func ts() uint64 { return tsoff + uint64(time.Now().UnixNano()/100) }

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

func init() {

	b := make([]byte, 8)

	// start with random clock sequence
	rand.Read(b)
	clockseq = bigEnd.Uint32(b[:4])

	// try to get first interface MAC and use that for node
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if len(i.HardwareAddr) >= 6 {
			node = uint64(bigEnd.Uint16(i.HardwareAddr[:2]))<<32 | uint64(bigEnd.Uint32(i.HardwareAddr[2:6]))
			break
		}
	}

	// no node yet, make it random
	if node == 0 {
		RandomizeNode()
	}

}

// Set the 'node' part of the UUID to a random value, instead of using one
// of the MAC addresses from the system.  Use this if you are concerned about
// the privacy aspect of using a MAC address.
func RandomizeNode() {
	b := make([]byte, 8)
	rand.Read(b)
	// mask out high 2 bytes and set the multicast bit
	node = (bigEnd.Uint64(b[:8]) & 0x0000FFFFFFFFFFFF) | 0x0000010000000000
}
