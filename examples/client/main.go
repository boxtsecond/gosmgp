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
	loginMode = flag.String("loginMode", "0", "登陆模式")
	spID      = flag.String("spID", "123456", "企业代码")
	spCode    = flag.String("spCode", "123456", "SP的接入号码")
	phone     = flag.String("phone", "8618012345678", "接收手机号码, 86..., 多个使用,分割")
	msg       = flag.String("msg", "验证码：1234", "短信内容")
	//msg = flag.String("msg", "【闪送】您的订单取件密码为 000000 (请妥善保管),订单尾号0000,预计08/16 10:24上门取件。闪送员张师傅,电话8618012345678(本单已开启号码保护,请务必使用本机号码呼叫)。查看闪送员实时位置请点击http://a.bcdefghigk.com/42bag0xV。打开微信-发现-搜一搜-搜索“闪送”查看。", "短信内容")
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

	t := time.NewTicker(time.Second)
	defer t.Stop()
	maxSubmit := 1
	count := 0
	for {
		select {
		case <-t.C:
			if count >= maxSubmit {
				continue
			}

			cont, err := pkg.Utf8ToGB18030(*msg)
			if err != nil {
				fmt.Printf("client %d: utf8 to gb18030 transform err: %s.", idx, err)
				return
			}
			destStrArr := strings.Split(*phone, ",")

			p := &pkg.SmgpSubmitReqPkt{
				MsgType:         pkg.MT,
				NeedReport:      pkg.NEED_REPORT,
				Priority:        pkg.NORMAL_PRIORITY,
				ServiceID:       "",
				FeeType:         "00",
				FeeCode:         "0",
				FixedFee:        "0",
				MsgFormat:       pkg.GB18030,
				ValidTime:       "",
				AtTime:          "",
				SrcTermID:       *spCode,
				ChargeTermID:    "",
				DestTermIDCount: uint8(len(destStrArr)),
				DestTermID:      destStrArr,
				MsgLength:       uint8(len(cont)),
				MsgContent:      cont,
				Reserve:         "",
				//Options: pkg.Options{
				//	pkg.TAG_PkTotal:  pkg.NewTLV(pkg.TAG_PkTotal, []byte{1}),
				//	pkg.TAG_PkNumber: pkg.NewTLV(pkg.TAG_PkNumber, []byte{uint8(1)}),
				//	pkg.TAG_TP_udhi:  pkg.NewTLV(pkg.TAG_TP_udhi, []byte{0}),
				//	pkg.TAG_TP_pid:   pkg.NewTLV(pkg.TAG_TP_pid, []byte{1}),
				//},
			}
			pkgs := make([]*pkg.SmgpSubmitReqPkt, 0)

			if len(cont) > 140 {
				pkgs, err = pkg.GetLongMsgPkgs(p)
				if err != nil {
					log.Printf("client %d: get long msg pkg error: %s.", idx, err)
					continue
				}
			} else {
				p.Options = pkg.Options{
					pkg.TAG_PkTotal:  pkg.NewTLV(pkg.TAG_PkTotal, []byte{1}),
					pkg.TAG_PkNumber: pkg.NewTLV(pkg.TAG_PkNumber, []byte{uint8(1)}),
					pkg.TAG_TP_udhi:  pkg.NewTLV(pkg.TAG_TP_udhi, []byte{0}),
					pkg.TAG_TP_pid:   pkg.NewTLV(pkg.TAG_TP_pid, []byte{1}),
				}
				pkgs = append(pkgs, p)
			}

			for _, req := range pkgs {
				_, err = c.SendReqPkt(req)
			}
			if err != nil {
				log.Printf("client %d: send a smgp submit request error: %s.", idx, err)
				return
			} else {
				log.Printf("client %d: send a smgp submit request ok", idx)
			}
			count += 1
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
			log.Printf("client %d: receive a smgp submit response: \n%v", idx, p)

		case *pkg.SmgpDeliverReqPkt:
			log.Printf("client %d: receive a smgp deliver request: \n%v", idx, p)
			if p.IsReport == 1 {
				log.Printf("client %d: the smgp deliver request: %s is a status report.", idx, p.MsgID)
			}
			rsp := &pkg.SmgpDeliverRespPkt{
				MsgID:  p.MsgID,
				Status: pkg.Status(0),
			}

			err := c.SendRspPkt(rsp, p.SequenceID)
			if err != nil {
				log.Printf("client %d: send smgp deliver response error: %s.", idx, err)
				break
			} else {
				log.Printf("client %d: send smgp deliver response ok.", idx)
			}

		case *pkg.SmgpActiveTestReqPkt:
			log.Printf("client %d: receive a smgp active request.", idx)
			rsp := &pkg.SmgpActiveTestRespPkt{}
			err := c.SendRspPkt(rsp, p.SequenceID)
			if err != nil {
				log.Printf("client %d: send smgp active response error: %s.", idx, err)
				break
			}
		case *pkg.SmgpActiveTestRespPkt:
			log.Printf("client %d: receive a smgp active response.", idx)

		case *pkg.SmgpExitReqPkt:
			log.Printf("client %d: receive a smgp exit request.", idx)
			rsp := &pkg.SmgpExitRespPkt{}
			err := c.SendRspPkt(rsp, p.SequenceID)
			if err != nil {
				log.Printf("client %d: send smgp exit response error: %s.", idx, err)
				break
			}
		case *pkg.SmgpExitRespPkt:
			log.Printf("client %d: receive a smgp exit response.", idx)
		}
	}
}

var wg sync.WaitGroup

func init() {
	flag.Parse()
}

func main() {
	log.Println("Client example start!")
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go startAClient(i + 1)
	}

	wg.Wait()
	log.Println("Client example ends!")
}
