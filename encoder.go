package jce

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"
)

// Encoder 编码器，用于序列化
type Encoder struct {
	buf   *bufio.Writer
	order binary.ByteOrder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		buf:   bufio.NewWriter(w),
		order: defulatByteOrder,
	}
}

// 序列化 head，即 type+tag
// 方案如下：
// 1. 如果 tag < 15, 则编码为：
// -------------------
// | Type	| Tag    |
// | 4 bits	| 4 bits |
// -------------------
//
// 2. 如果 tag >= 15, 则编码为：
// ----------------------------
// | Type	| Tag 1	 | Tag 2  |
// | 4 bits	| 4 bits | 1 byte |
// ----------------------------
// 其中 tag1 存默认值 15，真正的 tag 值存于 tag2 位置
//
// 为什么要像上面这样设计？而不是直接 type、tag 分别两个字节？
// 主要是考虑到 tag 很可能没有 15 大，只需 4bit 就能编码，而不用 8bit，同时 type 也 4bit 就能放下，那么
// 总的其实 1Byte 就能存，所以就根据 tag 的大小进行了位的压缩
//
//go:nosplit
func (e *Encoder) WriteHead(t JceEncodeType, tag byte) (err error) {
	ty := byte(t)

	// [setp 1] 如果 tag < 15,就直接写一个字节，即 type、tag 各占 4bit
	if tag < 15 {
		return e.writeByte((ty << 4) | tag)
	}

	// [step 2] 如果 tag>=15，则用两个字节，先写 type、15 为一个字节
	if err = e.writeByte((ty << 4) | 15); err != nil {
		return fmt.Errorf("failed to write type byte when tag>=15, err:%s", err)
	}

	// 然后写 tag 为一个字节，共两字节
	return e.writeByte(tag)
}

// 序列化 int8
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteInt8(data int8, tag byte) (err error) {
	// [step 1] 如果值等于 0，则直接写类型 ZeroTag，后面就不用写数据了(数据压缩优化)
	if data == 0 {
		return e.WriteHead(ZeroTag, tag)
	}

	// [step 2] 如果不为 0，则先写 type、tag
	if err = e.WriteHead(BYTE, tag); err != nil {
		return
	}

	// [step 3] 再写数据
	return e.writeByte(uint8(data))
}

// 序列化 int16
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteInt16(data int16, tag byte) (err error) {
	// [step 1] 如果值在 int8 的范围内，则写 int8
	if data >= math.MinInt8 && data <= math.MaxInt8 {
		return e.WriteInt8(int8(data), tag)
	}

	// [step 2] 否则，先写 type、tag
	if err = e.WriteHead(SHORT, tag); err != nil {
		return
	}

	// [step 3] 再写数据
	return e.writeByte2(uint16(data))
}

// 序列化 int32
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteInt32(data int32, tag byte) (err error) {
	// [step 1] 如果在 int16 范围内,则写入 int16
	if data >= math.MinInt16 && data <= math.MaxInt16 {
		return e.WriteInt16(int16(data), tag)
	}

	// [step 2] 否则先写 type、tag
	if err = e.WriteHead(INT, tag); err != nil {
		return
	}

	// [step 3] 然后写数据
	return e.writeByte4(uint32(data))
}

// 序列化 int64
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteInt64(data int64, tag byte) (err error) {
	// [step 1] 如果在 int32 范围内，则写入 int32
	if data >= math.MinInt32 && data <= math.MaxInt32 {
		return e.WriteInt32(int32(data), tag)
	}

	// [step 2] 佛则写 type、tag
	if err = e.WriteHead(LONG, tag); err != nil {
		return
	}

	// [step 3] 写数据
	return e.writeByte8(uint64(data))
}

// 序列化 uint8
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteUint8(data uint8, tag byte) (err error) {
	return e.WriteInt8(int8(data), tag)
}

// 序列化 uint16
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteUint16(data uint16, tag byte) (err error) {
	return e.WriteInt16(int16(data), tag)
}

// 序列化 uint32
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteUint32(data uint32, tag byte) (err error) {
	return e.WriteInt32(int32(data), tag)
}

// 序列化 uint64
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteUint64(data uint64, tag byte) (err error) {
	return e.WriteInt64(int64(data), tag)
}

// 序列化 float32
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteFloat32(data float32, tag byte) (err error) {
	// [step 1] 如果值等于 0，则直接写类型 ZeroTag，后面就不用写数据了(数据压缩优化)
	if data == 0 {
		return e.WriteHead(ZeroTag, tag)
	}

	// [step 2] 写 type、tag
	if err = e.WriteHead(FLOAT, tag); err != nil {
		return err
	}

	// [step 3] 然后写数据
	return e.writeByte4(math.Float32bits(data))
}

// 序列化 float64
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteFloat64(data float64, tag byte) (err error) {
	// [step 1] 如果值等于 0，则直接写类型 ZeroTag，后面就不用写数据了(数据压缩优化)
	if data == 0 {
		return e.WriteHead(ZeroTag, tag)
	}

	// -----------------------------------------------
	// tips: 注意，float 64 不能像 int 一样优化成存 float32，因为 IEEE 浮点数的标准，转换会导致失真,故直接写 double 即可
	// -----------------------------------------------
	// [step 2] 如果值在 float32 的范围内，则写 float32
	// if data >= math.SmallestNonzeroFloat32 && data <= math.MaxFloat32 {
	// return e.WriteFloat32(float32(data), tag)
	// }

	// [step 2] 否则写 type、tag
	if err = e.WriteHead(DOUBLE, tag); err != nil {
		return
	}

	// [step 3] 然后写数据
	return e.writeByte8(math.Float64bits(data))
}

