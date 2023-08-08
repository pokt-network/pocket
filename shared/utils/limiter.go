/*Copyright (c) 2020 Storj Labs, Inc.
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A  PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
* SOFTWARE.
 */

package utils

import (
	"context"
	"sync"
)

// Limiter implements concurrent goroutine limiting.
//
// After calling Wait or Close, no new goroutines are allowed
// to start.
type Limiter struct {
	noCopy noCopy //nolint:structcheck // see noCopy definition

	limit  chan struct{}
	close  sync.Once
	closed chan struct{}
}

// NewLimiter creates a new limiter with limit set to n.
func NewLimiter(n int) *Limiter {
	return &Limiter{
		limit:  make(chan struct{}, n),
		closed: make(chan struct{}),
	}
}

// Go tries to start fn as a goroutine.
// When the limit is reached it will wait until it can run it
// or the context is canceled.
func (limiter *Limiter) Go(ctx context.Context, fn func()) bool {
	if ctx.Err() != nil {
		return false
	}

	select {
	case limiter.limit <- struct{}{}:
	case <-limiter.closed:
		return false
	case <-ctx.Done():
		return false
	}

	go func() {
		defer func() { <-limiter.limit }()
		fn()
	}()

	return true
}

// Wait waits for all running goroutines to finish and
// disallows new goroutines to start.
func (limiter *Limiter) Wait() { limiter.Close() }

// Close waits for all running goroutines to finish and
// disallows new goroutines to start.
func (limiter *Limiter) Close() {
	limiter.close.Do(func() {
		close(limiter.closed)
		// ensure all goroutines are finished
		for i := 0; i < cap(limiter.limit); i++ {
			limiter.limit <- struct{}{}
		}
	})
}
