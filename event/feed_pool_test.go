// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package event

import (
	"sync"
	"testing"
	"time"
)

func TestFeedPool(t *testing.T) {
	var feedpool FeedPool
	var subscribed, quit sync.WaitGroup
	type ackedEvent struct {
		ack chan<- struct{}
	}

	done := make(chan struct{})
	subscriber := func() {
		defer quit.Done()
		subchan := make(chan ackedEvent)
		sub := feedpool.Subscribe(subchan)
		defer sub.Unsubscribe()
		subscribed.Done()

		for {
			select {
			case ev := <-subchan:
				ev.ack <- struct{}{}
			case <-done:
				return
			}
		}
	}

	checkNumberAcked := func(acksignal chan struct{}, nsub int) {
		timeout := time.NewTimer(1 * time.Second)
		nacked := 0
		for {
			select {
			case <-acksignal:
				nacked++
				if nacked == nsub {
					return
				}

			case <-timeout.C:
				if nacked != nsub {
					t.Errorf("received %d acks, want %d", nacked, nsub)
				}
				return
			}
		}
	}

	// test sending value with different number of subscribers
	nsub := 0
	for i := 0; i < 2; i++ {
		subscribed.Add(1)
		go subscriber()
		subscribed.Wait()
		nsub++
		acksignal := make(chan struct{})
		if nsent := feedpool.Send(ackedEvent{acksignal}); nsent != nsub {
			t.Errorf("send delivered %d times, want %d", nsent, nsub)
		}
		checkNumberAcked(acksignal, nsub)
	}

	// test sending value after subscribers unsubscribed events
	quit.Add(nsub)
	close(done)
	quit.Wait()

	acksignal := make(chan struct{})
	if nsent := feedpool.Send(ackedEvent{acksignal}); nsent != 0 {
		t.Errorf("send delivered %d times, want 0", nsent)
	}
	checkNumberAcked(acksignal, 0)

	// test sending value but without any subscriber
	if nsent := feedpool.Send(99); nsent != 0 {
		t.Errorf("send delivered %d times, want 0", nsent)
	}
}
