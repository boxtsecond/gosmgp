package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmgpActiveTestReqPktLen  uint32 = 12 //12d, 0xc
	SmgpActiveTestRespPktLen uint32 = 12 //12d, 0xc
)

type SmgpActiveTestReqPkt struct {
	// used in session
	SequenceID uint32
}
type SmgpActiveTestRespPkt struct {
	// used in session
	SequenceID uint32
}

func (p *SmgpActiveTestReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpActiveTestReqPktLen)

	// header
	w.WriteHeader(SmgpActiveTestReqPktLen, seqId, SMGP_ACTIVE_TEST)
	p.SequenceID = seqId

	return w.Bytes()
}

func (p *SmgpActiveTestReqPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmgpActiveTestReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP ActiveTest Req ---")
	return b.String()
}

func (p *SmgpActiveTestRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpActiveTestRespPktLen)

	// header
	w.WriteHeader(SmgpActiveTestRespPktLen, seqId, SMGP_ACTIVE_TEST_RESP)
	p.SequenceID = seqId

	return w.Bytes()
}

func (p *SmgpActiveTestRespPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmgpActiveTestRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP ActiveTest Resp ---")
	return b.String()
}
