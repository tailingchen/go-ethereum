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
	"reflect"
	"sync"
)

// FeedPool maintians a map to manage all feeds have been subscribed.
type FeedPool struct {
	once  sync.Once // ensures that init only runs once
	mu    sync.RWMutex
	feeds map[reflect.Type]*Feed
}

func (fp *FeedPool) init() {
	fp.feeds = make(map[reflect.Type]*Feed)
}

// Subscribe adds a channel to the feed by the type of channel.
func (fp *FeedPool) Subscribe(channel interface{}) Subscription {
	fp.once.Do(fp.init)

	fp.mu.Lock()
	chanval := reflect.ValueOf(channel)
	chantyp := chanval.Type()
	efeed := fp.feeds[chantyp.Elem()]
	if efeed == nil {
		efeed = new(Feed)
		fp.feeds[chantyp.Elem()] = efeed
	}
	fp.mu.Unlock()

	return efeed.Subscribe(channel)
}

// Send delivers value through responding feed.
// It returns zero if can't find responding feed that means no one subscribes the value.
// Otherwise, returns the number of subscribers that the value was sent to.
func (fp *FeedPool) Send(value interface{}) (nsent int) {
	fp.once.Do(fp.init)

	fp.mu.RLock()
	rvalue := reflect.ValueOf(value)
	efeed := fp.feeds[rvalue.Type()]
	fp.mu.RUnlock()

	if efeed == nil {
		return 0
	}

	return efeed.Send(value)
}
