package pkg

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	cmd "github.com/boxtsecond/gosmgp/pkg/request"
)

type State uint8

const (
	CONNECTION_CLOSED State = iota
	CONNECTION_CONNECTED
	CONNECTION_AUTHOK
)

type Conn struct {
	net.Conn
	State   State
	Version uint8

	// for SeqId generator goroutine
	SeqId <-chan uint32
	done  chan<- struct{}
}

func newSeqIdGenerator() (<-chan uint32, chan<- struct{}) {
	out := make(chan uint32)
	done := make(chan struct{})

	go func() {
		var i uint32
		for {
			select {
			case out <- i:
				i++
			case <-done:
				close(out)
				return
			}
		}
	}()
	return out, done
}

func NewConnection(conn net.Conn, v uint8) *Conn {
	seqId, done := newSeqIdGenerator()
	c := &Conn{
		Conn:    conn,
		Version: v,
		SeqId:   seqId,
		done:    done,
	}
	tc := c.Conn.(*net.TCPConn)
	tc.SetKeepAlive(true) //Keepalive as default
	return c
}

func (c *Conn) Close() {
	if c != nil {
		if c.State == CONNECTION_CLOSED {
			return
		}
		close(c.done)  // let the SeqId goroutine exit.
		c.Conn.Close() // close the underlying net.Conn
		c.State = CONNECTION_CLOSED
	}
}

func (c *Conn) SetState(state State) {
	c.State = state
}

func (c *Conn) SendPkt(packet cmd.Packer, seqId uint32) error {
	if c.State == CONNECTION_CLOSED {
		return cmd.ErrConnIsClosed
	}

	data, err := packet.Pack(seqId)
	if err != nil {
		return err
	}

	_, err = c.Conn.Write(data) //block write
	if err != nil {
		return err
	}

	return nil
}

const (
	defaultReadBufferSize = 4096
)

type readBuffer struct {
	Header   cmd.Header
	leftData [defaultReadBufferSize]byte
}

var readBufferPool = sync.Pool{
	New: func() interface{} {
		return &readBuffer{}
	},
}

func (c *Conn) RecvAndUnpackPkt(timeout time.Duration) (interface{}, error) {
	if c.State == CONNECTION_CLOSED {
		return nil, cmd.ErrConnIsClosed
	}
	defer c.SetReadDeadline(time.Time{})

	rb := readBufferPool.Get().(*readBuffer)
	defer readBufferPool.Put(rb)

	if timeout != 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}

	// packet header
	err := binary.Read(c.Conn, binary.BigEndian, &rb.Header)
	if err != nil {
		netErr, ok := err.(net.Error)
		if ok {
			if netErr.Timeout() {
				return nil, cmd.ErrReadHeaderTimeout
			}
		}
		return nil, err
	}

	if rb.Header.PacketLength < cmd.SMGP_PACKET_MIN || rb.Header.PacketLength > cmd.SMGP_PACKET_MAX {
		return nil, cmd.ErrTotalLengthInvalid
	}

	if !((rb.Header.RequestID > cmd.SMGP_REQUEST_MIN && rb.Header.RequestID < cmd.SMGP_REQUEST_MAX) ||
		(rb.Header.RequestID > cmd.SMGP_RESPONSE_MIN && rb.Header.RequestID < cmd.SMGP_RESPONSE_MAX)) {
		return nil, cmd.ErrRequestIDInvalid
	}

	if timeout != 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}

	// packet body
	var leftData = rb.leftData[0:(rb.Header.PacketLength - 12)]
	_, err = io.ReadFull(c.Conn, leftData)
	if err != nil {
		netErr, ok := err.(net.Error)
		if ok {
			if netErr.Timeout() {
				return nil, cmd.ErrReadPktBodyTimeout
			}
		}
		return nil, err
	}

	var p cmd.Packer

	switch rb.Header.RequestID {
	case cmd.SMGP_LOGIN:
		p = &cmd.SmgpLoginReqPkt{
			SequenceID: rb.Header.SequenceID,
		}
	//case CMPP_CONNECTIONECT_RESP:
	//	if c.Typ == V30 {
	//		p = &Cmpp3ConnRspPkt{}
	//	} else {
	//		p = &Cmpp2ConnRspPkt{}
	//	}
	//case CMPP_TERMINATE:
	//	p = &CmppTerminateReqPkt{}
	//case CMPP_TERMINATE_RESP:
	//	p = &CmppTerminateRspPkt{}
	//case CMPP_SUBMIT:
	//	if c.Typ == V30 {
	//		p = &Cmpp3SubmitReqPkt{}
	//	} else {
	//		p = &Cmpp2SubmitReqPkt{}
	//	}
	//case CMPP_SUBMIT_RESP:
	//	if c.Typ == V30 {
	//		p = &Cmpp3SubmitRspPkt{}
	//	} else {
	//		p = &Cmpp2SubmitRspPkt{}
	//	}
	//case CMPP_DELIVER:
	//	if c.Typ == V30 {
	//		p = &Cmpp3DeliverReqPkt{}
	//	} else {
	//		p = &Cmpp2DeliverReqPkt{}
	//	}
	//case CMPP_DELIVER_RESP:
	//	if c.Typ == V30 {
	//		p = &Cmpp3DeliverRspPkt{}
	//	} else {
	//		p = &Cmpp2DeliverRspPkt{}
	//	}
	//case CMPP_FWD:
	//	if c.Typ == V30 {
	//		p = &Cmpp3FwdReqPkt{}
	//	} else {
	//		p = &Cmpp2FwdReqPkt{}
	//	}
	//case CMPP_FWD_RESP:
	//	if c.Typ == V30 {
	//		p = &Cmpp3FwdRspPkt{}
	//	} else {
	//		p = &Cmpp2FwdRspPkt{}
	//	}
	//case CMPP_ACTIVE_TEST:
	//	p = &CmppActiveTestReqPkt{}
	//case CMPP_ACTIVE_TEST_RESP:
	//	p = &CmppActiveTestRspPkt{}

	default:
		return nil, cmd.ErrRequestIDNotSupported
	}

	err = p.Unpack(leftData)
	if err != nil {
		return nil, err
	}
	return p, nil
}
