package comet

import (
	"fmt"
	"os"
	"sync"

	"github.com/Terry-Mao/goim/api/protocol"
	"github.com/Terry-Mao/goim/internal/comet/errors"
	"github.com/Terry-Mao/goim/pkg/bufio"
)

// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	Room     *Room // For backwards compatibility, deprecated
	Rooms    map[string]*Room // Support multiple rooms
	roomLock sync.RWMutex // Protect Rooms map
	CliProto Ring
	signal   chan *protocol.Proto
	Writer   bufio.Writer
	Reader   bufio.Reader
	Next     *Channel
	Prev     *Channel

	Mid      int64
	Key      string
	IP       string
	watchOps map[int32]struct{}
	mutex    sync.RWMutex
}

// NewChannel new a channel.
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)
	c.signal = make(chan *protocol.Proto, svr)
	c.watchOps = make(map[int32]struct{})
	c.Rooms = make(map[string]*Room)
	return c
}

// Watch watch a operation.
func (c *Channel) Watch(accepts ...int32) {
	fmt.Fprintf(os.Stderr, "=== Channel Watch: accepts=%v ===\n", accepts)
	c.mutex.Lock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	c.mutex.Unlock()
	fmt.Fprintf(os.Stderr, "=== Channel Watch: watchOps now has %d items ===\n", len(c.watchOps))
}

// UnWatch unwatch an operation
func (c *Channel) UnWatch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		delete(c.watchOps, op)
	}
	c.mutex.Unlock()
}

// NeedPush verify if in watch.
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	if _, ok := c.watchOps[op]; ok {
		c.mutex.RUnlock()
		return true
	}
	c.mutex.RUnlock()
	return false
}

// Push server push message.
func (c *Channel) Push(p *protocol.Proto) (err error) {
	fmt.Fprintf(os.Stderr, "=== Channel Push: key=%s protoOp=%d bodyLen=%d ===\n", c.Key, p.Op, len(p.Body))
	select {
	case c.signal <- p:
		fmt.Fprintf(os.Stderr, "=== Channel Push: SUCCESS sent to signal ===\n")
	default:
		fmt.Fprintf(os.Stderr, "=== Channel Push: FAILED signal full ===\n")
		err = errors.ErrSignalFullMsgDropped
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() *protocol.Proto {
	return <-c.signal
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- protocol.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- protocol.ProtoFinish
}

// AddRoom add a room to the channel
func (c *Channel) AddRoom(room *Room) {
	c.roomLock.Lock()
	if c.Rooms == nil {
		c.Rooms = make(map[string]*Room)
	}
	c.Rooms[room.ID] = room
	// For backwards compatibility, keep Room field updated
	c.Room = room
	c.roomLock.Unlock()
}

// RemoveRoom remove a room from the channel
func (c *Channel) RemoveRoom(roomID string) {
	c.roomLock.Lock()
	delete(c.Rooms, roomID)
	// Update backwards compatible Room field
	if len(c.Rooms) > 0 {
		for _, room := range c.Rooms {
			c.Room = room
			break
		}
	} else {
		c.Room = nil
	}
	c.roomLock.Unlock()
}

// HasRoom check if channel is in a room
func (c *Channel) HasRoom(roomID string) bool {
	c.roomLock.RLock()
	_, ok := c.Rooms[roomID]
	c.roomLock.RUnlock()
	return ok
}

// GetRooms get all rooms the channel is in
func (c *Channel) GetRooms() map[string]*Room {
	c.roomLock.RLock()
	defer c.roomLock.RUnlock()
	// Return a copy to avoid race conditions
	rooms := make(map[string]*Room, len(c.Rooms))
	for k, v := range c.Rooms {
		rooms[k] = v
	}
	return rooms
}
