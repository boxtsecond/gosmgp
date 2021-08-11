package pkg

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func GenTimestamp() uint32 {
	s := time.Now().Format("0102150405")
	i, _ := strconv.Atoi(s)
	return uint32(i)
}

func GenNowTimeYYYYStr() string {
	s := time.Now().Format("20060102150405")
	return s
}

func GenNowTimeYYStr() string {
	return time.Unix(time.Now().Unix(), 0).Format("0601021504")
}

// 生成客户端认证码
// 其值通过单向MD5 hash计算得出，表示如下：
// AuthenticatorClient =MD5（ClientID+7 字节的二进制0（0x00） + Shared secret+Timestamp）
// Shared secret 由服务器端与客户端事先商定，最长15字节。
// 此处Timestamp格式为：MMDDHHMMSS（月日时分秒），经TimeStamp字段值转换成字符串，转换后右对齐，左补0x30得到。
// 例如3月1日0时0分0秒，TimeStamp字段值为0x11F0E540，此处为0301000000。
func GenAuthenticatorClient(clientId, secret string, timestamp uint32) ([16]byte, error) {
	auth := md5.Sum(bytes.Join([][]byte{NewOctetString(clientId).Byte(8),
		make([]byte, 7),
		[]byte(secret),
		[]byte(fmt.Sprintf("%010d", timestamp))},
		nil))

	return auth, nil
}

//MsgId字段包含以下三部分内容：
//SMGW代码：3字节（BCD码）
//	编码规则如下：
//	3位区号（不足前添0）+2位设备类别+1位序号
//	区号：所在省长途区号
//	设备类别：SMGW取06
//	序号：所在省的设备编码，例如第一个网关编号为1
//时间：4字节（BCD码），格式为MMDDHHMM（月日时分）
//序列号：3字节（BCD码），取值范围为000000～999999，从0开始，顺序累加，步长为1，循环使用。
//例如某SMGW的代码为010061，在2003年1月16日下午5时0分收到一条短消息，这条短消息的MsgID为：0x01006101161700012345，其中010061表示SMGW代码，01161700表示接收时间，012345表示消息序列号。
func GenMsgID(spId string, sequenceNum uint32) (string, error) {
	now := time.Now()
	month, _ := strconv.ParseInt(fmt.Sprintf("%d", now.Month()), 10, 8)
	day := now.Day()
	hour := now.Hour()
	min := now.Minute()
	spIdInt, _ := strconv.ParseInt(spId, 10, 24)
	binaryStr := fmt.Sprintf("%024b%08b%08b%08b%08b%024b", spIdInt, month, day, hour, min, sequenceNum)
	head, _ := strconv.ParseUint(binaryStr[:64], 2, 64)
	end, _ := strconv.ParseUint(binaryStr[64:], 2, 16)
	return fmt.Sprintf("%016x%04x", head, end), nil
}

func UnpackMsgId(msgId string) string {
	spId, _ := strconv.ParseUint(msgId[:6], 16, 24)
	month, _ := strconv.ParseUint(msgId[6:8], 16, 8)
	day, _ := strconv.ParseUint(msgId[8:10], 16, 8)
	hour, _ := strconv.ParseUint(msgId[10:12], 16, 8)
	min, _ := strconv.ParseUint(msgId[12:14], 16, 8)
	seqNum, _ := strconv.ParseUint(msgId[14:], 16, 24)
	return fmt.Sprintf("spId: %s, month: %d, day: %d, hour: %d, min: %d, seqNum: %d, ", NewOctetString(strconv.Itoa(int(spId))).FixedString(6), month, day, hour, min, seqNum)
}

func Utf8ToUcs2(in string) (string, error) {
	if !utf8.ValidString(in) {
		return "", errors.New("invalid utf8 runes")
	}

	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()) //UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Ucs2ToUtf8(in string) (string, error) {
	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()) //UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Utf8ToGB18030(in string) (string, error) {
	if !utf8.ValidString(in) {
		return "", errors.New("invalid utf8 runes")
	}

	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, simplifiedchinese.GB18030.NewEncoder())
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func GB18030ToUtf8(in string) (string, error) {
	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, simplifiedchinese.GB18030.NewDecoder())
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
