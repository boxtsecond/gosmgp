package pkg

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
	SequenceID <-chan uint32
	done       chan<- struct{}
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
		Conn:       conn,
		Version:    v,
		SequenceID: seqId,
		done:       done,
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

func (c *Conn) SendPkt(packet Packer, seqId uint32) error {
	if c.State == CONNECTION_CLOSED {
		return ErrConnIsClosed
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
	Header
	leftData [defaultReadBufferSize]byte
}

var readBufferPool = sync.Pool{
	New: func() interface{} {
		return &readBuffer{}
	},
}

func (c *Conn) RecvAndUnpackPkt(timeout time.Duration) (Packer, error) {
	if c.State == CONNECTION_CLOSED {
		return nil, ErrConnIsClosed
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
				return nil, ErrReadHeaderTimeout
			}
		}
		return nil, err
	}

	if rb.Header.PacketLength < SMGP_PACKET_MIN || rb.Header.PacketLength > SMGP_PACKET_MAX {
		return nil, ErrTotalLengthInvalid
	}

	if !((RequestID(rb.Header.RequestID) > SMGP_REQUEST_MIN && RequestID(rb.Header.RequestID) < SMGP_REQUEST_MAX) ||
		(RequestID(rb.Header.RequestID) > SMGP_RESPONSE_MIN && RequestID(rb.Header.RequestID) < SMGP_RESPONSE_MAX)) {
		return nil, ErrRequestIDInvalid
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
				return nil, ErrReadPktBodyTimeout
			}
		}
		return nil, err
	}

	var p Packer
	sequenceID := rb.Header.SequenceID
	fmt.Println("===============")
	fmt.Println(RequestID(rb.Header.RequestID))

	switch RequestID(rb.Header.RequestID) {
	case SMGP_ACTIVE_TEST:
		p = &SmgpActiveTestReqPkt{SequenceID: sequenceID}
	case SMGP_ACTIVE_TEST_RESP:
		p = &SmgpActiveTestRespPkt{SequenceID: sequenceID}
	case SMGP_LOGIN:
		p = &SmgpLoginReqPkt{SequenceID: sequenceID}
	case SMGP_LOGIN_RESP:
		p = &SmgpLoginRespPkt{SequenceID: sequenceID}
	case SMGP_SUBMIT:
		p = &SmgpSubmitReqPkt{SequenceID: sequenceID}
	case SMGP_SUBMIT_RESP:
		p = &SmgpSubmitRespPkt{SequenceID: sequenceID}
	case SMGP_DELIVER:
		p = &SmgpDeliverReqPkt{SequenceID: sequenceID}
	case SMGP_DELIVER_RESP:
		p = &SmgpDeliverRespPkt{SequenceID: sequenceID}
	case SMGP_EXIT:
		p = &SmgpExitReqPkt{SequenceID: sequenceID}
	case SMGP_EXIT_RESP:
		p = &SmgpExitRespPkt{SequenceID: sequenceID}

	default:
		return nil, ErrRequestIDNotSupported
	}

	err = p.Unpack(leftData)
	if err != nil {
		return nil, err
	}
	return p, nil
}
