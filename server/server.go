package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/boxtsecond/gosmgp/pkg"
)

var (
	ErrEmptyServerAddr    = errors.New("smgp server listen: empty server addr")
	ErrNoHandlers         = errors.New("smgp server: no connection handler")
	ErrUnsupportedPkt     = errors.New("smgp server read packet: receive a unsupported pkt")
	ErrUnsupportedVersion = errors.New("smgp server read packet: receive a unsupported version")
)

type Packet struct {
	pkg.Packer
	*pkg.Conn
}

type Response struct {
	*Packet
	pkg.Packer
	SequenceID uint32
}

type Handler interface {
	ServeSmgp(*Response, *Packet, *log.Logger) (bool, error)
}

type HandlerFunc func(*Response, *Packet, *log.Logger) (bool, error)

func (f HandlerFunc) ServeSmgp(r *Response, p *Packet, l *log.Logger) (bool, error) {
	return f(r, p, l)
}

type Server struct {
	Addr    string
	Handler Handler

	// protocol info
	Version uint8
	T       time.Duration
	N       int32

	ErrorLog *log.Logger
}

type conn struct {
	*pkg.Conn
	server *Server

	// for active test
	t       time.Duration // interval between two active tests
	n       int32         // continuous send times when no response back
	done    chan struct{}
	exceed  chan struct{}
	counter int32
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	var tempDelay time.Duration
	for {
		rw, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				srv.ErrorLog.Printf("accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c, err := srv.newConn(rw)
		if err != nil {
			continue
		}

		srv.ErrorLog.Printf("accept a connection from %v\n", c.Conn.RemoteAddr())
		go c.serve()
	}
}

func (c *conn) readPacket() (*Response, error) {
	readTimeout := time.Second * 2
	i, err := c.Conn.RecvAndUnpackPkt(readTimeout)
	if err != nil {
		return nil, err
	}
	ver := c.server.Version

	var rsp *Response
	switch p := i.(type) {
	case *pkg.SmgpLoginReqPkt:
		if p.ClientVersion != ver {
			return nil, pkg.NewOpError(ErrUnsupportedVersion,
				fmt.Sprintf("readPacket: receive unsupported version: %#v", p))
		}
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmgpLoginRespPkt{
				SequenceID: p.SequenceID,
			},
			SequenceID: p.SequenceID,
		}
		c.server.ErrorLog.Printf("receive a smgp login request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)

	case *pkg.SmgpSubmitReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmgpSubmitRespPkt{
				SequenceID: p.SequenceID,
			},
			SequenceID: p.SequenceID,
		}
		c.server.ErrorLog.Printf("receive a smgp submit request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)

	case *pkg.SmgpDeliverReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmgpDeliverRespPkt{
				SequenceID: p.SequenceID,
			},
			SequenceID: p.SequenceID,
		}
		c.server.ErrorLog.Printf("receive a smgp deliver response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)

	case *pkg.SmgpDeliverRespPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
		}
		c.server.ErrorLog.Printf("receive a smgp deliver response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)

	case *pkg.SmgpActiveTestReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmgpActiveTestRespPkt{
				SequenceID: p.SequenceID,
			},
			SequenceID: p.SequenceID,
		}
		c.server.ErrorLog.Printf("receive a smgp active request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)

	case *pkg.SmgpActiveTestRespPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
		}
		c.server.ErrorLog.Printf("receive a smgp active response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)

	case *pkg.SmgpExitReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmgpExitRespPkt{
				SequenceID: p.SequenceID,
			},
			SequenceID: p.SequenceID,
		}
		c.server.ErrorLog.Printf("receive a smgp exit request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)

	case *pkg.SmgpExitRespPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
		}
		c.server.ErrorLog.Printf("receive a smgp exit response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceID)
	default:
		return nil, pkg.NewOpError(ErrUnsupportedPkt,
			fmt.Sprintf("readPacket: receive unsupported packet type: %#v", p))
	}
	return rsp, nil
}

func (c *conn) close() {
	p := &pkg.SmgpExitReqPkt{}

	err := c.Conn.SendPkt(p, <-c.Conn.SequenceID)
	if err != nil {
		c.server.ErrorLog.Printf("send smgp exit request packet to %v error: %v\n", c.Conn.RemoteAddr(), err)
	}

	close(c.done)
	c.server.ErrorLog.Printf("close connection with %v!\n", c.Conn.RemoteAddr())
	c.Conn.Close()
}

func (c *conn) finishPacket(r *Response) error {
	if _, ok := r.Packet.Packer.(*pkg.SmgpActiveTestRespPkt); ok {
		atomic.AddInt32(&c.counter, -1)
		return nil
	}

	if r.Packer == nil {
		return nil
	}

	return c.Conn.SendPkt(r.Packer, r.SequenceID)
}

func startActiveTest(c *conn) {
	exceed, done := make(chan struct{}), make(chan struct{})
	c.done = done
	c.exceed = exceed

	go func() {
		t := time.NewTicker(c.t)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if atomic.LoadInt32(&c.counter) >= c.n {
					c.server.ErrorLog.Printf("no smgp active test response returned from %v for %d times!",
						c.Conn.RemoteAddr(), c.n)
					exceed <- struct{}{}
					break
				}
				p := &pkg.SmgpActiveTestReqPkt{}
				err := c.Conn.SendPkt(p, <-c.Conn.SequenceID)
				if err != nil {
					c.server.ErrorLog.Printf("send smgp active test request to %v error: %v", c.Conn.RemoteAddr(), err)
				} else {
					atomic.AddInt32(&c.counter, 1)
				}
			}
		}
	}()
}

func (c *conn) serve() {
	defer func() {
		if err := recover(); err != nil {
			c.server.ErrorLog.Printf("panic serving %v: %v\n", c.Conn.RemoteAddr(), err)
		}
	}()

	defer c.close()

	startActiveTest(c)

	for {
		select {
		case <-c.exceed:
			return
		default:
		}

		r, err := c.readPacket()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				continue
			}
			break
		}

		_, err = c.server.Handler.ServeSmgp(r, r.Packet, c.server.ErrorLog)
		if err1 := c.finishPacket(r); err1 != nil {
			break
		}

		if err != nil {
			break
		}
	}
}

func (srv *Server) newConn(rwc net.Conn) (c *conn, err error) {
	c = new(conn)
	c.server = srv
	c.Conn = pkg.NewConnection(rwc, srv.Version)
	c.Conn.SetState(pkg.CONNECTION_CONNECTED)
	c.n = c.server.N
	c.t = c.server.T
	return c, nil
}

func (srv *Server) listenAndServe() error {
	if srv.Addr == "" {
		return ErrEmptyServerAddr
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

func ListenAndServe(addr string, version uint8, t time.Duration, n int32, logWriter io.Writer, handlers ...Handler) error {
	if addr == "" {
		return ErrEmptyServerAddr
	}

	if handlers == nil {
		return ErrNoHandlers
	}

	var handler Handler
	handler = HandlerFunc(func(r *Response, p *Packet, l *log.Logger) (bool, error) {
		for _, h := range handlers {
			next, err := h.ServeSmgp(r, p, l)
			if err != nil || !next {
				return next, err
			}
		}
		return false, nil
	})

	if logWriter == nil {
		logWriter = os.Stderr
	}
	server := &Server{Addr: addr, Handler: handler, Version: version,
		T: t, N: n,
		ErrorLog: log.New(logWriter, "smgp server: ", log.LstdFlags)}
	return server.listenAndServe()
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(1 * time.Minute) // 1min
	return tc, nil
}
