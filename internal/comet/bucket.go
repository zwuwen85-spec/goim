package comet

import (
	"sync"
	"sync/atomic"

	pb "github.com/Terry-Mao/goim/api/comet"
	"github.com/Terry-Mao/goim/api/protocol"
	"github.com/Terry-Mao/goim/internal/comet/conf"
)

// Bucket is a channel holder.
type Bucket struct {
	c     *conf.Bucket
	cLock sync.RWMutex        // protect the channels for chs
	chs   map[string]*Channel // map sub key to a channel
	// room
	rooms       map[string]*Room // bucket room channels
	routines    []chan *pb.BroadcastRoomReq
	routinesNum uint64

	ipCnts map[string]int32
}

// NewBucket new a bucket struct. store the key with im channel.
func NewBucket(c *conf.Bucket) (b *Bucket) {
	b = new(Bucket)
	b.chs = make(map[string]*Channel, c.Channel)
	b.ipCnts = make(map[string]int32)
	b.c = c
	b.rooms = make(map[string]*Room, c.Room)
	b.routines = make([]chan *pb.BroadcastRoomReq, c.RoutineAmount)
	for i := uint64(0); i < c.RoutineAmount; i++ {
		c := make(chan *pb.BroadcastRoomReq, c.RoutineSize)
		b.routines[i] = c
		go b.roomproc(c)
	}
	return
}

// ChannelCount channel count in the bucket
func (b *Bucket) ChannelCount() int {
	return len(b.chs)
}

// RoomCount room count in the bucket
func (b *Bucket) RoomCount() int {
	return len(b.rooms)
}

// RoomsCount get all room id where online number > 0.
func (b *Bucket) RoomsCount() (res map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	res = make(map[string]int32)
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = room.Online
		}
	}
	b.cLock.RUnlock()
	return
}

// ChangeRoom change to room (backwards compatible - only one room at a time)
// For multi-room support, use JoinRoom/LeaveRoom instead
func (b *Bucket) ChangeRoom(nrid string, ch *Channel) (err error) {
	// change to no room
	if nrid == "" {
		// Leave all rooms
		rooms := ch.GetRooms()
		for roomID := range rooms {
			b.LeaveRoom(ch, roomID)
		}
		ch.Room = nil
		return
	}
	// For multi-room support: just join the new room without leaving old rooms
	// This allows channels to be in multiple rooms simultaneously
	return b.JoinRoom(ch, nrid)
}

// JoinRoom join a room (supports multiple rooms)
func (b *Bucket) JoinRoom(ch *Channel, roomID string) (err error) {
	// Check if already in this room
	if ch.HasRoom(roomID) {
		return nil
	}

	var (
		room *Room
		ok   bool
	)

	b.cLock.Lock()
	if room, ok = b.rooms[roomID]; !ok {
		room = NewRoom(roomID)
		b.rooms[roomID] = room
	}
	b.cLock.Unlock()

	if err = room.Put(ch); err != nil {
		return
	}

	ch.AddRoom(room)
	return
}

// LeaveRoom leave a room (supports multiple rooms)
func (b *Bucket) LeaveRoom(ch *Channel, roomID string) {
	// Check if in this room
	if !ch.HasRoom(roomID) {
		return
	}

	b.cLock.Lock()
	room := b.rooms[roomID]
	b.cLock.Unlock()

	if room != nil && room.Del(ch) {
		b.DelRoom(room)
	}

	ch.RemoveRoom(roomID)
}

// Put put a channel according with sub key.
func (b *Bucket) Put(rid string, ch *Channel) (err error) {
	b.cLock.Lock()
	// close old channel
	if dch := b.chs[ch.Key]; dch != nil {
		dch.Close()
	}
	b.chs[ch.Key] = ch
	b.ipCnts[ch.IP]++
	b.cLock.Unlock()
	// If rid is provided, join the room
	if rid != "" {
		err = b.JoinRoom(ch, rid)
	}
	return
}

// Del delete the channel by sub key.
func (b *Bucket) Del(dch *Channel) {
	b.cLock.Lock()
	if ch, ok := b.chs[dch.Key]; ok {
		if ch == dch {
			delete(b.chs, ch.Key)
		}
		// ip counter
		if b.ipCnts[ch.IP] > 1 {
			b.ipCnts[ch.IP]--
		} else {
			delete(b.ipCnts, ch.IP)
		}
	}
	b.cLock.Unlock()

	// Remove from all rooms
	rooms := dch.GetRooms()
	for roomID := range rooms {
		b.LeaveRoom(dch, roomID)
	}
}

// Channel get a channel by sub key.
func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	ch = b.chs[key]
	b.cLock.RUnlock()
	return
}

// Broadcast push msgs to all channels in the bucket.
func (b *Bucket) Broadcast(p *protocol.Proto, op int32) {
	var ch *Channel
	b.cLock.RLock()
	for _, ch = range b.chs {
		if !ch.NeedPush(op) {
			continue
		}
		_ = ch.Push(p)
	}
	b.cLock.RUnlock()
}

// Room get a room by roomid.
func (b *Bucket) Room(rid string) (room *Room) {
	b.cLock.RLock()
	room = b.rooms[rid]
	b.cLock.RUnlock()
	return
}

// DelRoom delete a room by roomid.
func (b *Bucket) DelRoom(room *Room) {
	b.cLock.Lock()
	delete(b.rooms, room.ID)
	b.cLock.Unlock()
	room.Close()
}

// BroadcastRoom broadcast a message to specified room
func (b *Bucket) BroadcastRoom(arg *pb.BroadcastRoomReq) {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.c.RoutineAmount
	b.routines[num] <- arg
}

// Rooms get all room id where online number > 0.
func (b *Bucket) Rooms() (res map[string]struct{}) {
	var (
		roomID string
		room   *Room
	)
	res = make(map[string]struct{})
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = struct{}{}
		}
	}
	b.cLock.RUnlock()
	return
}

// IPCount get ip count.
func (b *Bucket) IPCount() (res map[string]struct{}) {
	var (
		ip string
	)
	b.cLock.RLock()
	res = make(map[string]struct{}, len(b.ipCnts))
	for ip = range b.ipCnts {
		res[ip] = struct{}{}
	}
	b.cLock.RUnlock()
	return
}

// UpRoomsCount update all room count
func (b *Bucket) UpRoomsCount(roomCountMap map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		room.AllOnline = roomCountMap[roomID]
	}
	b.cLock.RUnlock()
}

// roomproc
func (b *Bucket) roomproc(c chan *pb.BroadcastRoomReq) {
	for {
		arg := <-c
		if room := b.Room(arg.RoomID); room != nil {
			room.Push(arg.Proto)
		}
	}
}
