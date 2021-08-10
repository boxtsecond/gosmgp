package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boxtsecond/gosmgp/client"
	"github.com/boxtsecond/gosmgp/pkg"
)

var (
	addr      = flag.String("addr", ":8890", "smgp addr(运营商地址)")
	clientID  = flag.String("clientID", "10000001", "登陆账号")
	secret    = flag.String("secret", "12345678", "登陆密码")
	loginMode = flag.String("loginMode", "2", "登陆密码")
	spID      = flag.String("spID", "", "企业代码")
	spCode    = flag.String("spCode", "", "SP的接入号码")
	phone     = flag.String("phone", "8618012345678", "接收手机号码, 86..., 多个使用,分割")
	msg       = flag.String("msg", "验证码：1234", "短信内容")
)

func startAClient(idx int) {
	c := client.NewClient(pkg.VERSION)
	defer wg.Done()
	defer c.Disconnect()

	mode, _ := strconv.Atoi(*loginMode)
	err := c.Connect(*addr, *clientID, *secret, uint8(mode), 3*time.Second)
	if err != nil {
		log.Printf("client %d: connect error: %s.", idx, err)
		return
	}
	log.Printf("client %d: connect and auth ok", idx)

	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			cont, err := pkg.Utf8ToUcs2(*msg)
			if err != nil {
				fmt.Printf("client %d: utf8 to ucs2 transform err: %s.", idx, err)
				return
			}
			destStrArr := strings.Split(*phone, ",")

			p := &pkg.SmgpSubmitReqPkt{
				MsgType:         pkg.MT,
				NeedReport:      pkg.NEED_REPORT,
				Priority:        1,
				ServiceID:       "",
				FeeType:         "00",
				FeeCode:         "0",
				FixedFee:        "0",
				MsgFormat:       8,
				ValidTime:       "",
				AtTime:          "",
				SrcTermID:       *spCode,
				ChargeTermID:    "",
				DestTermIDCount: uint8(len(destStrArr)),
				DestTermID:      destStrArr,
				MsgLength:       uint8(len(cont)),
				MsgContent:      cont,
				Reserve:         "",
			}

			_, err = c.SendReqPkt(p)
			if err != nil {
				log.Printf("client %d: send a smgp submit request error: %s.", idx, err)
			} else {
				log.Printf("client %d: send a smgp submit request ok", idx)
			}
			break
		default:
		}

		// recv packets
		i, err := c.RecvAndUnpackPkt(0)
		if err != nil {
			log.Printf("client %d: client read and unpack pkt error: %s.", idx, err)
			break
		}

		switch p := i.(type) {
		case *pkg.SmgpSubmitRespPkt:
			log.Printf("client %d: receive a smgp submit response: %v.", idx, p)

		case *pkg.SmgpDeliverReqPkt:
			log.Printf("client %d: receive a smgp deliver request: %v.", idx, p)
			if p.IsReport == 1 {
				log.Printf("client %d: the smgp deliver request: %d is a status report.", idx, p.MsgID)
			}

		case *pkg.SmgpActiveTestReqPkt:
			log.Printf("client %d: receive a smgp active request: %v.", idx, p)
			rsp := &pkg.SmgpActiveTestRespPkt{}
			err := c.SendRspPkt(rsp, p.SequenceID)
			if err != nil {
				log.Printf("client %d: send smgp active response error: %s.", idx, err)
				break
			}
		case *pkg.SmgpActiveTestRespPkt:
			log.Printf("client %d: receive a smgp active response: %v.", idx, p)

		case *pkg.SmgpExitReqPkt:
			log.Printf("client %d: receive a smgp exit request: %v.", idx, p)
			rsp := &pkg.SmgpExitRespPkt{}
			err := c.SendRspPkt(rsp, p.SequenceID)
			if err != nil {
				log.Printf("client %d: send smgp exit response error: %s.", idx, err)
				break
			}
		case *pkg.SmgpExitRespPkt:
			log.Printf("client %d: receive a smgp exit response: %v.", idx, p)
		}
	}
}

var wg sync.WaitGroup

func main() {
	log.Println("Client example start!")
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go startAClient(i + 1)
	}
	wg.Wait()
	log.Println("Client example ends!")
}
