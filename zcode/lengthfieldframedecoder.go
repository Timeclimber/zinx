/**
 * @author uuxia
 * @date 15:57 2023/3/10
 * @description 通用解码器
 **/

package zcode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// EncoderData
// A decoder that splits the received {@link ByteBuf}s dynamically by the
// value of the length field in the message.  It is particularly useful when you
// decode a binary message which has an integer header field that represents the
// length of the message body or the whole message.
// <p>
// {@link LengthFieldBasedFrameDecoder} has many configuration parameters so
// that it can decode any message with a length field, which is often seen in
// proprietary client-server protocols. Here are some example that will give
// you the basic idea on which option does what.
//
// <h3>2 bytes length field at offset 0, do not strip header</h3>
//
// The value of the length field in this example is <tt>12 (0x0C)</tt> which
// represents the length of "HELLO, WORLD".  By default, the decoder assumes
// that the length field represents the number of the bytes that follows the
// length field.  Therefore, it can be decoded with the simplistic parameter
// combination.
// <pre>
// <b>lengthFieldOffset</b>   = <b>0</b>
// <b>lengthFieldLength</b>   = <b>2</b>
// lengthAdjustment    = 0
// initialBytesToStrip = 0 (= do not strip header)
//
// BEFORE DECODE (14 bytes)         AFTER DECODE (14 bytes)
// +--------+----------------+      +--------+----------------+
// | Length | Actual Content |----->| Length | Actual Content |
// | 0x000C | "HELLO, WORLD" |      | 0x000C | "HELLO, WORLD" |
// +--------+----------------+      +--------+----------------+
// </pre>
//
// <h3>2 bytes length field at offset 0, strip header</h3>
//
// Because we can get the length of the content by calling
// {@link ByteBuf#readableBytes()}, you might want to strip the length
// field by specifying <tt>initialBytesToStrip</tt>.  In this example, we
// specified <tt>2</tt>, that is same with the length of the length field, to
// strip the first two bytes.
// <pre>
// lengthFieldOffset   = 0
// lengthFieldLength   = 2
// lengthAdjustment    = 0
// <b>initialBytesToStrip</b> = <b>2</b> (= the length of the Length field)
//
// BEFORE DECODE (14 bytes)         AFTER DECODE (12 bytes)
// +--------+----------------+      +----------------+
// | Length | Actual Content |----->| Actual Content |
// | 0x000C | "HELLO, WORLD" |      | "HELLO, WORLD" |
// +--------+----------------+      +----------------+
// </pre>
//
// <h3>2 bytes length field at offset 0, do not strip header, the length field
//
//	represents the length of the whole message</h3>
//
// In most cases, the length field represents the length of the message body
// only, as shown in the previous examples.  However, in some protocols, the
// length field represents the length of the whole message, including the
// message header.  In such a case, we specify a non-zero
// <tt>lengthAdjustment</tt>.  Because the length value in this example message
// is always greater than the body length by <tt>2</tt>, we specify <tt>-2</tt>
// as <tt>lengthAdjustment</tt> for compensation.
// <pre>
// lengthFieldOffset   =  0
// lengthFieldLength   =  2
// <b>lengthAdjustment</b>    = <b>-2</b> (= the length of the Length field)
// initialBytesToStrip =  0
//
// BEFORE DECODE (14 bytes)         AFTER DECODE (14 bytes)
// +--------+----------------+      +--------+----------------+
// | Length | Actual Content |----->| Length | Actual Content |
// | 0x000E | "HELLO, WORLD" |      | 0x000E | "HELLO, WORLD" |
// +--------+----------------+      +--------+----------------+
// </pre>
//
// <h3>3 bytes length field at the end of 5 bytes header, do not strip header</h3>
//
// The following message is a simple variation of the first example.  An extra
// header value is prepended to the message.  <tt>lengthAdjustment</tt> is zero
// again because the decoder always takes the length of the prepended data into
// account during frame length calculation.
// <pre>
// <b>lengthFieldOffset</b>   = <b>2</b> (= the length of Header 1)
// <b>lengthFieldLength</b>   = <b>3</b>
// lengthAdjustment    = 0
// initialBytesToStrip = 0
//
// BEFORE DECODE (17 bytes)                      AFTER DECODE (17 bytes)
// +----------+----------+----------------+      +----------+----------+----------------+
// | Header 1 |  Length  | Actual Content |----->| Header 1 |  Length  | Actual Content |
// |  0xCAFE  | 0x00000C | "HELLO, WORLD" |      |  0xCAFE  | 0x00000C | "HELLO, WORLD" |
// +----------+----------+----------------+      +----------+----------+----------------+
// </pre>
//
// <h3>3 bytes length field at the beginning of 5 bytes header, do not strip header</h3>
//
// This is an advanced example that shows the case where there is an extra
// header between the length field and the message body.  You have to specify a
// positive <tt>lengthAdjustment</tt> so that the decoder counts the extra
// header into the frame length calculation.
// <pre>
// lengthFieldOffset   = 0
// lengthFieldLength   = 3
// <b>lengthAdjustment</b>    = <b>2</b> (= the length of Header 1)
// initialBytesToStrip = 0
//
// BEFORE DECODE (17 bytes)                      AFTER DECODE (17 bytes)
// +----------+----------+----------------+      +----------+----------+----------------+
// |  Length  | Header 1 | Actual Content |----->|  Length  | Header 1 | Actual Content |
// | 0x00000C |  0xCAFE  | "HELLO, WORLD" |      | 0x00000C |  0xCAFE  | "HELLO, WORLD" |
// +----------+----------+----------------+      +----------+----------+----------------+
// </pre>
//
// <h3>2 bytes length field at offset 1 in the middle of 4 bytes header,
//
//	strip the first header field and the length field</h3>
//
// This is a combination of all the examples above.  There are the prepended
// header before the length field and the extra header after the length field.
// The prepended header affects the <tt>lengthFieldOffset</tt> and the extra
// header affects the <tt>lengthAdjustment</tt>.  We also specified a non-zero
// <tt>initialBytesToStrip</tt> to strip the length field and the prepended
// header from the frame.  If you don't want to strip the prepended header, you
// could specify <tt>0</tt> for <tt>initialBytesToSkip</tt>.
// <pre>
// lengthFieldOffset   = 1 (= the length of HDR1)
// lengthFieldLength   = 2
// <b>lengthAdjustment</b>    = <b>1</b> (= the length of HDR2)
// <b>initialBytesToStrip</b> = <b>3</b> (= the length of HDR1 + LEN)
//
// BEFORE DECODE (16 bytes)                       AFTER DECODE (13 bytes)
// +------+--------+------+----------------+      +------+----------------+
// | HDR1 | Length | HDR2 | Actual Content |----->| HDR2 | Actual Content |
// | 0xCA | 0x000C | 0xFE | "HELLO, WORLD" |      | 0xFE | "HELLO, WORLD" |
// +------+--------+------+----------------+      +------+----------------+
// </pre>
//
// <h3>2 bytes length field at offset 1 in the middle of 4 bytes header,
//
//	strip the first header field and the length field, the length field
//	represents the length of the whole message</h3>
//
// Let's give another twist to the previous example.  The only difference from
// the previous example is that the length field represents the length of the
// whole message instead of the message body, just like the third example.
// We have to count the length of HDR1 and Length into <tt>lengthAdjustment</tt>.
// Please note that we don't need to take the length of HDR2 into account
// because the length field already includes the whole header length.
// <pre>
// lengthFieldOffset   =  1
// lengthFieldLength   =  2
// <b>lengthAdjustment</b>    = <b>-3</b> (= the length of HDR1 + LEN, negative)
// <b>initialBytesToStrip</b> = <b> 3</b>
//
// BEFORE DECODE (16 bytes)                       AFTER DECODE (13 bytes)
// +------+--------+------+----------------+      +------+----------------+
// | HDR1 | Length | HDR2 | Actual Content |----->| HDR2 | Actual Content |
// | 0xCA | 0x0010 | 0xFE | "HELLO, WORLD" |      | 0xFE | "HELLO, WORLD" |
// +------+--------+------+----------------+      +------+----------------+
// https://blog.csdn.net/weixin_45271492/article/details/125347939

