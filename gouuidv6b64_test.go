package gouuidv6

import (
	"sort"
	"testing"
	"time"
)

func TestB64(t *testing.T) {

	m := make(map[UUIDB64]bool)

	// make a set, storing in a map so we lose the sequence
	for i := 0; i < 1000; i++ {
		u := NewB64()
		m[u] = true
		time.Sleep(time.Microsecond)
		// t.Logf("value: %v", u)
	}

	slb64 := make(UUIDB64Slice, 0, len(m))
	for u := range m {
		slb64 = append(slb64, u)
	}

	sort.Sort(slb64)

	thist := slb64[0].Time()
	for idx, u := range slb64 {
		if u.Time().Before(thist) {
			t.Fatalf("Time %v is before last one %v (idx=%d)", u.Time(), thist, idx)
		}
		// t.Logf("entry: %v / %s", u.Time(), u)
		thist = u.Time()
	}

	// now do the same thing using raw string sorting and make sure the sequence is still good

	sls := make([]string, 0, len(m))
	for u := range m {
		sls = append(sls, u.String())
	}

	sort.Strings(sls)

	thist = time.Time{}
	for idx, s := range sls {
		u, err := ParseB64(s)
		if err != nil {
			t.Fatal(err)
		}
		if u.Time().Before(thist) {
			t.Fatalf("Time %v is before last one %v (idx=%d)", u.Time(), thist, idx)
		}
		// t.Logf("entry: %v / %s", u.Time(), u)
		thist = u.Time()
	}

}
