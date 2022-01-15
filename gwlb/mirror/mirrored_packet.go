package mirror

type Type int

const (
	TypeUnknown = Type(iota)
	TypePreRewrite
	TypePostRewrite
)

type Packet struct {
	Packet []byte
	Type   Type
}

func New(pkt []byte, pre bool) Packet {
	mirrored := make([]byte, len(pkt))
	copy(mirrored, pkt)

	if pre {
		return Packet{Packet: mirrored, Type: TypePreRewrite}
	} else {
		return Packet{Packet: mirrored, Type: TypePostRewrite}
	}
}
