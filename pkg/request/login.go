package request

const (
	SmgpLoginReqPktLen  = HeaderPktLen + 8 + 16 + 1 + 4 + 1 //42d, 0x2a
	SmgpLoginRespPktLen = HeaderPktLen + 4 + 16 + 1         //33d, 0x21
)

type SmgpLoginReqPkt struct {
	ClientID            *OctetString
	AuthenticatorClient *OctetString
	LoginMode           uint8
	TimeStamp           uint32
	ClientVersion       uint8

	// used in session
	SequenceID uint32
}

func (p *SmgpLoginReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmgpLoginReqPktLen)
	// header
	w.WriteHeader(SmgpLoginReqPktLen, seqId, SMGP_LOGIN)

	w.WriteString(p.ClientID.String())
}

type SmgpLoginRespPkt struct {
	Status              Status       // 请求返回结果
	AuthenticatorServer *OctetString // 服务器端返回给客户端的认证码
	ServerVersion       uint8        // 服务器端支持的最高版本号

	// used in session
	SequenceID uint32
}
