package levels

import "testing"

func TestRegistryComplete(t *testing.T) {
	all := All()
	if len(all) != 12 {
		t.Fatalf("%d levels", len(all))
	}
	for i, l := range all {
		if l.N != i+1 || l.Title == "" || l.Story == "" || l.Goal == "" || l.Hints[0] == "" || l.Hints[1] == "" || l.Hints[2] == "" {
			t.Fatalf("bad level %d", i+1)
		}
	}
}
