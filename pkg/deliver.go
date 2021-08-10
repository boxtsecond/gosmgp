package pkg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	SmgpDeliverRespPktLen = HeaderPktLen + 10 + 4 //26d, 0x1a
)

type SmgpDeliverReqPkt struct {
	MsgID      string       // 短消息流水号
	IsReport   uint8        // 短消息流水号
	MsgFormat  uint8        // 短消息格式
	RecvTime   *OctetString // 短消息接收时间
	SrcTermID  *OctetString // 短消息发送号码
	DestTermID *OctetString // 短消息接收号码
	MsgLength  uint8        //  短消息长度
	MsgContent []byte       // 短消息内容
	Reserve    *OctetString // 保留

	// 可选字段
	Options Options

	// used in session
	SequenceID uint32
}

func (p *SmgpDeliverReqPkt) Pack(seqId uint32) ([]byte, error) {
	var pktLen = HeaderPktLen + 77 + uint32(p.MsgLength) + uint32(p.Options.Len())
	var w = newPkgWriter(pktLen)
	// header
	w.WriteHeader(pktLen, seqId, SMGP_DELIVER)
	p.SequenceID = seqId

	// body
	w.WriteFixedSizeString(p.MsgID, 10)
	w.WriteByte(p.IsReport)
	w.WriteByte(p.MsgFormat)
	w.WriteBytes(p.RecvTime.Byte())
	w.WriteBytes(p.SrcTermID.Byte())
	w.WriteBytes(p.DestTermID.Byte())
	w.WriteByte(p.MsgLength)
	w.WriteByte(p.MsgFormat)
	w.WriteBytes(p.MsgContent)
	w.WriteBytes(p.Reserve.Byte())

	for k, o := range p.Options {
		w.WriteByte(uint8(k))
		w.WriteBytes(o)
	}

	return w.Bytes()
}

func (p *SmgpDeliverReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)
	offset := 0

	p.MsgID = hex.EncodeToString(r.ReadCString(10))
	p.IsReport = r.ReadByte()
	p.MsgFormat = r.ReadByte()
	p.RecvTime = &OctetString{
		Data:     r.ReadCString(14),
		FixedLen: 14,
	}
	p.SrcTermID = &OctetString{
		Data:     r.ReadCString(21),
		FixedLen: 21,
	}
	p.DestTermID = &OctetString{
		Data:     r.ReadCString(21),
		FixedLen: 21,
	}
	p.MsgLength = r.ReadByte()
	msgContent := make([]byte, p.MsgLength)
	r.ReadBytes(msgContent)
	p.MsgContent = msgContent
	p.Reserve = &OctetString{
		Data:     r.ReadCString(8),
		FixedLen: 8,
	}
	offset += 10 + 1 + 1 + 14 + 21 + 21 + 1 + int(p.MsgLength) + 8

	options, err := ParseOptions(data[offset:])
	if err != nil {
		return err
	}
	p.Options = options

	return r.Error()
}

func (p *SmgpDeliverReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Deliver Req ---")
	fmt.Fprintln(&b, "MsgID: ", p.MsgID)
	fmt.Fprintln(&b, "IsReport: ", p.IsReport)
	fmt.Fprintln(&b, "MsgFormat: ", p.MsgFormat)
	fmt.Fprintln(&b, "RecvTime: ", p.RecvTime)
	fmt.Fprintln(&b, "SrcTermID: ", p.SrcTermID)
	fmt.Fprintln(&b, "DestTermID: ", p.DestTermID)
	fmt.Fprintln(&b, "MsgLength: ", p.MsgLength)
	fmt.Fprintln(&b, "MsgContent: ", string(p.MsgContent))
	//fmt.Fprintln(&b, "Options: ", p.Options)

	return b.String()
}

type SmgpDeliverRespPkt struct {
	MsgID  string
	Status Status

	// used in session
	SequenceID uint32
}

func (p *SmgpDeliverRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpDeliverRespPktLen)
	// header
	w.WriteHeader(SmgpDeliverRespPktLen, seqId, SMGP_DELIVER_RESP)
	p.SequenceID = seqId

	// body
	w.WriteFixedSizeString(p.MsgID, 10)
	w.WriteInt(binary.BigEndian, p.Status)

	return w.Bytes()
}

func (p *SmgpDeliverRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	// Body: MsgID
	var s = make([]byte, 10)
	r.ReadBytes(s)
	p.MsgID = hex.EncodeToString(s)
	// Body: Status
	r.ReadInt(binary.BigEndian, &p.Status)

	return r.Error()
}

func (p *SmgpDeliverRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Deliver Resp ---")
	fmt.Fprintln(&b, "MsgID: ", p.MsgID)
	fmt.Fprintln(&b, "Status: ", p.Status)
	return b.String()
}
