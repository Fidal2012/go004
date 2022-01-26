package week09

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// fix length
// 客户端与服务端约定好发送和接收固定字节大小的数据
// 缺点：如果发送数据少于约定约定字节大小则会造成资源浪费
// 理解：适合单包数据需求不大且利用率高的场景，即使出现发送数据不满也不会造成过高浪费的场景
// 有时是业务场景，但我想更多的是受硬件环境约束
// 做过基于MODBUS协议通信模块的数据服务，受改模块约束，每次发送只能发送13字节数据，其中还有5字节帧头
// 结构紧凑，基本发送和接受的数据没有空位。

// delimiter based
// 使用`\n`等符号界定一次完整的请求包
// 缺点：数据量过长会消耗资源在查找界定符上
// 个人理解优点可能是可以达到类似HTTP协议的灵活性

// 这个例子有点粗糙，甚至可能不能正常工作， 重要的是理念=。。=
func delimiterBasedServer(conn net.Conn) {
	reader := bufio.NewReader(conn)
	// 保证reader.ReadSlice出现错误情况下不丢包
	// 应该封装一下，时间短写的比较糙
	bufSlice := make([]byte, 0, 1024)
	r, w := 0, 0
	overFlow := false
	for {
		if overFlow { //上次读取分界符出现err,需获取
			if i := bytes.IndexByte(bufSlice[r:w], '\n'); i >= 0 {
				line := bufSlice[r : r+i+1]
				copy(bufSlice, bufSlice[r+i+1:w])
				w -= r
				r = 0
				if r == w {
					overFlow = false
				}
				serverFunc(line)
			}
		}
		slice, err := reader.ReadSlice('\n')
		if err != nil {
			switch err {
			case bufio.ErrBufferFull:
				overFlow = true
				bufSlice = append(bufSlice, slice...)
				copy(bufSlice[w:], slice)
				w += len(bufSlice)
			case io.EOF:
				overFlow = true
				bufSlice = append(bufSlice, slice...)
				copy(bufSlice[w:], slice)
				w += len(bufSlice)
			}
		} else { // 若溢出buf中数据为`x\nx` 则当前读取数据并非完整数据
			if !overFlow {
				serverFunc(slice)
			} else {
				bufSlice = append(bufSlice, slice...)
				copy(bufSlice[w:], slice)
				w += len(bufSlice)
			}
		}
	}
}

func serverFunc(s []byte) {
	fmt.Println(s)
}

//length field based frame decoder
// 根据消息头数据确定消息头长度和数据长度，进而读取完整信息

func decoderServer(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(Splitting)
	err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		panic(err)
	}
	for scanner.Scan() {
		serverFunc(scanner.Bytes())
	}
}

func Splitting(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !atEOF {
		lenData := len(data)
		var PI uint16

		if lenData > 7 {
			_ = binary.Read(bytes.NewReader(data[2:4]), binary.BigEndian, &PI)
			if PI == 0 {
				length := int16(0)
				_ = binary.Read(bytes.NewReader(data[4:6]), binary.BigEndian, &length)
				if int(length)+6 <= len(data) {
					return int(length) + 6, data[:int(length)+6], nil
				}
			}
		}
	}
	return len(data), data, errors.New("data frame error")
}
