// Package protocal 协议包处理类
// 包的格式(LTV) length type from value
package protocal

import (
	"encoding/binary"
	"errors"
	"fmt"
	pb "gameserver/api/protocol"
	"gameserver/pkg/config"
	"gameserver/pkg/json"
	"io"
	"net"
	"reflect"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/golang/protobuf/proto"
)

// 包长度定义
const (
	_lengthSize    = 2    // 消息包长度位所占字节数
	_headerSize    = 4    // 消息包头所占字节数
	_typeSize      = 2    // 消息协议类型所占字节数
	_fromSzie      = 2    // 来源类型所占字节数
	_bodyMaxLength = 2048 // 消息最大长度
)

type ImPacket struct {
	buff []byte
}

// Serialize ..
func (imPacket *ImPacket) Serialize() []byte {
	return imPacket.buff
}

// GetLength 获取消息长度，前2字节
func (imPacket *ImPacket) GetLength() uint16 {
	return binary.BigEndian.Uint16(imPacket.buff[0:_lengthSize])
}

// GetType 获取消息类型，第5-6位置
func (imPacket *ImPacket) GetType() uint16 {
	from := _lengthSize
	to := from + _typeSize
	return binary.BigEndian.Uint16(imPacket.buff[from:to])
}

// GetFrom 获取来源类型
func (imPacket *ImPacket) GetFrom() uint16 {
	from := _lengthSize + _typeSize
	to := from + _fromSzie
	return binary.BigEndian.Uint16(imPacket.buff[from:to])
}

// GetHeader 获取包头，包头包括消息类型、发送者id、接受者id
func (imPacket *ImPacket) GetHeader() []byte {
	from := _lengthSize
	to := _lengthSize + _headerSize
	return imPacket.buff[from:to]
}

// GetBody 获取包内容
func (imPacket *ImPacket) GetBody() []byte {
	from := _lengthSize + _headerSize
	return imPacket.buff[from:]
}

// NewHeader 生成一个包头
func NewHeader(imType uint16, fromType uint16) []byte {
	headerBytes := make([]byte, _headerSize)
	// 消息类型
	binary.BigEndian.PutUint16(headerBytes[0:_typeSize], imType)
	// 来源类型
	binary.BigEndian.PutUint16(headerBytes[_typeSize:], fromType)

	return headerBytes
}

// NewImPacket 生成一条消息
func NewImPacket(header []byte, body []byte) *ImPacket {
	p := &ImPacket{}

	p.buff = make([]byte, _lengthSize+_headerSize+len(body))
	binary.BigEndian.PutUint16(p.buff[0:_lengthSize], _headerSize+uint16(len(body))) // 包头长度 + 协议内容长度

	copy(p.buff[_lengthSize:_lengthSize+_headerSize], header)
	copy(p.buff[_lengthSize+_headerSize:], body)

	return p
}

// ReadPacket 读取一条消息
func ReadPacket(conn *net.TCPConn) (*ImPacket, error) {
	var (
		lengthBytes []byte = make([]byte, _lengthSize)
		headerBytes []byte = make([]byte, _headerSize)
		length      uint16
	)

	// read length
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		if err == io.EOF {
			return nil, err
		}
		errMsg := fmt.Sprintf("read packet length: %s", err.Error())
		return nil, errors.New(errMsg)
	}

	// 包内容的长度最长2048
	length = binary.BigEndian.Uint16(lengthBytes)
	lengthErr := 0
	if length > _bodyMaxLength+_headerSize {
		lengthErr = 1
	}

	// read header
	if _, err := io.ReadFull(conn, headerBytes); err != nil {
		errMsg := fmt.Sprintf("Error: read packet header: %s", err.Error())
		return nil, errors.New(errMsg)
	}

	// read body
	// 扣除包头的长度
	bodyBytes := make([]byte, length-_headerSize)
	if _, err := io.ReadFull(conn, bodyBytes); err != nil {
		errMsg := fmt.Sprintf("Error: read packet body: %s", err.Error())
		return nil, errors.New(errMsg)
	}

	if lengthErr == 1 {
		errMsg := fmt.Sprintf("Error: the size of packet is exceeded the limit:%d, given:%d", _bodyMaxLength, length)
		return nil, errors.New(errMsg)
	}

	return NewImPacket(headerBytes, bodyBytes), nil
}

// SendError 给客户端发送一个错误
func SendError(conn *net.TCPConn, errorCode int, errorMsg string) (*ImPacket, error) {
	temp := new(pb.CommonMsg)
	temp.Code = int32(errorCode)
	temp.Msg = errorMsg
	body, _ := proto.Marshal(temp)
	imPacket, err := SendProto(conn, config.ImError, config.ImFromTypeSytem, body)
	return imPacket, err
}

