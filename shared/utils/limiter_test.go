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

package utils_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pokt-network/pocket/shared/utils"
)

func TestLimiterLimiting(t *testing.T) {
	t.Parallel()

	const N, Limit = 1000, 10
	ctx := context.Background()
	limiter := utils.NewLimiter(Limit)
	counter := int32(0)
	for i := 0; i < N; i++ {
		limiter.Go(ctx, func() {
			if atomic.AddInt32(&counter, 1) > Limit {
				panic("limit exceeded")
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&counter, -1)
		})
	}
	limiter.Close()
}

func TestLimiterCanceling(t *testing.T) {
	t.Parallel()

	const N, Limit = 1000, 10
	limiter := utils.NewLimiter(Limit)

	ctx, cancel := context.WithCancel(context.Background())

	counter := int32(0)

	waitForCancel := make(chan struct{}, N)
	block := make(chan struct{})
	allreturned := make(chan struct{})

	go func() {
		for i := 0; i < N; i++ {
			limiter.Go(ctx, func() {
				if atomic.AddInt32(&counter, 1) > Limit {
					panic("limit exceeded")
				}

				waitForCancel <- struct{}{}
				<-block
			})
		}
		close(allreturned)
	}()

	for i := 0; i < Limit; i++ {
		<-waitForCancel
	}
	cancel()
	<-allreturned
	close(block)

	limiter.Close()
	if counter > Limit {
		t.Fatal("too many times run")
	}

	started := limiter.Go(context.Background(), func() {
		panic("should not start")
	})
	if started {
		t.Fatal("should not start")
	}
}
