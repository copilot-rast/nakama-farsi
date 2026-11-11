package server

import (
	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ClusterRouter struct {
	c     *Cluster
	local MessageRouter
}

var _ MessageRouter = (*ClusterRouter)(nil)

func (r *ClusterRouter) SendToPresenceIDs(logger *zap.Logger, presenceIDs []*PresenceID, envelope *rtapi.Envelope, reliable bool) {
	if len(presenceIDs) == 0 {
		return
	}
	var local []*PresenceID
	var remote map[string][]uuid.UUID
	for _, id := range presenceIDs {
		if id.Node == r.c.name || id.Node == "" {
			local = append(local, id)
			continue
		}
		if remote == nil {
			remote = make(map[string][]uuid.UUID, 2)
		}
		remote[id.Node] = append(remote[id.Node], id.SessionID)
	}
	if len(local) > 0 {
		r.local.SendToPresenceIDs(logger, local, envelope, reliable)
	}
	if len(remote) == 0 {
		return
	}
	payload, err := proto.Marshal(envelope)
	if err != nil {
		logger.Error("Could not marshal message for another node", zap.Error(err))
		return
	}
	for node, sessionIDs := range remote {
		e := cluster.NewEnc(len(payload) + 16*len(sessionIDs) + 8).Bool(reliable).Uvarint(uint64(len(sessionIDs)))
		for _, sid := range sessionIDs {
			e.Raw16(sid)
		}
		e.Bytes(payload)
		r.c.send(node, frameEnvelopeSessions, e.Out())
	}
}

func (r *ClusterRouter) SendToStream(logger *zap.Logger, stream PresenceStream, envelope *rtapi.Envelope, reliable bool) {
	r.local.SendToStream(logger, stream, envelope, reliable)
	nodes := r.c.streamNodes(stream)
	if len(nodes) == 0 {
		return
	}
	payload, err := proto.Marshal(envelope)
	if err != nil {
		logger.Error("Could not marshal message for another node", zap.Error(err))
		return
	}
	e := cluster.NewEnc(len(payload) + 48).Bool(reliable)
	encStream(e, stream)
	e.Bytes(payload)
	frame := e.Out()
	for _, node := range nodes {
		if node != r.c.name {
			r.c.send(node, frameEnvelopeStream, frame)
		}
	}
}

func (r *ClusterRouter) SendDeferred(logger *zap.Logger, messages []*DeferredMessage) {
	for _, message := range messages {
		r.SendToPresenceIDs(logger, message.PresenceIDs, message.Envelope, message.Reliable)
	}
}

func (r *ClusterRouter) SendToAll(logger *zap.Logger, envelope *rtapi.Envelope, reliable bool) {
	r.local.SendToAll(logger, envelope, reliable)
	payload, err := proto.Marshal(envelope)
	if err != nil {
		logger.Error("Could not marshal message for other nodes", zap.Error(err))
		return
	}
	r.c.broadcast(frameEnvelopeAll, cluster.NewEnc(len(payload)+8).Bool(reliable).Bytes(payload).Out())
}

func (r *ClusterRouter) handleSessions(payload []byte) {
	d := cluster.NewDec(payload)
	reliable := d.Bool()
	n := int(d.Uvarint())
	if d.Err() != nil || n > d.Len()/16 {
		return
	}
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.UUID(d.Raw16())
	}
	data := d.Bytes()
	if d.Err() != nil {
		return
	}
	r.c.deliver(ids, data, reliable)
}

func (r *ClusterRouter) handleStream(payload []byte) {
	d := cluster.NewDec(payload)
	reliable := d.Bool()
	stream := decStream(d)
	data := d.Bytes()
	if d.Err() != nil {
		return
	}
	r.c.deliver(r.c.localTracker.ListLocalSessionIDByStream(stream), data, reliable)
}

func (r *ClusterRouter) handleAll(payload []byte) {
	d := cluster.NewDec(payload)
	reliable := d.Bool()
	data := d.Bytes()
	if d.Err() != nil {
		return
	}
	env := &rtapi.Envelope{}
	if err := proto.Unmarshal(data, env); err != nil {
		return
	}
	r.local.SendToAll(r.c.logger, env, reliable)
}

func (c *Cluster) deliver(sessionIDs []uuid.UUID, payload []byte, reliable bool) {
	var payloadJSON []byte
	for _, sid := range sessionIDs {
		session := c.sessionRegistry.Get(sid)
		if session == nil {
			continue
		}
		var err error
		switch session.Format() {
		case SessionFormatProtobuf:
			err = session.SendBytes(payload, reliable)
		default:
			if payloadJSON == nil {
				env := &rtapi.Envelope{}
				if err := proto.Unmarshal(payload, env); err != nil {
					c.logger.Error("Could not decode message from another node", zap.Error(err))
					return
				}
				if payloadJSON, err = c.marshaler.Marshal(env); err != nil {
					c.logger.Error("Could not encode message from another node", zap.Error(err))
					return
				}
			}
			err = session.SendBytes(payloadJSON, reliable)
		}
		if err != nil {
			c.logger.Debug("Failed to deliver message from another node", zap.String("sid", sid.String()), zap.Error(err))
		}
	}
}

func (c *Cluster) streamNodes(stream PresenceStream) []string {
	if c.tracker == nil {
		return c.peers()
	}
	return c.tracker.remoteNodes(stream)
}
