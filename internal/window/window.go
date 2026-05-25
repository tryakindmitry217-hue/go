package window

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/search-trends/internal/stoplist"
)

const windowSeconds = 300

type QueryCount struct {
	Query string `json:"query"`
	Count uint64 `json:"count"`
}

type bucket struct {
	mu     sync.RWMutex
	counts map[string]uint64
}

type Window struct {
	buckets   [windowSeconds]*bucket
	topN      atomic.Value
	stopList  *stoplist.StopList
	topSize   int
}

func New(topSize int, sl *stoplist.StopList) *Window {
	w := &Window{
		topSize:  topSize,
		stopList: sl,
	}
	for i := 0; i < windowSeconds; i++ {
		w.buckets[i] = &bucket{counts: make(map[string]uint64)}
	}
	w.topN.Store([]QueryCount{})
	return w
}

func (w *Window) Add(query string, ts time.Time) {
	if time.Since(ts) > windowSeconds*time.Second {
		return
	}
	slot := uint64(ts.Unix()) % windowSeconds
	b := w.buckets[slot]
	b.mu.Lock()
	b.counts[query]++
	b.mu.Unlock()
}

func (w *Window) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.slide()
			w.recalcTop()
		case <-ctx.Done():
			return
		}
	}
}

func (w *Window) slide() {
	now := time.Now().Unix()
	oldSlot := (uint64(now) - windowSeconds) % windowSeconds
	b := w.buckets[oldSlot]
	b.mu.Lock()
	b.counts = make(map[string]uint64)
	b.mu.Unlock()
}

func (w *Window) recalcTop() {
	total := make(map[string]uint64)
	for _, b := range w.buckets {
		b.mu.RLock()
		for q, c := range b.counts {
			total[q] += c
		}
		b.mu.RUnlock()
	}

	sl := w.stopList.Words()
	for _, word := range sl {
		delete(total, word)
	}

	top := topNFromMap(total, w.topSize)
	w.topN.Store(top)
}

func (w *Window) Top() []QueryCount {
	v := w.topN.Load()
	if v == nil {
		return nil
	}
	return v.([]QueryCount)
}

type item struct {
	query string
	count uint64
}

type minHeap []item

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool   { return h[i].count < h[j].count }
func (h minHeap) Swap(i, j int)        { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{})  { *h = append(*h, x.(item)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func topNFromMap(counts map[string]uint64, n int) []QueryCount {
	if n <= 0 {
		return nil
	}
	h := &minHeap{}
	heap.Init(h)
	for q, c := range counts {
		if h.Len() < n {
			heap.Push(h, item{q, c})
		} else if c > (*h)[0].count {
			heap.Pop(h)
			heap.Push(h, item{q, c})
		}
	}
	result := make([]QueryCount, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		it := heap.Pop(h).(item)
		result[i] = QueryCount{Query: it.query, Count: it.count}
	}
	return result
}
