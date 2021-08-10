package pkg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

var ErrLength = errors.New("Options: error length")

type Tag uint16

// 可选参数标签定义  Option Tag
const (
	TAG_TP_pid Tag = 0x0001 + iota
	TAG_TP_udhi
	TAG_LinkID
	TAG_ChargeUserType
	TAG_ChargeTermType
	TAG_ChargeTermPseudo
	TAG_DestTermType
	TAG_DestTermPseudo
	TAG_PkTotal
	TAG_PkNumber
	TAG_SubmitMsgType
	TAG_SPDealResult
	TAG_SrcTermType
	TAG_SrcTermPseudo
	TAG_NodesCount
	TAG_MsgSrc
	TAG_SrcType
	TAG_MServiceID
)

var TagName = map[Tag]string{
	TAG_TP_pid:           "TAG_TP_pid",
	TAG_TP_udhi:          "TAG_TP_udhi",
	TAG_LinkID:           "TAG_LinkID",
	TAG_ChargeUserType:   "TAG_ChargeUserType",
	TAG_ChargeTermType:   "TAG_ChargeTermType",
	TAG_ChargeTermPseudo: "TAG_ChargeTermPseudo",
	TAG_DestTermType:     "TAG_DestTermType",
	TAG_PkTotal:          "TAG_PkTotal",
	TAG_PkNumber:         "TAG_PkNumber",
	TAG_SubmitMsgType:    "TAG_SubmitMsgType",
	TAG_SPDealResult:     "TAG_SPDealResult",
	TAG_SrcTermType:      "TAG_SrcTermType",
	TAG_SrcTermPseudo:    "TAG_SrcTermPseudo",
	TAG_NodesCount:       "TAG_NodesCount",
	TAG_MsgSrc:           "TAG_MsgSrc",
	TAG_SrcType:          "TAG_SrcType",
	TAG_MServiceID:       "TAG_MServiceID",
}

// 可选参数 map
type Options map[Tag]*TLV

// 返回可选字段部分的长度
func (o Options) Len() int {
	length := 0

	for _, v := range o {
		length += 2 + 2 + int(v.Length)
	}

	return length
}

func (o Options) String() string {
	var b bytes.Buffer

	for k, _ := range o {
		fmt.Fprintln(&b, "--- Options ---")
		fmt.Fprintln(&b, "Tag: ", TagName[k])
	}
	return b.String()
}

func ParseOptions(rawData []byte) (Options, error) {
	var (
		p      = 0
		ops    = make(Options)
		length = len(rawData)
	)

	for p < length {
		if length-p < 2+2 { // less than Tag len + Length len
			return nil, ErrLength
		}

		tag := binary.BigEndian.Uint16(rawData[p:])
		p += 2

		vlen := binary.BigEndian.Uint16(rawData[p:])
		p += 2

		if length-p < int(vlen) { // remaining not enough
			return nil, ErrLength
		}

		value := rawData[p : p+int(vlen)]
		p += int(vlen)

		ops[Tag(tag)] = NewTLV(Tag(tag), value)
	}

	return ops, nil
}
