package Map

import "testing"

func TestSearch(t *testing.T) {

	dict := Dictionary {"test": "this is just a test", "unknown": ""}
	assertString := func(t *testing.T, got, want error) {
		t.Helper()
		if got!=want {
			t.Errorf("got error '%s' want '%s'", got, want)
		}
	}


	t.Run("Unknown word", func(t *testing.T) {
		_, got:= dict.Search("unknown")

		assertString(t, got, ErrNotFound)
	})
}

