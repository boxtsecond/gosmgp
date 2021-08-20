package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	SmgpQueryReqPktLen  = HeaderPktLen + 8 + 1 + 10           //42d, 0x2a
	SmgpQueryRespPktLen = HeaderPktLen + 8 + 1 + 10 + 4*8 + 8 //33d, 0x21
)

type SmgpQueryReqPkt struct {
	QueryTime string
	QueryType uint8
	QueryCode string

	// used in session
	SequenceID uint32
}

func (p *SmgpQueryReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpQueryReqPktLen)
	// header
	w.WriteHeader(SmgpQueryReqPktLen, seqId, SMGP_QUERY)
	p.SequenceID = seqId

	// body
	w.WriteFixedSizeString(p.QueryTime, 8)
	w.WriteByte(p.QueryType)
	w.WriteFixedSizeString(p.QueryCode, 10)

	return w.Bytes()
}

func (p *SmgpQueryReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	p.QueryTime = string(r.ReadCString(8))
	p.QueryType = r.ReadByte()
	p.QueryCode = string(r.ReadCString(10))

	return r.Error()
}

func (p *SmgpQueryReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Query Req ---")
	fmt.Fprintln(&b, "QueryTime: ", p.QueryTime)
	fmt.Fprintln(&b, "QueryType: ", p.QueryType)
	fmt.Fprintln(&b, "QueryCode: ", p.QueryCode)
	return b.String()
}

type SmgpQueryRespPkt struct {
	QueryTime string
	QueryType uint8
	QueryCode string
	MT_TLMsg  uint32
	MT_Tlusr  uint32
	MT_Scs    uint32
	MT_WT     uint32
	MT_FL     uint32
	MO_Scs    uint32
	MO_WT     uint32
	MO_FL     uint32
	Reserve   string

	// used in session
	SequenceID uint32
}

func (p *SmgpQueryRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpQueryRespPktLen)
	// header
	w.WriteHeader(SmgpQueryRespPktLen, seqId, SMGP_QUERY_RESP)
	p.SequenceID = seqId

	// body
	w.WriteFixedSizeString(p.QueryTime, 8)
	w.WriteByte(p.QueryType)
	w.WriteFixedSizeString(p.QueryCode, 10)
	w.WriteInt(binary.BigEndian, p.MT_TLMsg)
	w.WriteInt(binary.BigEndian, p.MT_Tlusr)
	w.WriteInt(binary.BigEndian, p.MT_Scs)
	w.WriteInt(binary.BigEndian, p.MT_WT)
	w.WriteInt(binary.BigEndian, p.MT_FL)
	w.WriteInt(binary.BigEndian, p.MO_Scs)
	w.WriteInt(binary.BigEndian, p.MO_WT)
	w.WriteInt(binary.BigEndian, p.MO_FL)
	w.WriteFixedSizeString(p.Reserve, 8)

	return w.Bytes()
}

func (p *SmgpQueryRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	p.QueryTime = string(r.ReadCString(8))
	p.QueryType = r.ReadByte()
	p.QueryCode = string(r.ReadCString(10))
	r.ReadInt(binary.BigEndian, &p.MT_TLMsg)
	r.ReadInt(binary.BigEndian, &p.MT_Tlusr)
	r.ReadInt(binary.BigEndian, &p.MT_Scs)
	r.ReadInt(binary.BigEndian, &p.MT_WT)
	r.ReadInt(binary.BigEndian, &p.MT_FL)
	r.ReadInt(binary.BigEndian, &p.MO_Scs)
	r.ReadInt(binary.BigEndian, &p.MO_WT)
	r.ReadInt(binary.BigEndian, &p.MO_FL)
	p.Reserve = string(r.ReadCString(8))

	return r.Error()
}

func (p *SmgpQueryRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Query Resp ---")
	fmt.Fprintln(&b, "QueryTime: ", p.QueryTime)
	fmt.Fprintln(&b, "QueryType: ", p.QueryType)
	fmt.Fprintln(&b, "QueryCode: ", p.QueryCode)
	fmt.Fprintln(&b, "MT_TLMsg: ", p.MT_TLMsg)
	fmt.Fprintln(&b, "MT_Tlusr: ", p.MT_Tlusr)
	fmt.Fprintln(&b, "MT_Scs: ", p.MT_Scs)
	fmt.Fprintln(&b, "MT_WT: ", p.MT_WT)
	fmt.Fprintln(&b, "MT_FL: ", p.MT_FL)
	fmt.Fprintln(&b, "MO_Scs: ", p.MO_Scs)
	fmt.Fprintln(&b, "MO_WT: ", p.MO_WT)
	fmt.Fprintln(&b, "MO_FL: ", p.MO_FL)
	fmt.Fprintln(&b, "Reserve: ", p.Reserve)
	return b.String()
}
