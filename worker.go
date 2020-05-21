package main

import (
	"context"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type worker struct {
	h []uLoc                   // full history is stored in this slice
	r chan func([]uLoc)        // read only jobs
	w chan func([]uLoc) []uLoc // read/write jobs
}

// NewWorker takes a context.Context which is used to interrupt the job closure execution go routine if necessary
func NewWorker(ctx context.Context) *worker {
	w := worker{
		h: []uLoc{},
		r: make(chan func([]uLoc)),
		w: make(chan func([]uLoc) []uLoc),
	}

	// a single go routine to serialize access to the history data
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-w.w:
				// job writes to history
				w.h = job(w.h)
			case job := <-w.r:
				// job only reads history
				job(w.h)
			}
		}
	}(ctx)

	return &w
}

func (w *worker) Close(context.Context) {}

// WriteHistory pushes a new record into the history slice
func (w *worker) WriteHistory(ctx context.Context, e uLoc) (int, error) {
	r := make(chan int)

	w.w <- func(h []uLoc) []uLoc {
		h = append(h, e)
		r <- len(h)
		return h
	}

	select {
	case <-ctx.Done():
		return -1, ErrContextCancelled
	case id := <-r:
		return id, nil
	}
}

type getHistoryByVisitIDData struct {
	u uLoc
	e error
}

// GetHistoryByVisitID takes a visitId and returns the record found
// GetHistoryByVisitID returns an empty slice when no records are found
func (w *worker) GetHistoryByVisitID(ctx context.Context, vID string) ([]uLocVisit, error) {
	id, err := internalID(vID)
	if err != nil {
		return []uLocVisit{}, err
	}

	r := make(chan getHistoryByVisitIDData)

	w.r <- func(h []uLoc) {
		if id < 1 || int(id) > len(h) { // dont >= because we subtract 1 - we begin at user index 1 rather than 0
			r <- getHistoryByVisitIDData{uLoc{}, ErrIDNotFound}
			return
		}
		r <- getHistoryByVisitIDData{h[id-1], nil} // remember to subtract 1
		return
	}

	select {
	case <-ctx.Done():
		return []uLocVisit{}, ErrContextCancelled
	case ret := <-r:
		if ret.e != nil {
			// return empty array in case of invalid id
			return []uLocVisit{}, nil
		}
		return []uLocVisit{uLocVisit{ret.u, visitID{vID}}}, ret.e
	}
}

// GetHistoryByUserID takes a context.Context because it may need to cancel searching through many records
// GetHistoryByUserID returns an empty slice when no records are found
func (w *worker) GetHistoryByUserID(ctx context.Context, userID, searchString string) ([]uLocVisit, error) {
	r := []uLocVisit{}

	done := make(chan interface{})

	haystack := strings.ToLower(searchString)

	w.r <- func(h []uLoc) {
		for i, entry := range h {
			select {
			case <-ctx.Done():
				return
			default:
				needle := strings.ToLower(entry.Name)
				if strings.Compare(entry.UserID, userID) == 0 &&
					fuzzy.Match(needle, haystack) {
					r = append(r, uLocVisit{entry, externalID(i + 1)})
				}
			}
		}
		done <- nil
		return
	}

	select {
	case <-ctx.Done():
		return []uLocVisit{}, ErrContextCancelled
	case <-done:
		if len(r) <= 5 {
			return r, nil
		}
		return r[len(r)-5:], nil
	}
}
