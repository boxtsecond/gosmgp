package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/boxtsecond/gosmgp/pkg"
	"github.com/boxtsecond/gosmgp/server"
)

const (
	user     string = "10000001"
	password string = "12345678"
	spId     string = "123456"
)

func handleLogin(r *server.Response, p *server.Packet, l *log.Logger) (bool, error) {
	req, ok := p.Packer.(*pkg.SmgpLoginReqPkt)
	if !ok {
		return true, nil
	}

	l.Println("remote addr:", p.Conn.Conn.RemoteAddr().(*net.TCPAddr).IP.String())
	resp := r.Packer.(*pkg.SmgpLoginRespPkt)

	resp.ServerVersion = pkg.VERSION
	if req.ClientID != user {
		resp.Status = pkg.Status(21)
		l.Println("handleLogin error:", resp.Status.Error())
		return false, resp.Status.Error()
	}

	tm := req.TimeStamp
	auth, err := pkg.GenAuthenticatorClient(req.ClientID, password, tm)

	if err != nil || req.AuthenticatorClient != string(auth[:]) {
		resp.Status = pkg.Status(21)
		l.Println("handleLogin error:", resp.Status.Error())
		return false, resp.Status.Error()
	}

	resp.AuthenticatorServer = string(auth[:])
	l.Printf("handleLogin: %s login ok\n", req.ClientID)

	return false, nil
}

func handleSubmit(r *server.Response, p *server.Packet, l *log.Logger) (bool, error) {
	req, ok := p.Packer.(*pkg.SmgpSubmitReqPkt)
	if !ok {
		return true, nil
	}

	resp := r.Packer.(*pkg.SmgpSubmitRespPkt)
	resp.MsgID, _ = pkg.GenMsgID(spId, <-p.Conn.SequenceNum)
	deliverPkgs := make([]*pkg.SmgpDeliverReqPkt, 0)
	for i, d := range req.DestTermID {
		l.Printf("handleSubmit: handle submit from %s ok! msgid[%s], destTerminalId[%s]\n",
			req.SrcTermID, fmt.Sprintf("%s_%d", resp.MsgID, i), d)
		t := pkg.GenNowTimeYYStr()
		msgStat := pkg.SmgpDeliverMsgContent{
			ID:         resp.MsgID,
			Sub:        "",
			Dlvrd:      "",
			SubmitDate: t,
			DoneDate:   t,
			Stat:       "DELIVRD",
			Err:        "",
			Txt:        "",
		}
		msgContent, _ := msgStat.Encode()
		deliverPkgs = append(deliverPkgs, &pkg.SmgpDeliverReqPkt{
			MsgID:      resp.MsgID,
			IsReport:   pkg.IS_REPORT,
			MsgFormat:  pkg.ASCII,
			RecvTime:   pkg.GenNowTimeYYYYStr(),
			SrcTermID:  req.SrcTermID,
			DestTermID: d,
			MsgLength:  uint8(len(msgContent)),
			MsgContent: msgContent,
			Reserve:    "",
			Options:    nil,
			SequenceID: <-p.Conn.SequenceID,
		})
	}
	go mockDeliver(deliverPkgs, p)
	return true, nil
}

func mockDeliver(pkgs []*pkg.SmgpDeliverReqPkt, s *server.Packet) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:

			for _, p := range pkgs {
				err := s.SendPkt(p, p.SequenceID)
				if err != nil {
					log.Printf("server smgp: send a smgp deliver request error: %s.", err)
					return
				} else {
					log.Printf("server smgp: send a smgp deliver request ok.")
				}
			}

		default:
		}

	}
}

func main() {
	var handlers = []server.Handler{
		server.HandlerFunc(handleLogin),
		server.HandlerFunc(handleSubmit),
	}

	err := server.ListenAndServe(":8890",
		pkg.VERSION,
		5*time.Second,
		3,
		nil,
		handlers...,
	)
	if err != nil {
		log.Println("smgp Listen And Server error:", err)
	}
	return
}
