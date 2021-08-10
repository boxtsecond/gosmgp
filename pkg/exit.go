package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmgpExitReqPktLen  uint32 = 12 //12d, 0xc
	SmgpExitRespPktLen uint32 = 12 //12d, 0xc
)

type SmgpExitReqPkt struct {
	// used in session
	SequenceID uint32
}
type SmgpExitRespPkt struct {
	// used in session
	SequenceID uint32
}

func (p *SmgpExitReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpExitReqPktLen)

	// header
	w.WriteHeader(SmgpExitReqPktLen, seqId, SMGP_ACTIVE_TEST)
	p.SequenceID = seqId

	return w.Bytes()
}

func (p *SmgpExitReqPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmgpExitReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Exit Req ---")
	return b.String()
}

func (p *SmgpExitRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpExitRespPktLen)

	// header
	w.WriteHeader(SmgpExitRespPktLen, seqId, SMGP_ACTIVE_TEST_RESP)
	p.SequenceID = seqId

	return w.Bytes()
}

func (p *SmgpExitRespPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmgpExitRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Exit Resp ---")
	return b.String()
}
