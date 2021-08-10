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
	fmt.Println("-------handle submit-------")
	fmt.Println(req)

	resp := r.Packer.(*pkg.SmgpSubmitRespPkt)
	resp.MsgID = "12878564852733378560" //0xb2, 0xb9, 0xda, 0x80, 0x00, 0x01, 0x00, 0x00
	for i, d := range req.DestTermID {
		l.Printf("handleSubmit: handle submit from %s ok! msgid[%d], destTerminalId[%s]\n",
			req.SrcTermID, fmt.Sprintf("%s_%d", resp.MsgID, i), d)
	}
	return true, nil
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
