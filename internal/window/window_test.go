package window

import (
	"testing"
	"time"

	"github.com/example/search-trends/internal/stoplist"
)

func TestWindowAddAndTop(t *testing.T) {
	sl := stoplist.New()
	w := New(3, sl)
	now := time.Now()

	w.Add("shoes", now)
	w.Add("shoes", now)
	w.Add("hat", now)
	w.Add("hat", now)
	w.Add("hat", now)
	w.Add("socks", now)
	w.Add("socks", now)

	w.recalcTop()
	top := w.Top()
	if len(top) != 3 {
		t.Fatalf("expected 3 items, got %d", len(top))
	}
	if top[0].Query != "hat" || top[0].Count != 3 {
		t.Errorf("top item wrong: %+v", top[0])
	}
	if top[1].Query != "shoes" || top[1].Count != 2 {
		t.Errorf("second item wrong: %+v", top[1])
	}
}

func TestWindowSlide(t *testing.T) {
	sl := stoplist.New()
	w := New(2, sl)
	now := time.Now()
	old := now.Add(-6 * time.Minute)
	w.Add("old", old)
	w.Add("recent", now)
	w.slide()
	w.recalcTop()
	top := w.Top()
	if len(top) != 1 || top[0].Query != "recent" {
		t.Errorf("expected only 'recent', got %+v", top)
	}
}

func TestStoplistFilter(t *testing.T) {
	sl := stoplist.New()
	sl.Add("bad")
	w := New(2, sl)
	now := time.Now()
	w.Add("good", now)
	w.Add("bad", now)
	w.Add("good", now)
	w.recalcTop()
	top := w.Top()
	for _, q := range top {
		if q.Query == "bad" {
			t.Error("stoplist word should not appear")
		}
	}
}