// SendCommon 给客户端发送一个成功消息
func SendCommon(conn *net.TCPConn, imType uint16, fromType uint16, errorCode int, errorMsg string) (*ImPacket, error) {
	temp := new(pb.CommonMsg)
	temp.Code = int32(errorCode)
	temp.Msg = errorMsg
	body, _ := proto.Marshal(temp)
	imPacket, err := SendProto(conn, imType, fromType, body)
	return imPacket, err
}

// SendSuccess 给客户端发送一个成功的response
func SendSuccess(conn *net.TCPConn, imType uint16, token string, responseCode int) (*ImPacket, error) {
	// 生成协议内容

	body := make(map[string]interface{})
	body["imType"] = imType
	body["token"] = token
	body["code"] = responseCode

	// 发送消息
	imPacket, err := Send(conn, config.ImResponse, config.ImFromTypeSytem, body)

	return imPacket, err
}

// SendSuccessWithExtra 给客户端发送一个成功的response，不同的是，可以支持客户端扩展参数
func SendSuccessWithExtra(conn *net.TCPConn, imType uint16, token string, responseCode int, extra map[string]interface{}) (*ImPacket, error) {
	// 生成协议内容
	body := make(map[string]interface{})
	body["imType"] = imType
	body["token"] = token
	body["code"] = responseCode

	// 附加扩展参数
	if len(extra) > 0 {
		for key, value := range extra {
			body[key] = value
		}
	}

	// 发送消息
	imPacket, err := Send(conn, config.ImResponse, config.ImFromTypeSytem, body)

	return imPacket, err
}

// Send 发送消息封装
func Send(conn *net.TCPConn, imType uint16, fromType uint16, body map[string]interface{}) (*ImPacket, error) {
	// 生成协议头
	headerBytes := NewHeader(imType, fromType)
	// 生成协议内容
	bodyBytes, _ := json.Encode(body)

	// 生成完整包数据
	imPacket := NewImPacket(headerBytes, bodyBytes)

	now := time.Now()
	begin := now.Local().UnixNano() / (1000 * 1000)

	// 发送消息
	if conn == nil {
		return imPacket, errors.New("conn is nil")
	}
	if _, err := conn.Write(imPacket.Serialize()); err != nil {
		// 这个错误是不能转换*net.OpError
		if err == syscall.EINVAL {
			return imPacket, errors.New("syscall.EINVAL")
		}
		// 转换成*net.OpError
		opErr := (*net.OpError)(unsafe.Pointer(reflect.ValueOf(err).Pointer()))
		if opErr.Err.Error() == "i/o timeout" {
			end := time.Now().Local().UnixNano() / (1000 * 1000)
			fmt.Printf("Write timeout! end: %d, begin: %d, timeOut: %dms", end, begin, end-begin)
		}
		return imPacket, errors.New(opErr.Err.Error())
	}

	return imPacket, nil
}

// Send 发送消息封装
func SendProto(conn *net.TCPConn, imType uint16, fromType uint16, bodyBytes []byte) (*ImPacket, error) {
	// 生成协议头
	headerBytes := NewHeader(imType, fromType)
	// 生成协议内容

	// 生成完整包数据
	imPacket := NewImPacket(headerBytes, bodyBytes)

	now := time.Now()
	begin := now.Local().UnixNano() / (1000 * 1000)

	// 发送消息
	if conn == nil {
		return imPacket, errors.New("conn is nil")
	}
	if _, err := conn.Write(imPacket.Serialize()); err != nil {
		// 这个错误是不能转换*net.OpError
		if err == syscall.EINVAL {
			return imPacket, errors.New("syscall.EINVAL")
		}
		// 转换成*net.OpError
		opErr := (*net.OpError)(unsafe.Pointer(reflect.ValueOf(err).Pointer()))
		if opErr.Err.Error() == "i/o timeout" {
			end := time.Now().Local().UnixNano() / (1000 * 1000)
			fmt.Printf("Write timeout! end: %d, begin: %d, timeOut: %dms", end, begin, end-begin)
		}
		return imPacket, errors.New(opErr.Err.Error())
	}

	return imPacket, nil
}

// GetBodyUint16 获取body中的int值
func GetBodyUint16(body map[string]interface{}, key string) (uint16, bool) {
	val, exists := body[key]
	if !exists {
		return 0, false
	}
	return uint16(val.(float64)), true
}

// GetBodyInt 获取body中的int值
func GetBodyInt(body map[string]interface{}, key string) (int, bool) {
	val, exists := body[key]
	if !exists {
		return 0, false
	}
	return int(val.(float64)), true
}

// GetBodyString 获取body中的string值
func GetBodyString(body map[string]interface{}, key string) (string, bool) {
	val, exists := body[key]
	if !exists {
		return "", false
	}
	return val.(string), true
}

// GetUserID 获取userID
func GetUserID(body map[string]interface{}, key string) (int64, bool) {
	userIDStr, exists := GetBodyString(body, key)
	if !exists {
		return 0, false
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, false
	}
	return int64(userID), true
}