// 序列化 bool
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteBool(data bool, tag byte) (err error) {
	// [step 1] 如果 data 为 true，则写 byte(0),否则写 byte(1)
	tmp := int8(0)
	if data {
		tmp = 1
	}
	return e.WriteInt8(tmp, tag)
}

// TODO: 这里看是不是增加一种情况，只写 data 就行，而不用写什么 tag
// 主要是 vector<string> 这种情况，写内部 string 时，也每次都写了个 tag，都默认是 0，感觉不太好，这个是无效信息
// 序列化 string
// 方案如下：
// |---------------------------------------|
// | type | tag | length(1B or 4B) | data  |
// |---------------------------------------|
// 注意点在于根据长度选择 length 字段的字节数，这个主要是进行了优化
func (e *Encoder) WriteString(data string, tag byte) (err error) {
	// [step 1] 写头部，根据 string 长度来决定长度字段的字节数
	if len(data) <= 255 {
		// [step 1.1.1] 长度 <=255,用 string1 类型，长度用 1B 表示
		// 写 type、tag
		if err = e.WriteHead(STRING1, tag); err != nil {
			return err
		}

		// [step 1.1.2] 写长度 (1B)
		if err = e.writeByte(uint8(len(data))); err != nil {
			return err
		}
	} else {
		// [step 1.2.1] 如果长度大于 255，即长度不能用 1Byte 来表示，则使用 string4 类型，长度用 4Byte 来表示
		// 写 type、tag
		if err = e.WriteHead(STRING4, tag); err != nil {
			return err
		}

		// [step 1.2.1] 写长度 (4B)
		if err = e.writeByte4(uint32(len(data))); err != nil {
			return err
		}
	}

	// [step 2] 写数据
	return e.writeString(data)
}

// TODO: list 的长度是否可以像 string 一样进行优化？即分为两种部分，1B 和 4B 来表示
// 暂时就写死 4B

// []uint8 类型的序列化，方案如下：
// ----------------------------------------------------
// | simpleList head | data length | data type | data |
// ----------------------------------------------------
func (e *Encoder) WriteSliceUint8(data []uint8, tag byte) (err error) {
	// [step 1] 写 simpleList type、tag
	if err = e.WriteHead(SimpleList, tag); err != nil {
		return fmt.Errorf("write head failed, type:%s, tag:%d ,err: %s", SimpleList, tag, err)
	}

	// [step 2] 写数据长度
	if err = e.writeByte4(uint32(len(data))); err != nil {
		return fmt.Errorf("write list length failed, tag:%d ,err: %s", tag, err)
	}

	// [step 3] 写 list 里的类型
	if err = e.writeByte(uint8(BYTE)); err != nil {
		return fmt.Errorf("write list item data type failed, type:%s, tag:%d ,err: %s", BYTE, tag, err)
	}

	// [step 4] 写数据
	return e.writeByteN(data)
}

// []int8 类型的序列化，同 []uint8
func (e *Encoder) WriteSliceInt8(data []int8, tag byte) (err error) {
	return e.WriteSliceUint8(*(*[]uint8)(unsafe.Pointer(&data)), tag)
}

// 序列化一个长度字段
// TODO: 这里看能不能优化一下长度字段？看怎么重新设计一下，把 list、map、string 等都优化在一起
// 现在默认直接用 4B 来表示长度
func (e *Encoder) WriteLength(length uint32) (err error) {
	return e.writeByte4(length)
}

// 将缓存刷新到 writer 中，最后都要手动调这个函数
func (e *Encoder) Flush() (err error) {
	return e.buf.Flush()
}

// return writer
func (e *Encoder) Writer() (writer *bufio.Writer) {
	return e.buf
}

// ---------------------------------------------------------------------------------
// 内部函数

// 写入 n 个字节
//
//go:nosplit
func (e *Encoder) writeByteN(data []byte) (err error) {
	_, err = e.buf.Write(data)
	return
}

// 写入一个字节
//
//go:nosplit
func (e *Encoder) writeByte(data uint8) (err error) {
	return e.buf.WriteByte(data)
}

// 写入两个字节
//
//go:nosplit
func (e *Encoder) writeByte2(data uint16) (err error) {
	// [step 1] 开个 2 字节的缓冲区
	b := make([]byte, 2)

	// [step 2] 转换一下字节序
	e.order.PutUint16(b, data)

	// [step 3] 写
	_, err = e.buf.Write(b)
	return
}

// 写入 4 个字节
//
//go:nosplit
func (e *Encoder) writeByte4(data uint32) (err error) {
	// [step 1] 开个 4 字节的缓冲区
	b := make([]byte, 4)

	// [step 2] 转换一下字节序
	e.order.PutUint32(b, data)

	// [step 3] 写
	_, err = e.buf.Write(b)
	return
}

// 写入 8 个字节
//
//go:nosplit
func (e *Encoder) writeByte8(data uint64) (err error) {
	// [step 1] 开个 8 字节的缓冲区
	b := make([]byte, 8)

	// [step 2] 转换一下字节序
	e.order.PutUint64(b, data)

	// [step 3] 写
	_, err = e.buf.Write(b)
	return
}

// 写入原生 string
//
//go:nosplit
func (e *Encoder) writeString(s string) (err error) {
	_, err = e.buf.WriteString(s)
	return err
}
