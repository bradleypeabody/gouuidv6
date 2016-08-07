package gouuidv6

import (
	"encoding/json"
	"runtime"
	"sort"
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

		prefixCounter[prefix] += 1

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

func TestSort(t *testing.T) {

	c := 1 << 18 // 131072
	uuids1 := make(UUIDSlice, 0, c)
	uuids2 := make(UUIDSlice, 0, c)
	times := make([]time.Time, 0, c)
	for i := 0; i < c; i++ {
		u := New()
		uuids1 = append(uuids1, u)
		uuids2 = append(uuids2, u)
		times = append(times, u.Time())
	}

	sort.Sort(uuids1)
	lastt := time.Time{}
	for i := 0; i < c; i++ {
		// make sure they came in sequence
		if uuids1[i] != uuids2[i] {
			t.Fatalf("UUIDs out of sequence at index %d", i)
		}
		ut := times[i]
		if lastt.After(ut) {
			t.Fatalf("Time that should have been prior was after at index %d", i)
		}
		lastt = ut
	}

	// also just throw a few random ones together and make sure sort does
	// the right thing with them
	uuids := make(UUIDSlice, 5)
	uuids[0], _ = Parse(`1e65ced7-cdca-6947-0405-c8bcc8a0b1fd`)
	uuids[1], _ = Parse(`1e65ced7-cdca-694f-0405-c8bcc8a0b1fd`)
	uuids[2], _ = Parse(`1e65ced7-cdcb-679f-0405-c8bcc8a0b1fd`) // last
	uuids[3], _ = Parse(`1e65ced7-cdc6-6e80-0405-c8bcc8a0b1fd`) // first
	uuids[4], _ = Parse(`1e65ced7-cdc6-6e8e-0405-c8bcc8a0b1fd`)

	sort.Sort(uuids)

	if uuids[0].String() != `1e65ced7-cdc6-6e80-0405-c8bcc8a0b1fd` {
		t.Fatalf("Wrong first value, got %s instead", uuids[0].String())
	}

	if uuids[4].String() != `1e65ced7-cdcb-679f-0405-c8bcc8a0b1fd` {
		t.Fatalf("Wrong last value, got %s instead", uuids[4].String())
	}

}
