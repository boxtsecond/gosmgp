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

type SmgpDeliverMsgContent struct {
	SubmitMsgID string // submit resp 的 MsgID
	Sub         string
	Dlvrd       string
	SubmitDate  string
	DoneDate    string
	Stat        string
	Err         string
	Txt         string
}

func (p *SmgpDeliverMsgContent) Encode() string {
	//id:!rT sub:001 dlrvd:001 submit_date:2108131621 done_date:2108131621 stat:GB:0005 err:000 txt:00000000000000000000
	id, _ := hex.DecodeString(p.SubmitMsgID)
	msgStatStr := fmt.Sprintf("id:%s sub:%s dlvrd:%s submit_date:%s done_date:%s stat:%s err:%s Text:%s", id, p.Sub, p.Dlvrd, p.SubmitDate, p.DoneDate, p.Stat, p.Err, p.Txt)

	return msgStatStr
}

func DecodeDeliverMsgContent(data []byte) *SmgpDeliverMsgContent {
	p := &SmgpDeliverMsgContent{}
	p.SubmitMsgID = hex.EncodeToString(data[3:13])
	p.Sub = string(data[18:21])
	p.Dlvrd = string(data[28:31])
	p.SubmitDate = string(data[44:54])
	p.DoneDate = string(data[65:75])
	p.Stat = string(data[81:88])
	p.Err = string(data[93:96])
	p.Txt = string(data[101:])
	return p
}

func (p *SmgpDeliverMsgContent) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "\tID(SubmitMsgID): ", p.SubmitMsgID)
	fmt.Fprintln(&b, "\tSub: ", p.Sub)
	fmt.Fprintln(&b, "\tDlvrd: ", p.Dlvrd)
	fmt.Fprintln(&b, "\tSubmitDate: ", p.SubmitDate)
	fmt.Fprintln(&b, "\tDoneDate: ", p.DoneDate)
	fmt.Fprintln(&b, "\tStat: ", p.Stat)
	fmt.Fprintln(&b, "\tErr: ", p.Err)
	fmt.Fprintln(&b, "\tTxt: ", p.Txt)

	return b.String()
}

type SmgpDeliverReqPkt struct {
	MsgID      string // 短消息流水号
	IsReport   uint8  // 短消息流水号
	MsgFormat  uint8  // 短消息格式
	RecvTime   string // 短消息接收时间
	SrcTermID  string // 短消息发送号码
	DestTermID string // 短消息接收号码
	MsgLength  uint8  //  短消息长度
	MsgContent []byte // 短消息内容
	Reserve    string // 保留

	// 可选字段
	Options Options

	// used in session
	SequenceID     uint32
	MsgStatContent *SmgpDeliverMsgContent
}

func (p *SmgpDeliverReqPkt) Pack(seqId uint32) ([]byte, error) {
	var pktLen = HeaderPktLen + 77 + uint32(p.MsgLength) + uint32(p.Options.Len())
	var w = newPkgWriter(pktLen)
	// header
	w.WriteHeader(pktLen, seqId, SMGP_DELIVER)
	p.SequenceID = seqId

	// body
	msgId, _ := hex.DecodeString(p.MsgID)
	w.WriteBytes(NewOctetString(fmt.Sprintf("%s", msgId)).Byte(10))
	w.WriteByte(p.IsReport)
	w.WriteByte(p.MsgFormat)
	w.WriteFixedSizeString(p.RecvTime, 14)
	w.WriteFixedSizeString(p.SrcTermID, 21)
	w.WriteFixedSizeString(p.DestTermID, 21)
	w.WriteByte(p.MsgLength)
	w.WriteBytes(p.MsgContent)
	w.WriteFixedSizeString(p.Reserve, 8)

	for _, o := range p.Options {
		b, _ := o.Byte()
		w.WriteBytes(b)
	}

	return w.Bytes()
}

func (p *SmgpDeliverReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)
	offset := 0

	p.MsgID = hex.EncodeToString(r.ReadCString(10))
	p.IsReport = r.ReadByte()
	p.MsgFormat = r.ReadByte()
	p.RecvTime = string(r.ReadCString(14))
	p.SrcTermID = string(r.ReadCString(21))
	p.DestTermID = string(r.ReadCString(21))
	p.MsgLength = r.ReadByte()

	s := make([]byte, p.MsgLength)
	r.ReadBytes(s)
	p.MsgContent = s
	p.MsgStatContent = DecodeDeliverMsgContent(p.MsgContent)

	p.Reserve = string(r.ReadCString(8))
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
	fmt.Fprintln(&b, "MsgContent: ", p.MsgStatContent.String())
	fmt.Fprintln(&b, "Options: ", p.Options.String())

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
	msgId, _ := hex.DecodeString(p.MsgID)
	w.WriteBytes(NewOctetString(fmt.Sprintf("%s", msgId)).Byte(10))
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
