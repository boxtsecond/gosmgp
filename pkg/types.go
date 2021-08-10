package pkg

import (
	"encoding/binary"
	"strings"
)

// 不强制以0x00结尾的定长字符串。
// 当位数不足时，在不明确注明的
// 情况下， 应左对齐，右补0x00。
// 在明确注明的情况下，以该字段的明确注明为准。
type OctetString string

func NewOctetString(o string) OctetString {
	return OctetString(o)
}

// 去除补零，转为字符串
func (o OctetString) String(fixedLength int) string {
	length := len(o)
	s := o
	if length == fixedLength {
		return string(s)
	}

	if length > fixedLength {
		return string(s[length-fixedLength:])
	}

	return strings.Join([]string{string(s), string(make([]byte, fixedLength-length))}, "")
}

// 按需补零
func (o OctetString) Byte(fixedLength int) []byte {
	data := []byte(o)
	if len(data) < fixedLength {
		// fill 0x00
		tmp := make([]byte, fixedLength-len(data))
		data = append(data, tmp...)
	}

	if len(data) > fixedLength {
		data = data[0:fixedLength]
	}

	return data
}

// 可选参数采用TLV（Tag、Length、Value）形式定义，
// 每个可选参数的Tag、Length、Value的定义见6.3节。
type TLV struct {
	// 字段的标签，用于唯一标识可选参数
	Tag Tag
	// 字段的长度
	Length uint16
	// 可变类型 字段内容
	Value []byte
}

func NewTLV(tag Tag, value []byte) *TLV {
	return &TLV{
		Tag:    tag,
		Length: uint16(len(value)),
		Value:  value,
	}
}

// 序列化为字节流
func (t *TLV) Byte() []byte {
	b := []byte{}
	b = append(b, packUi16(uint16(t.Tag))...)
	b = append(b, packUi16(t.Length)...)
	b = append(b, t.Value...)
	return b
}

func (t *TLV) String() string {
	return ""
}

func packUi16(n uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	return b
}
