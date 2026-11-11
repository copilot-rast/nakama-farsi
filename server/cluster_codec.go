package server

import (
	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"google.golang.org/protobuf/proto"
)

func encStream(e *cluster.Enc, s PresenceStream) {
	e.U8(s.Mode).Raw16(s.Subject).Raw16(s.Subcontext).Str(s.Label)
}

func decStream(d *cluster.Dec) PresenceStream {
	return PresenceStream{Mode: d.U8(), Subject: uuid.UUID(d.Raw16()), Subcontext: uuid.UUID(d.Raw16()), Label: d.Str()}
}

func encMeta(e *cluster.Enc, m PresenceMeta) {
	e.U8(uint8(m.Format)).Bool(m.Hidden).Bool(m.Persistence).Str(m.Username).Str(m.Status).Uvarint(uint64(m.Reason))
}

func decMeta(d *cluster.Dec) PresenceMeta {
	return PresenceMeta{Format: SessionFormat(d.U8()), Hidden: d.Bool(), Persistence: d.Bool(), Username: d.Str(), Status: d.Str(), Reason: uint32(d.Uvarint())}
}

func encPresence(e *cluster.Enc, p *Presence) {
	e.Str(p.ID.Node).Raw16(p.ID.SessionID).Raw16(p.UserID)
	encMeta(e, p.Meta)
}

func decPresence(d *cluster.Dec, stream PresenceStream) *Presence {
	p := &Presence{Stream: stream}
	p.ID.Node = d.Str()
	p.ID.SessionID = uuid.UUID(d.Raw16())
	p.UserID = uuid.UUID(d.Raw16())
	p.Meta = decMeta(d)
	return p
}

func marshalEnvelope(env *rtapi.Envelope) ([]byte, error) {
	return proto.Marshal(env)
}
