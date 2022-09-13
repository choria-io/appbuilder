// Copyright (c) 2017-2021, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package exec

// https://blog.gopheracademy.com/advent-2014/backoff/

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// policy implements a backoff policy, randomizing its delays
// and saturating at the final value in Millis.
type policy struct {
	Millis []int64
}

func newPolicy(steps uint, min time.Duration, max time.Duration) (*policy, error) {
	p := policy{}

	if steps == 0 {
		return nil, fmt.Errorf("steps must be more than 0")
	}
	if min == 0 {
		return nil, fmt.Errorf("minimum retry can not be 0")
	}
	if max == 0 {
		return nil, fmt.Errorf("maximum retry can not be 0")
	}

	if max < min {
		max, min = min, max
	}

	stepSize := uint(max-min) / steps
	for i := uint(0); i < steps; i += 1 {
		p.Millis = append(p.Millis, (min + time.Duration(i*stepSize).Round(time.Millisecond)).Milliseconds())
	}

	return &p, nil
}

// duration returns the time duration of the n'th wait cycle in a
// backoff policy. This is b.Millis[n], randomized to avoid thundering
// herds.
func (b policy) duration(n int) time.Duration {
	if n >= len(b.Millis) {
		n = len(b.Millis) - 1
	}

	return time.Duration(jitter(b.Millis[n])) * time.Millisecond
}

// sleep sleeps for the duration t and can be interrupted by ctx. An error
// is returns if the context cancels the sleep
func (b policy) sleep(ctx context.Context, t time.Duration) error {
	timer := time.NewTimer(t)

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	}
}

// jitter returns a random integer uniformly distributed in the range
// [0.5 * millis .. 1.5 * millis]
func jitter(millis int64) int64 {
	if millis == 0 {
		return 0
	}

	return millis/2 + rand.Int63n(millis)
}
