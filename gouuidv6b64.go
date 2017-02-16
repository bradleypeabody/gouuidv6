package gouuidv6

import (
	"bytes"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Base64UUIDAlphabet uses the same characters as "url safe" base64, but puts them in ascii sequence so encoded data sorts the same as the raw bytes.
var Base64UUIDAlphabet = `-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz`

// Base64UUIDEncoding is a base64.Encoding that has no padding and uses Base64UUIDAlphabet.
var Base64UUIDEncoding = base64.NewEncoding(Base64UUIDAlphabet).WithPadding(base64.NoPadding)

// UUIDB64 is like UUID but encodes using a "base64 uuid" representation for it's string value (as well as in the database).
type UUIDB64 UUID

type UUIDB64Slice []UUIDB64

func (s UUIDB64Slice) Len() int           { return len(s) }
func (s UUIDB64Slice) Less(i, j int) bool { return bytes.Compare(s[i][:], s[j][:]) < 0 }
func (s UUIDB64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// B64String returns the UUID encoded with Base64UUIDEncoding.
func (u UUIDB64) String() string {
	return Base64UUIDEncoding.EncodeToString(u[:])
}

// Parse base64 text representation
func ParseB64(us string) (UUIDB64, error) {

	ret := UUIDB64{}

	b, err := Base64UUIDEncoding.DecodeString(us)
	if err != nil {
		return ret, err
	}

	copy(ret[:], b)

	return ret, nil

}

func (u UUIDB64) MarshalText() ([]byte, error)           { return []byte(u.String()), nil }
func (u *UUIDB64) UnmarshalText(text []byte) (err error) { *u, err = ParseB64(string(text)); return }

func (u UUIDB64) MarshalBinary() ([]byte, error)     { return u[:], nil }
func (u *UUIDB64) UnmarshalBinary(data []byte) error { copy(u[:], data); return nil }

func (u UUIDB64) MarshalJSON() ([]byte, error) { return []byte(`"` + u.String() + `"`), nil }
func (u *UUIDB64) UnmarshalJSON(data []byte) error {
	s := ""
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*u, err = ParseB64(s)
	return err
}

func (u UUIDB64) Value() (driver.Value, error) {

	return []byte(u.String()), nil

	// return []byte(u[:]), nil
}

func (u *UUIDB64) Scan(value interface{}) error {

	switch v := value.(type) {

	case []byte:

		u2, err := ParseB64(string(v))
		if err != nil {
			return err
		}

		*u = u2

		return nil

	case string:

		u2, err := ParseB64(v)
		if err != nil {
			return err
		}

		*u = u2

		return nil

	}

	// TODO: should we support strings, even though it's not a good way to go?
	return fmt.Errorf("cannot convert from UUIDB64 to sql driver type %T", value)
}

// Return as byte slice.
func (u UUIDB64) Bytes() []byte { return u[:] }

// Return true if all UUIDB64 bytes are zero.
func (u UUIDB64) IsNil() bool { return UUID(u).IsNil() }

// Extract and return the time from the UUIDB64.
func (u UUIDB64) Time() time.Time { return UUID(u).Time() }

func NewB64FromTime(t time.Time) UUIDB64 { return UUIDB64(NewFromTime(t)) }

func NewB64() UUIDB64 { return NewB64FromTime(time.Now()) }
