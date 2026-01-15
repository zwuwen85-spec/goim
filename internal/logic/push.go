package logic

import (
	"context"
	"fmt"
	"os"

	"github.com/Terry-Mao/goim/internal/logic/model"

	log "github.com/golang/glog"
)

// PushKeys push a message by keys.
func (l *Logic) PushKeys(c context.Context, op int32, keys []string, msg []byte) (err error) {
	fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys START: op=%d keys=%v msg_len=%d ===\n", op, keys, len(msg))
	servers, err := l.dao.ServersByKeys(c, keys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys ServersByKeys ERROR: %v ===\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys servers: %v ===\n", servers)
	pushKeys := make(map[string][]string)
	for i, key := range keys {
		server := servers[i]
		fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys key[%d]=%s server=%s ===\n", i, key, server)
		if server != "" && key != "" {
			pushKeys[server] = append(pushKeys[server], key)
		}
	}
	fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys pushKeys map: %+v ===\n", pushKeys)
	for server, serverKeys := range pushKeys {
		fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys calling PushMsg: server=%s keys=%v ===\n", server, serverKeys)
		if err = l.dao.PushMsg(c, op, server, serverKeys, msg); err != nil {
			fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys PushMsg ERROR: %v ===\n", err)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "=== LOGIC PushKeys SUCCESS ===\n")
	return
}

// PushMids push a message by mid.
func (l *Logic) PushMids(c context.Context, op int32, mids []int64, msg []byte) (err error) {
	keyServers, _, err := l.dao.KeysByMids(c, mids)
	if err != nil {
		return
	}
	keys := make(map[string][]string)
	for key, server := range keyServers {
		if key == "" || server == "" {
			log.Warningf("push key:%s server:%s is empty", key, server)
			continue
		}
		keys[server] = append(keys[server], key)
	}
	for server, keys := range keys {
		if err = l.dao.PushMsg(c, op, server, keys, msg); err != nil {
			return
		}
	}
	return
}

// PushRoom push a message by room.
func (l *Logic) PushRoom(c context.Context, op int32, typ, room string, msg []byte) (err error) {
	return l.dao.BroadcastRoomMsg(c, op, model.EncodeRoomKey(typ, room), msg)
}

// PushAll push a message to all.
func (l *Logic) PushAll(c context.Context, op, speed int32, msg []byte) (err error) {
	return l.dao.BroadcastMsg(c, op, speed, msg)
}
