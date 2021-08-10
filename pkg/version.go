package pkg

const (
	VERSION uint8 = 0x30
)

const (
	SEND_MODE = iota
	RECEIVE_MODE
	TRANSMIT_MODE
)

// MsgType
const (
	MO = 0 // MO消息（终端发给SP）
	MT = 6 // MT消息（SP发给终端，包括WEB上发送的点对点短消息）
)

// MsgFormat
// 短消息内容体的编码格式
// 对于文字短消息，要求MsgFormat＝15, 对于回执消息，要求MsgFormat＝0
const (
	ASCII   = 0  // ASCII编码
	BINARY  = 4  // 二进制短消息
	UCS2    = 8  // UCS2编码
	GB18030 = 15 // GB18030编码
)

const (
	NOT_REPORT = 0 // 不是状态报告
	IS_REPORT  = 1 // 是状态报告
)

// 是否要求返回状态报告
const (
	NO_NEED_REPORT = 0
	NEED_REPORT    = 1
)