type LengthFieldFrameDecoder interface {
	Decode(buff []byte) [][]byte
}
type EncoderData struct {
	//大小端排序
	//大端模式：是指数据的高字节保存在内存的低地址中，而数据的低字节保存在内存的高地址中，地址由小向大增加，而数据从高位往低位放；
	//小端模式：是指数据的高字节保存在内存的高地址中，而数据的低字节保存在内存的低地址中，高地址部分权值高，低地址部分权值低，和我们的日常逻辑方法一致。
	//不了解的自行查阅一下资料
	byteOrder              binary.ByteOrder
	maxFrameLength         int64 //最大帧长度
	lengthFieldOffset      int   //长度字段偏移量
	lengthFieldLength      int   //长度域字段的字节数
	lengthFieldEndOffset   int   //长度字段结束位置的偏移量  lengthFieldOffset+lengthFieldLength
	lengthAdjustment       int   //长度调整
	initialBytesToStrip    int   //需要跳过的字节数
	failFast               bool  //快速失败
	discardingTooLongFrame bool  //true 表示开启丢弃模式，false 正常工作模式
	tooLongFrameLength     int64 //当某个数据包的长度超过maxLength，则开启丢弃模式，此字段记录需要丢弃的数据长度
	bytesToDiscard         int64 //记录还剩余多少字节需要丢弃
	in                     *bytes.Buffer
}

