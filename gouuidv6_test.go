package gouuidv6

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestUUIDSimple(t *testing.T) {

	uuid := New()
	t.Logf("Example UUID: %v (time=`%v`)", uuid, uuid.Time())

	if uuid.IsNil() {
		t.Fatalf("New UUID should never be nil but was")
	}

	if !(UUID{}.IsNil()) {
		t.Fatalf("Empty value should be IsNil() == true but is not!")
	}

	if uuid[6]&0xF0 != 0x60 {
		t.Fatalf("Version number was not 6! (offending byte: %02x)", uuid[7])
	}

	uuid = New()
	tim := time.Now()

	uuidtim := uuid.Time()
	tdiff := tim.Sub(uuidtim)
	if tdiff > time.Millisecond || tdiff < -time.Millisecond {
		t.Fatalf("%v :: Time sample was more than a millisecond away from UUID time: %v vs %v", uuid, tim, uuidtim)
	}

	str := `f81d4fae-7dec-11d0-a765-00a0c91e6bf6`
	uuid, _ = Parse(str)
	if uuid.String() != str {
		t.Fatalf("String conversion did not get expected value, wanted %q, got %q", str, uuid.String())
	}

	// example of uuidv6 generated from another source (and manually pasted in here)
	str2 := `1E65DA3A-36E8-617E-9FCC-C8BCC8A0B17D`
	uuid2, _ := Parse(str2)
	t.Logf("Extracted time: %v", uuid2.Time())

}

func TestUUIDJSON(t *testing.T) {

	uuid := New()

	ex := struct {
		ID UUID `json:"id"`
	}{uuid}

	b, err := json.Marshal(&ex)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != `{"id":"`+uuid.String()+`"}` {
		t.Fatalf("Did not get expected JSON, instead got: %s", b)
	}

	err = json.Unmarshal([]byte(`{
		"id": "1e65c43f-c7c4-47fb-28fc-c8bcc8a0b1fd"
}`), &ex)

	if err != nil {
		t.Fatal(err)
	}

	if ex.ID.String() != "1e65c43f-c7c4-47fb-28fc-c8bcc8a0b1fd" {
		t.Fatalf("Did not get expected ID back from JSON unmarshal, instead got: %s", ex.ID.String())
	}

}

func TestDuplicates(t *testing.T) {

	c := 1 << 18 // 131072

	cpus := runtime.NumCPU()

	allUUIDs := make([][]UUID, cpus)

	wg := &sync.WaitGroup{}

	// make a bunch as fast as possible
	for j := 0; j < cpus; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()

			uuids := make([]UUID, 0, c)

			start := time.Now()
			for i := 0; i < c/cpus; i++ {
				uuids = append(uuids, New())
			}
			t.Logf("Mean time for new UUID: %v", time.Since(start)/time.Duration(c/cpus))

			allUUIDs[j] = uuids

		}(j)
	}
	wg.Wait()

	// concat them all together
	uuids := make([]UUID, 0, c)
	for j := 0; j < cpus; j++ {
		uuids = append(uuids, allUUIDs[j]...)
	}

	// now look for duplicates
	uuidMap := make(map[UUID]bool, c)

	prefixCounter := make(map[string]int)

	for _, u := range uuids {
		if uuidMap[u] {
			t.Fatalf("Was able to get duplicate UUID: %v", u)
		}
		uuidMap[u] = true

		prefix := strings.Join(strings.Split(u.String(), "-")[:3], "-")

		prefixCounter[prefix]++
	}

	max := 0
	maxp := ""
	for prefix, c := range prefixCounter {
		if c > max {
			max = c
			maxp = prefix
		}
	}

	t.Logf("Prefix with highest count was: %q (%d)", maxp, max)

}
