// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package util

import (
	"sync"
	"sync/atomic"
	"time"
)

type atomic_AtomicInteger int32

type TimerPool struct {
	timeout  time.Duration
	pool     sync.Pool
	TotalNew atomic_AtomicInteger
	TotalGet atomic_AtomicInteger
}

func NewTimerPool(timeout time.Duration) *TimerPool {
	var totalNew atomic_AtomicInteger
	p := &TimerPool{
		timeout:  timeout,
		TotalNew: totalNew,
		TotalGet: 0,
	}
	p.pool.New = func() interface{} {
		t := time.NewTimer(timeout)
		atomic.AddInt32((*int32)(&totalNew), 1)
		return t
	}

	return p
}

func (tp *TimerPool) AcquireTimer() *time.Timer {
	tv := tp.pool.Get()
	t := tv.(*time.Timer)
	t.Reset(tp.timeout)
	atomic.AddInt32((*int32)(&tp.TotalGet), 1)
	return t
}

func (tp *TimerPool) ReleaseTimer(t *time.Timer) {
	if !t.Stop() {
		// Collect possibly added time from the channel
		// if timer has been stopped and nobody collected its' value.
		select {
		case <-t.C:
		default:
		}
	}
	tp.pool.Put(t)
}
