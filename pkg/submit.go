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
	MsgType         uint8    // 短消息类型
	NeedReport      uint8    // SP是否要求返回状态报告
	Priority        uint8    // 短消息发送优先级
	ServiceID       string   // 业务代码
	FeeType         string   // 收费类型
	FeeCode         string   // 资费代码
	FixedFee        string   // 包月费/封顶费
	MsgFormat       uint8    // 短消息格式
	ValidTime       string   // 短消息有效时间
	AtTime          string   // 短消息定时发送时间
	SrcTermID       string   // 短信息发送方号码
	ChargeTermID    string   // 计费用户号码
	DestTermIDCount uint8    // 短消息接收号码总数，最多 100
	DestTermID      []string // 短消息接收号码
	MsgLength       uint8    // 短消息长度
	MsgContent      string   // 短消息内容
	Reserve         string   // 保留

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
	w.WriteFixedSizeString(p.ServiceID, 10)
	w.WriteFixedSizeString(p.FeeType, 2)
	w.WriteFixedSizeString(p.FeeCode, 6)
	w.WriteFixedSizeString(p.FixedFee, 6)
	w.WriteByte(p.MsgFormat)
	w.WriteFixedSizeString(p.ValidTime, 17)
	w.WriteFixedSizeString(p.AtTime, 17)
	w.WriteFixedSizeString(p.SrcTermID, 21)
	w.WriteFixedSizeString(p.ChargeTermID, 21)
	w.WriteByte(p.DestTermIDCount)

	for _, d := range p.DestTermID {
		w.WriteFixedSizeString(d, 21)
	}
	w.WriteByte(p.MsgLength)
	w.WriteString(p.MsgContent)
	w.WriteFixedSizeString(p.Reserve, 8)

	for _, o := range p.options {
		//w.WriteByte(uint8(k))
		w.WriteBytes(o.Byte())
	}

	return w.Bytes()
}

func (p *SmgpSubmitReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)
	offset := 0

	p.MsgType = r.ReadByte()
	p.NeedReport = r.ReadByte()
	p.Priority = r.ReadByte()
	p.ServiceID = string(r.ReadCString(10))
	p.FeeType = string(r.ReadCString(2))
	p.FeeCode = string(r.ReadCString(6))
	p.FixedFee = string(r.ReadCString(6))
	p.MsgFormat = r.ReadByte()
	p.ValidTime = string(r.ReadCString(17))
	p.AtTime = string(r.ReadCString(17))
	p.SrcTermID = string(r.ReadCString(21))
	p.ChargeTermID = string(r.ReadCString(21))
	p.DestTermIDCount = r.ReadByte()
	offset += 1 + 1 + 1 + 10 + 2 + 6 + 6 + 1 + 17 + 17 + 21 + 21 + 1
	for i := 0; i < int(p.DestTermIDCount); i++ {
		p.DestTermID = append(p.DestTermID, string(r.ReadCString(21)))
	}
	offset += 21 * int(p.DestTermIDCount)
	p.MsgLength = r.ReadByte()
	msgContent := make([]byte, p.MsgLength)
	r.ReadBytes(msgContent)
	p.MsgContent = string(msgContent)
	p.Reserve = string(r.ReadCString(8))
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