func NewLengthFieldFrameDecoder(maxFrameLength int64, lengthFieldOffset, lengthFieldLength, lengthAdjustment, initialBytesToStrip int) LengthFieldFrameDecoder {
	return &EncoderData{
		maxFrameLength:       maxFrameLength,
		lengthFieldOffset:    lengthFieldOffset,
		lengthFieldLength:    lengthFieldLength,
		lengthAdjustment:     lengthAdjustment,
		initialBytesToStrip:  initialBytesToStrip,
		lengthFieldEndOffset: lengthFieldOffset + lengthFieldLength,
		byteOrder:            binary.BigEndian,
		in:                   bytes.NewBuffer([]byte{}),
	}
}

func (this *EncoderData) fail(frameLength int64) {
	//丢弃完成或未完成都抛异常
	//if frameLength > 0 {
	//	msg := fmt.Sprintf("Adjusted frame length exceeds %d : %d - discarded", this.maxFrameLength, frameLength)
	//	panic(msg)
	//} else {
	//	msg := fmt.Sprintf("Adjusted frame length exceeds %d - discarded", this.maxFrameLength)
	//	panic(msg)
	//}
}

func (this *EncoderData) discardingTooLongFrameFunc(buffer *bytes.Buffer) {
	//保存还需丢弃多少字节
	bytesToDiscard := this.bytesToDiscard
	//获取当前可以丢弃的字节数，有可能出现半包
	localBytesToDiscard := math.Min(float64(bytesToDiscard), float64(buffer.Len()))
	fmt.Println("--->", bytesToDiscard, buffer.Len(), localBytesToDiscard)
	localBytesToDiscard = 2
	//丢弃
	buffer.Next(int(localBytesToDiscard))
	//更新还需丢弃的字节数
	bytesToDiscard -= int64(localBytesToDiscard)
	this.bytesToDiscard = bytesToDiscard
	//是否需要快速失败，回到上面的逻辑
	this.failIfNecessary(false)
}

func (this *EncoderData) getUnadjustedFrameLength(buf *bytes.Buffer, offset int, length int, order binary.ByteOrder) int64 {
	//长度字段的值
	var frameLength int64
	arr := buf.Bytes()
	arr = arr[offset : offset+length]
	buffer := bytes.NewBuffer(arr)
	switch length {
	case 1:
		//byte
		var value byte
		binary.Read(buffer, order, &value)
		frameLength = int64(value)
	case 2:
		//short
		var value int16
		binary.Read(buffer, order, &value)
		frameLength = int64(value)
	case 3:
		//int占32位，这里取出后24位，返回int类型
		if order == binary.LittleEndian {
			n := int(uint(arr[0]) | uint(arr[1])<<8 | uint(arr[2])<<16)
			frameLength = int64(n)
		} else {
			n := int(uint(arr[2]) | uint(arr[1])<<8 | uint(arr[0])<<16)
			frameLength = int64(n)
		}
	case 4:
		//int
		var value int32
		binary.Read(buffer, order, &value)
		frameLength = int64(value)
	case 8:
		//long
		binary.Read(buffer, order, &frameLength)
	default:
		panic(fmt.Sprintf("unsupported lengthFieldLength: %d (expected: 1, 2, 3, 4, or 8)", this.lengthFieldLength))
	}
	return frameLength
}

func (this *EncoderData) failOnNegativeLengthField(in *bytes.Buffer, frameLength int64, lengthFieldEndOffset int) {
	in.Next(lengthFieldEndOffset)
	panic(fmt.Sprintf("negative pre-adjustment length field: %d", frameLength))
}

func (this *EncoderData) failIfNecessary(firstDetectionOfTooLongFrame bool) {
	if this.bytesToDiscard == 0 {
		//说明需要丢弃的数据已经丢弃完成
		//保存一下被丢弃的数据包长度
		tooLongFrameLength := this.tooLongFrameLength
		this.tooLongFrameLength = 0
		//关闭丢弃模式
		this.discardingTooLongFrame = false
		//failFast：默认true
		//firstDetectionOfTooLongFrame：传入true
		if !this.failFast || firstDetectionOfTooLongFrame {
			//快速失败
			this.fail(tooLongFrameLength)
		}
	} else {
		//说明还未丢弃完成
		if this.failFast && firstDetectionOfTooLongFrame {
			//快速失败
			this.fail(this.tooLongFrameLength)
		}
	}
}

