package klevdb

import (
	"context"
	"sync/atomic"

	"github.com/klev-dev/kleverr"
)

type OffsetNotify struct {
	nextOffset atomic.Int64
	barrier    chan chan struct{}
}

func NewOffsetNotify(nextOffset int64) *OffsetNotify {
	w := &OffsetNotify{
		barrier: make(chan chan struct{}, 1),
	}

	w.nextOffset.Store(nextOffset)
	w.barrier <- make(chan struct{})

	return w
}

func (w *OffsetNotify) Wait(ctx context.Context, offset int64) error {
	// quick path, just load and check
	if w.nextOffset.Load() > offset {
		return nil
	}

	// acquire current barrier
	b := <-w.barrier

	// probe the current offset
	updated := w.nextOffset.Load() > offset

	// release current barrier
	w.barrier <- b

	// already has a new value, return
	if updated {
		return nil
	}

	// now wait for something to happen
	select {
	case <-b:
		return nil
	case <-ctx.Done():
		return kleverr.Ret(ctx.Err())
	}
}

func (w *OffsetNotify) Set(nextOffset int64) {
	// acquire current barrier
	b := <-w.barrier

	// set the new offset
	if w.nextOffset.Load() < nextOffset {
		w.nextOffset.Store(nextOffset)
	}

	// close the current barrier, e.g. broadcast update
	close(b)

	// create new barrier
	w.barrier <- make(chan struct{})
}
