package job

import (
	"context"
	"fmt"
	"time"

	"github.com/Terry-Mao/goim/api/comet"
	pb "github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/api/protocol"
	log "github.com/golang/glog"
)

func (j *Job) push(ctx context.Context, pushMsg *pb.PushMsg) (err error) {
	switch pushMsg.Type {
	case pb.PushMsg_PUSH:
		err = j.pushKeys(pushMsg.Operation, pushMsg.Server, pushMsg.Keys, pushMsg.Msg)
	case pb.PushMsg_ROOM:
		err = j.getRoom(pushMsg.Room).Push(pushMsg.Operation, pushMsg.Msg)
	case pb.PushMsg_BROADCAST:
		err = j.broadcast(pushMsg.Operation, pushMsg.Msg, pushMsg.Speed)
	default:
		err = fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	return
}

// pushKeys push a message to a batch of subkeys.
func (j *Job) pushKeys(operation int32, serverID string, subKeys []string, body []byte) (err error) {
	// Use OpRaw to indicate the body is already encoded
	p := &protocol.Proto{
		Ver:  1,
		Op:   protocol.OpRaw,
		Body: body,  // body is already the JSON string, don't encode it again
	}
	var args = comet.PushMsgReq{
		Keys:    subKeys,
		ProtoOp: operation,
		Proto:   p,
	}
	if c, ok := j.cometServers[serverID]; ok {
		maxRetries := 3 // 最大重试次数
		retryCount := 0
		for retryCount < maxRetries {
			if err = c.Push(&args); err != nil {
				retryCount++
				if retryCount >= maxRetries {
					log.Errorf("c.Push(%v) serverID:%s error(%v) after %d retries, giving up", args, serverID, err, maxRetries)
					break
				}
				log.Warningf("c.Push(%v) serverID:%s error(%v), retry %d/%d", args, serverID, err, retryCount, maxRetries)
				// 短暂延迟后重试
				time.Sleep(time.Millisecond * 10 * time.Duration(retryCount))
				continue
			}
			// 成功，跳出重试循环
			break
		}
		log.Infof("pushKey:%s comets:%d", serverID, len(j.cometServers))
	}
	return
}

// broadcast broadcast a message to all.
func (j *Job) broadcast(operation int32, body []byte, speed int32) (err error) {
	// Use OpRaw to indicate the body is already encoded
	p := &protocol.Proto{
		Ver:  1,
		Op:   protocol.OpRaw,
		Body: body,  // body is already the JSON string, don't encode it again
	}
	comets := j.cometServers
	speed /= int32(len(comets))
	var args = comet.BroadcastReq{
		ProtoOp: operation,
		Proto:   p,
		Speed:   speed,
	}
	for serverID, c := range comets {
		maxRetries := 3 // 最大重试次数
		retryCount := 0
		var lastErr error
		for retryCount < maxRetries {
			if lastErr = c.Broadcast(&args); lastErr != nil {
				retryCount++
				if retryCount >= maxRetries {
					log.Errorf("c.Broadcast(%v) serverID:%s error(%v) after %d retries, giving up", args, serverID, lastErr, maxRetries)
					break
				}
				log.Warningf("c.Broadcast(%v) serverID:%s error(%v), retry %d/%d", args, serverID, lastErr, retryCount, maxRetries)
				// 短暂延迟后重试
				time.Sleep(time.Millisecond * 10 * time.Duration(retryCount))
				continue
			}
			// 成功，跳出重试循环
			break
		}
	}
	log.Infof("broadcast comets:%d", len(comets))
	return
}

// broadcastRoomRawBytes broadcast aggregation messages to room.
func (j *Job) broadcastRoomRawBytes(roomID string, body []byte) (err error) {
	args := comet.BroadcastRoomReq{
		RoomID: roomID,
		Proto: &protocol.Proto{
			Ver:  1,
			Op:   protocol.OpRaw,
			Body: body,
		},
	}
	comets := j.cometServers
	maxRetries := 3 // 最大重试次数
	for serverID, c := range comets {
		retryCount := 0
		for retryCount < maxRetries {
			if err = c.BroadcastRoom(&args); err != nil {
				retryCount++
				if retryCount >= maxRetries {
					log.Errorf("c.BroadcastRoom(%v) roomID:%s serverID:%s error(%v) after %d retries, giving up", args, roomID, serverID, err, maxRetries)
					break
				}
				log.Warningf("c.BroadcastRoom(%v) roomID:%s serverID:%s error(%v), retry %d/%d", args, roomID, serverID, err, retryCount, maxRetries)
				// 短暂延迟后重试
				time.Sleep(time.Millisecond * 10 * time.Duration(retryCount))
				continue
			}
			// 成功，跳出重试循环
			break
		}
	}
	log.Infof("broadcastRoom comets:%d", len(comets))
	return
}