// frameLength：数据包的长度
func (this *EncoderData) exceededFrameLength(in *bytes.Buffer, frameLength int64) {
	//数据包长度-可读的字节数  两种情况
	//1. 数据包总长度为100，可读的字节数为50，说明还剩余50个字节需要丢弃但还未接收到
	//2. 数据包总长度为100，可读的字节数为150，说明缓冲区已经包含了整个数据包
	discard := frameLength - int64(in.Len())
	//记录一下最大的数据包的长度
	this.tooLongFrameLength = frameLength
	if discard < 0 {
		//说明是第二种情况，直接丢弃当前数据包
		in.Next(int(frameLength))
	} else {
		//说明是第一种情况，还有部分数据未接收到
		//开启丢弃模式
		this.discardingTooLongFrame = true
		//记录下次还需丢弃多少字节
		this.bytesToDiscard = discard
		//丢弃缓冲区所有数据
		in.Next(in.Len())
	}
	//跟进去
	this.failIfNecessary(true)
}

func (this *EncoderData) failOnFrameLengthLessThanInitialBytesToStrip(in *bytes.Buffer, frameLength int64, initialBytesToStrip int) {
	in.Next(int(frameLength))
	panic(fmt.Sprintf("Adjusted frame length (%d) is less  than initialBytesToStrip: %d", frameLength, initialBytesToStrip))
}

// https://blog.csdn.net/qq_39280718/article/details/125762004
func (this *EncoderData) decode(in *bytes.Buffer) []byte {
	//丢弃模式
	if this.discardingTooLongFrame {
		this.discardingTooLongFrameFunc(in)
	}
	////判断缓冲区中可读的字节数是否小于长度字段的偏移量
	if in.Len() < this.lengthFieldOffset {
		//说明长度字段的包都还不完整，半包
		return nil
	}
	//执行到这，说明可以解析出长度字段的值了

	//计算出长度字段的开始偏移量
	actualLengthFieldOffset := this.lengthFieldOffset
	//获取长度字段的值，不包括lengthAdjustment的调整值
	frameLength := this.getUnadjustedFrameLength(in, actualLengthFieldOffset, this.lengthFieldLength, this.byteOrder)
	//如果数据帧长度小于0，说明是个错误的数据包
	if frameLength < 0 {
		//内部会跳过这个数据包的字节数，并抛异常
		this.failOnNegativeLengthField(in, frameLength, this.lengthFieldEndOffset)
	}

	//套用前面的公式：长度字段后的数据字节数=长度字段的值+lengthAdjustment
	//frameLength就是长度字段的值，加上lengthAdjustment等于长度字段后的数据字节数
	//lengthFieldEndOffset为lengthFieldOffset+lengthFieldLength
	//那说明最后计算出的framLength就是整个数据包的长度
	frameLength += int64(this.lengthAdjustment) + int64(this.lengthFieldEndOffset)
	//丢弃模式就是在这开启的
	//如果数据包长度大于最大长度
	if frameLength > int64(this.maxFrameLength) {
		//对超过的部分进行处理
		this.exceededFrameLength(in, frameLength)
		return nil
	}

	//执行到这说明是正常模式
	//数据包的大小
	frameLengthInt := int(frameLength)
	//判断缓冲区可读字节数是否小于数据包的字节数
	if in.Len() < frameLengthInt {
		//半包，等会再来解析
		return nil
	}

	//执行到这说明缓冲区的数据已经包含了数据包

	//跳过的字节数是否大于数据包长度
	if this.initialBytesToStrip > frameLengthInt {
		this.failOnFrameLengthLessThanInitialBytesToStrip(in, frameLength, this.initialBytesToStrip)
	}
	//跳过initialBytesToStrip个字节
	in.Next(this.initialBytesToStrip)
	//解码
	//获取跳过后的真实数据长度
	actualFrameLength := frameLengthInt - this.initialBytesToStrip
	//提取真实的数据
	buff := make([]byte, actualFrameLength)
	in.Read(buff)
	//bytes.NewBuffer([]byte{})
	//_in := bytes.NewBuffer(buff)
	return buff
}

func (this *EncoderData) Decode(buff []byte) [][]byte {
	this.in.Write(buff)
	resp := make([][]byte, 0)
	for {
		arr := this.decode(this.in)
		if arr != nil {
			resp = append(resp, arr)
		} else {
			return resp
		}
	}
	return nil

}
