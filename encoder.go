package jce

import (
	"bufio"
	"encoding/binary"
	"io"
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
func (e *Encoder) WriteHead(t JceEncodeType, tag byte) (err error) {
	return e.writeHead(t, tag)
}

// 序列化 int8
// 方案如下：
// |----------------------|
// | type  | tag |  data  |
// |----------------------|
func (e *Encoder) WriteInt8(data int8, tag byte) (err error) {
	return e.writeInt1((uint8)(data), tag)
}

// 序列化 uint8
func (e *Encoder) WriteUint8(data uint8, tag byte) (err error) {
	return e.writeInt1(data, tag)
}

// 序列化 int16
func (e *Encoder) WriteInt16(data int16, tag byte) (err error) {
	return e.writeInt2((uint16)(data), tag)
}

// 序列化 uint16
func (e *Encoder) WriteUint16(data uint16, tag byte) (err error) {
	return e.writeInt2(data, tag)
}

// 序列化 int32
func (e *Encoder) WriteInt32(data int32, tag byte) (err error) {
	return e.writeInt4((uint32)(data), tag)
}

// 序列化 uint32
func (e *Encoder) WriteUint32(data uint32, tag byte) (err error) {
	return e.writeInt4(data, tag)
}

// 序列化 int64
func (e *Encoder) WriteInt64(data int64, tag byte) (err error) {
	return e.writeInt8((uint64)(data), tag)
}

// 序列化 uint64
func (e *Encoder) WriteUint64(data uint64, tag byte) (err error) {
	return e.writeInt8(data, tag)
}

// 序列化 float32
func (e *Encoder) WriteFloat32(data float32, tag byte) (err error) {
	return e.writeFloat4(data, tag)
}

// 序列化 float64
func (e *Encoder) WriteFloat64(data float64, tag byte) (err error) {
	return e.writeFloat8(data, tag)
}

// 序列化 bool
func (e *Encoder) WriteBool(data bool, tag byte) (err error) {
	// [step 1] 如果 data 为 true，则写 byte(0),否则写 byte(1)
	tmp := uint8(0)
	if data {
		tmp = 1
	}
	return e.writeInt1(tmp, tag)
}

// 主要是 vector<string> 这种情况，写内部 string 时，也每次都写了个 tag，都默认是 0，感觉不太好，这个是无效信息
// 序列化 string
// 方案如下：
// |---------------------------------------|
// | type | tag | length(1B or 4B) | data  |
// |---------------------------------------|
// 注意点在于根据长度选择 length 字段的字节数，这个主要是进行了优化
func (e *Encoder) WriteString(data string, tag byte) (err error) {
	return e.writeStringC(data, tag)
}

// []uint8 类型的序列化，方案如下：
// ----------------------------------------------------
// | simpleList head | data length | data type | data |
// ----------------------------------------------------
func (e *Encoder) WriteSliceUint8(data []uint8, tag byte) (err error) {
	return e.writeSimpleList(data, tag)
}

// []int8 类型的序列化，同 []uint8
func (e *Encoder) WriteSliceInt8(data []int8, tag byte) (err error) {
	return e.writeSimpleList(*(*[]uint8)(unsafe.Pointer(&data)), tag)
}

// 序列化一个长度字段
func (e *Encoder) WriteLength(length uint32) (err error) {
	return e.writeLength(length)
}

// 将缓存刷新到 writer 中，最后都要手动调这个函数
func (e *Encoder) Flush() (err error) {
	return e.buf.Flush()
}

// return writer
func (e *Encoder) Writer() (writer *bufio.Writer) {
	return e.buf
}

// write struct begin type
func (e *Encoder) WriteStructBegin() (err error) {
	return e.writeByte(uint8(StructBegin))
}

// write struct end type
func (e *Encoder) WriteStructEnd() (err error) {
	return e.writeByte(uint8(StructEnd))
}
