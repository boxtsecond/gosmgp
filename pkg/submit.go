package pkg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	SmgpSubmitRespPktLen = HeaderPktLen + 10 + 4 //26d, 0x1a
)

type SmgpSubmitReqPkt struct {
	MsgType         uint8          // 短消息类型
	NeedReport      uint8          // SP是否要求返回状态报告
	Priority        uint8          // 短消息发送优先级
	ServiceID       *OctetString   // 业务代码
	FeeType         *OctetString   // 收费类型
	FeeCode         *OctetString   // 资费代码
	FixedFee        *OctetString   // 包月费/封顶费
	MsgFormat       uint8          // 短消息格式
	ValidTime       *OctetString   // 短消息有效时间
	AtTime          *OctetString   // 短消息定时发送时间
	SrcTermID       *OctetString   // 短信息发送方号码
	ChargeTermID    *OctetString   // 计费用户号码
	DestTermIDCount uint8          // 短消息接收号码总数，最多 100
	DestTermID      []*OctetString // 短消息接收号码
	MsgLength       uint8          // 短消息长度
	MsgContent      []byte         // 短消息内容
	Reserve         *OctetString   // 保留

	// 可选参数
	options Options

	// used in session
	SequenceID uint32
}

func (p *SmgpSubmitReqPkt) Pack(seqId uint32) ([]byte, error) {
	var pktLen = HeaderPktLen + 114 + uint32(p.DestTermIDCount)*21 + uint32(p.MsgLength) + uint32(p.options.Len())
	var w = newPkgWriter(pktLen)
	// header
	w.WriteHeader(pktLen, seqId, SMGP_SUBMIT)
	p.SequenceID = seqId

	// body
	w.WriteByte(p.MsgType)
	w.WriteByte(p.NeedReport)
	w.WriteByte(p.Priority)
	w.WriteBytes(p.ServiceID.Byte())
	w.WriteBytes(p.FeeType.Byte())
	w.WriteBytes(p.FeeCode.Byte())
	w.WriteBytes(p.FixedFee.Byte())
	w.WriteByte(p.MsgFormat)
	w.WriteBytes(p.ValidTime.Byte())
	w.WriteBytes(p.AtTime.Byte())
	w.WriteBytes(p.SrcTermID.Byte())
	w.WriteBytes(p.ChargeTermID.Byte())
	w.WriteByte(p.DestTermIDCount)

	for _, d := range p.DestTermID {
		w.WriteBytes(d.Byte())
	}
	w.WriteByte(p.MsgLength)
	w.WriteBytes(p.MsgContent)
	w.WriteBytes(p.Reserve.Byte())

	for k, o := range p.options {
		w.WriteByte(uint8(k))
		w.WriteBytes(o)
	}

	return w.Bytes()
}

func (p *SmgpSubmitReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)
	offset := 0

	p.MsgType = r.ReadByte()
	p.NeedReport = r.ReadByte()
	p.Priority = r.ReadByte()
	p.ServiceID = &OctetString{
		Data:     r.ReadCString(10),
		FixedLen: 10,
	}
	p.FeeType = &OctetString{
		Data:     r.ReadCString(2),
		FixedLen: 2,
	}
	p.FeeCode = &OctetString{
		Data:     r.ReadCString(6),
		FixedLen: 6,
	}
	p.FixedFee = &OctetString{
		Data:     r.ReadCString(6),
		FixedLen: 6,
	}
	p.MsgFormat = r.ReadByte()
	p.ValidTime = &OctetString{
		Data:     r.ReadCString(17),
		FixedLen: 17,
	}
	p.AtTime = &OctetString{
		Data:     r.ReadCString(17),
		FixedLen: 17,
	}
	p.SrcTermID = &OctetString{
		Data:     r.ReadCString(21),
		FixedLen: 21,
	}
	p.ChargeTermID = &OctetString{
		Data:     r.ReadCString(21),
		FixedLen: 21,
	}
	p.DestTermIDCount = r.ReadByte()
	offset += 1 + 1 + 1 + 10 + 2 + 6 + 6 + 1 + 17 + 17 + 21 + 21 + 1
	for i := 0; i < int(p.DestTermIDCount); i++ {
		p.DestTermID = append(p.DestTermID, &OctetString{
			Data:     r.ReadCString(21),
			FixedLen: 21,
		})
	}
	offset += 21 * int(p.DestTermIDCount)
	p.MsgLength = r.ReadByte()
	msgContent := make([]byte, p.MsgLength)
	r.ReadBytes(msgContent)
	p.MsgContent = msgContent
	p.Reserve = &OctetString{
		Data:     r.ReadCString(8),
		FixedLen: 8,
	}
	offset += 1 + int(p.MsgLength) + 8

	options, err := ParseOptions(data[offset:])
	if err != nil {
		return err
	}
	p.options = options

	return r.Error()
}

func (p *SmgpSubmitReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Submit Req ---")
	fmt.Fprintln(&b, "MsgType: ", p.MsgType)
	fmt.Fprintln(&b, "NeedReport: ", p.NeedReport)
	fmt.Fprintln(&b, "Priority: ", p.Priority)

	fmt.Fprintln(&b, "ServiceID: ", p.ServiceID)
	fmt.Fprintln(&b, "FeeType: ", p.FeeType)
	fmt.Fprintln(&b, "FeeCode: ", p.FeeCode)
	fmt.Fprintln(&b, "FixedFee: ", p.FixedFee)

	fmt.Fprintln(&b, "MsgFormat: ", p.MsgFormat)
	fmt.Fprintln(&b, "ValidTime: ", p.ValidTime)
	fmt.Fprintln(&b, "AtTime: ", p.AtTime)
	fmt.Fprintln(&b, "SrcTermID: ", p.SrcTermID)
	fmt.Fprintln(&b, "ChargeTermID: ", p.ChargeTermID)

	fmt.Fprintln(&b, "DestTermIDCount: ", p.DestTermIDCount)
	for i := 0; i < int(p.DestTermIDCount); i++ {
		fmt.Fprintln(&b, "DestTermID: ", p.DestTermID[i])
	}

	fmt.Fprintln(&b, "MsgLength: ", p.MsgLength)
	fmt.Fprintln(&b, "MsgContent: ", string(p.MsgContent))

	return b.String()
}

type SmgpSubmitRespPkt struct {
	MsgID  string
	Status Status

	// used in session
	SequenceID uint32
}

func (p *SmgpSubmitRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpSubmitRespPktLen)
	// header
	w.WriteHeader(SmgpSubmitRespPktLen, seqId, SMGP_SUBMIT_RESP)
	p.SequenceID = seqId

	// body
	w.WriteFixedSizeString(p.MsgID, 10)
	w.WriteInt(binary.BigEndian, p.Status)

	return w.Bytes()
}

func (p *SmgpSubmitRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	// Body: MsgID
	var s = make([]byte, 10)
	r.ReadBytes(s)
	p.MsgID = hex.EncodeToString(s)
	// Body: Status
	r.ReadInt(binary.BigEndian, &p.Status)

	return r.Error()
}

func (p *SmgpSubmitRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Submit Resp ---")
	fmt.Fprintln(&b, "MsgID: ", p.MsgID)
	fmt.Fprintln(&b, "Status: ", p.Status)
	return b.String()
}
