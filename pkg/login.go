package pkg

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

const (
	SmgpLoginReqPktLen  = HeaderPktLen + 8 + 16 + 1 + 4 + 1 //42d, 0x2a
	SmgpLoginRespPktLen = HeaderPktLen + 4 + 16 + 1         //33d, 0x21
)

type SmgpLoginReqPkt struct {
	ClientID            string
	AuthenticatorClient string
	LoginMode           uint8
	TimeStamp           uint32
	ClientVersion       uint8

	// used in session
	SequenceID uint32
	Secret     string
}

func (p *SmgpLoginReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpLoginReqPktLen)
	// header
	w.WriteHeader(SmgpLoginReqPktLen, seqId, SMGP_LOGIN)
	p.SequenceID = seqId

	// body
	w.WriteString(p.ClientID)
	if p.TimeStamp == 0 {
		p.TimeStamp = GenTimestamp()
	}

	auth, err := GenAuthenticatorClient(p.ClientID, p.Secret, p.TimeStamp)
	if err != nil {
		return nil, err
	}
	p.AuthenticatorClient = string(auth[:])
	w.WriteString(p.AuthenticatorClient)

	w.WriteInt(binary.BigEndian, p.LoginMode)
	w.WriteInt(binary.BigEndian, p.TimeStamp)
	w.WriteInt(binary.BigEndian, p.ClientVersion)

	return w.Bytes()
}

func (p *SmgpLoginReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	// Body: ClientID
	var sa = make([]byte, 8)
	r.ReadBytes(sa)
	p.ClientID = string(sa)

	// Body: AuthenticatorClient
	var as = make([]byte, 16)
	r.ReadBytes(as)
	p.AuthenticatorClient = string(as)

	// Body: LoginMode
	r.ReadInt(binary.BigEndian, &p.LoginMode)
	// Body: Timestamp
	r.ReadInt(binary.BigEndian, &p.TimeStamp)
	// Body: ClientVersion
	r.ReadInt(binary.BigEndian, &p.ClientVersion)

	return r.Error()
}

func (p *SmgpLoginReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Login Req ---")
	fmt.Fprintln(&b, "ClientID: ", p.ClientID)
	fmt.Fprintln(&b, "Secret: ", p.Secret)
	fmt.Fprintln(&b, "AuthenticatorClient: ", p.AuthenticatorClient)
	fmt.Fprintln(&b, "LoginMode: ", p.LoginMode)
	fmt.Fprintln(&b, "TimeStamp: ", p.TimeStamp)
	fmt.Fprintln(&b, "ClientVersion: ", p.ClientVersion)
	return b.String()
}

type SmgpLoginRespPkt struct {
	Status              Status // 请求返回结果
	AuthenticatorServer string // 服务器端返回给客户端的认证码
	ServerVersion       uint8  // 服务器端支持的最高版本号

	// auth
	Secret              string
	AuthenticatorClient string
	// used in session
	SequenceID uint32
}

func (p *SmgpLoginRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpLoginRespPktLen)
	// header
	w.WriteHeader(SmgpLoginRespPktLen, seqId, SMGP_LOGIN_RESP)
	p.SequenceID = seqId

	// body
	w.WriteInt(binary.BigEndian, p.Status)
	authClient := NewOctetString(p.AuthenticatorClient).Byte(16)
	auth := md5.Sum(bytes.Join([][]byte{{uint8(p.Status.Data())},
		authClient,
		[]byte(p.Secret)},
		nil))

	p.AuthenticatorServer = string(auth[:])
	w.WriteString(NewOctetString(p.AuthenticatorServer).String(16))
	w.WriteInt(binary.BigEndian, p.ServerVersion)

	return w.Bytes()
}

func (p *SmgpLoginRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	// Body: Status
	r.ReadInt(binary.BigEndian, &p.Status)

	// Body: AuthenticatorServer
	var s = make([]byte, 16)
	r.ReadBytes(s)
	p.AuthenticatorServer = string(s)

	// Body: Version
	r.ReadInt(binary.BigEndian, &p.ServerVersion)
	return r.Error()
}

func (p *SmgpLoginRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMGP Login Resp ---")
	fmt.Fprintln(&b, "Status: ", p.Status)
	fmt.Fprintln(&b, "AuthenticatorServer: ", p.AuthenticatorServer)
	fmt.Fprintln(&b, "ServerVersion: ", p.ServerVersion)
	return b.String()
}
