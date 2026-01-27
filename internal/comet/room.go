package comet

import (
	"sync"

	"github.com/Terry-Mao/goim/api/protocol"
	"github.com/Terry-Mao/goim/internal/comet/errors"
)

// Room is a room and store channel room info.
type Room struct {
	ID        string
	rLock     sync.RWMutex
	channels  map[*Channel]struct{} // Use map for multi-room support
	next      *Channel              // Kept for backwards compatibility
	drop      bool
	Online    int32 // dirty read is ok
	AllOnline int32
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id string) (r *Room) {
	r = new(Room)
	r.ID = id
	r.drop = false
	r.channels = make(map[*Channel]struct{})
	r.next = nil
	r.Online = 0
	return
}

// Put put channel into the room.
func (r *Room) Put(ch *Channel) (err error) {
	r.rLock.Lock()
	if !r.drop {
		// Check if already in room
		if _, exists := r.channels[ch]; !exists {
			r.channels[ch] = struct{}{}
			// Update backwards compatible linked list
			ch.Next = r.next
			r.next = ch
			r.Online++
		}
	} else {
		err = errors.ErrRoomDroped
	}
	r.rLock.Unlock()
	return
}

// Del delete channel from the room.
func (r *Room) Del(ch *Channel) bool {
	r.rLock.Lock()
	if _, exists := r.channels[ch]; exists {
		delete(r.channels, ch)
		// Update linked list for backwards compatibility
		// Rebuild the linked list without this channel
		var prev *Channel
		var curr *Channel
		for curr = r.next; curr != nil; curr = curr.Next {
			if curr == ch {
				if prev != nil {
					prev.Next = curr.Next
				} else {
					r.next = curr.Next
				}
				curr.Next = nil
				break
			}
			prev = curr
		}
		r.Online--
	}
	r.drop = r.Online == 0
	r.rLock.Unlock()
	return r.drop
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(p *protocol.Proto) {
	r.rLock.RLock()
	for ch := range r.channels {
		_ = ch.Push(p)
	}
	r.rLock.RUnlock()
}

// Close close the room.
func (r *Room) Close() {
	r.rLock.RLock()
	for ch := range r.channels {
		ch.Close()
	}
	r.rLock.RUnlock()
}

// OnlineNum the room all online.
func (r *Room) OnlineNum() int32 {
	if r.AllOnline > 0 {
		return r.AllOnline
	}
	return r.Online
}
