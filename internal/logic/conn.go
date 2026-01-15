package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Terry-Mao/goim/api/protocol"
	"github.com/Terry-Mao/goim/internal/logic/model"
	log "github.com/golang/glog"
	"github.com/google/uuid"
)

// Connect connected a conn.
func (l *Logic) Connect(c context.Context, server, cookie string, token []byte) (mid int64, key, roomID string, accepts []int32, hb int64, err error) {
	fmt.Fprintf(os.Stderr, "=== LOGIC Connect START: server=%s token=%s ===\n", server, string(token))
	log.Infof("Connect called: server=%s cookie=%s token=%s", server, cookie, string(token))
	var params struct {
		Mid      int64   `json:"mid"`
		Key      string  `json:"key"`
		RoomID   string  `json:"room_id"`
		Platform string  `json:"platform"`
		Accepts  []int32 `json:"accepts"`
	}
	if err = json.Unmarshal(token, &params); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		fmt.Fprintf(os.Stderr, "=== LOGIC JSON UNMARSHAL ERROR: %v ===\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "=== LOGIC Connect parsed: mid=%d key=%s roomID=%s ===\n", params.Mid, params.Key, params.RoomID)
	log.Infof("Connect parsed: mid=%d key=%s roomID=%s", params.Mid, params.Key, params.RoomID)
	mid = params.Mid
	roomID = params.RoomID
	accepts = params.Accepts
	hb = int64(l.c.Node.Heartbeat) * int64(l.c.Node.HeartbeatMax)
	if key = params.Key; key == "" {
		key = uuid.New().String()
	}
	fmt.Fprintf(os.Stderr, "=== LOGIC before AddMapping: mid=%d key=%s server=%s ===\n", mid, key, server)
	log.Infof("Connect before AddMapping: mid=%d key=%s server=%s", mid, key, server)
	if err = l.dao.AddMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		fmt.Fprintf(os.Stderr, "=== LOGIC AddMapping ERROR: %v ===\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "=== LOGIC Connect SUCCESS: mid=%d key=%s ===\n", mid, key)
	log.Infof("conn connected key:%s server:%s mid:%d token:%s", key, server, mid, token)
	return
}

// Disconnect disconnect a conn.
func (l *Logic) Disconnect(c context.Context, mid int64, key, server string) (has bool, err error) {
	if has, err = l.dao.DelMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.DelMapping(%d,%s) error(%v)", mid, key, server)
		return
	}
	log.Infof("conn disconnected key:%s server:%s mid:%d", key, server, mid)
	return
}

// Heartbeat heartbeat a conn.
func (l *Logic) Heartbeat(c context.Context, mid int64, key, server string) (err error) {
	has, err := l.dao.ExpireMapping(c, mid, key)
	if err != nil {
		log.Errorf("l.dao.ExpireMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if !has {
		if err = l.dao.AddMapping(c, mid, key, server); err != nil {
			log.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	log.Infof("conn heartbeat key:%s server:%s mid:%d", key, server, mid)
	return
}

// RenewOnline renew a server online.
func (l *Logic) RenewOnline(c context.Context, server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &model.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.dao.AddServerOnline(context.Background(), server, online); err != nil {
		return nil, err
	}
	return l.roomCount, nil
}

// Receive receive a message.
func (l *Logic) Receive(c context.Context, mid int64, proto *protocol.Proto) (err error) {
	log.Infof("receive mid:%d message:%+v", mid, proto)
	return
}
