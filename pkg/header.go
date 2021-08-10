package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const HeaderPktLen uint32 = 4 + 4 + 4

// 消息头(所有消息公共包头)
type Header struct {
	PacketLength uint32 // 数据包长度
	RequestID    uint32 // 请求标识
	SequenceID   uint32 // 消息流水号
}

func (p *Header) Pack(w *pkgWriter, pktLen, requestId, seqId uint32) *pkgWriter {
	w.WriteInt(binary.BigEndian, pktLen)
	w.WriteInt(binary.BigEndian, requestId)
	w.WriteInt(binary.BigEndian, seqId)
	return w
}

func (p *Header) Unpack(r *pkgReader) *Header {
	r.ReadInt(binary.BigEndian, &p.PacketLength)
	r.ReadInt(binary.BigEndian, &p.RequestID)
	r.ReadInt(binary.BigEndian, &p.SequenceID)
	return p
}

func (p *Header) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- Header ---")
	fmt.Fprintln(&b, "Length: ", p.PacketLength)
	fmt.Fprintf(&b, "RequestID: 0x%x\n", p.RequestID)
	fmt.Fprintln(&b, "SequenceID: ", p.SequenceID)

	return b.String()

}
