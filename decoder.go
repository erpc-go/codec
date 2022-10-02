package jce

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

// ---------------------------------------------------------------------------
// go 编码函数、也是对外暴露的 API
// 基本的层次结构：decoder -> decoder_common -> decoder_internal
// ---------------------------------------------------------------------------

// 为什么 read data 时传指针而不是返回值？
// 因为代码生成的时候，如果是 optional，可能有一个默认值，而默认值是反序列化前设置的，
// 所以如果是返回值，那么在这里其实是不知道对应的默认值是多少的，那么就需要再更改代码生成的逻辑，
// 比较麻烦，所以这里暂时传指针，这样如果不需要修改时，就不动指针即可，则默认值也不会变 Decoder 编码器，用于反序列化

type Decoder struct {
	buf   *bufio.Reader
	order binary.ByteOrder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		buf:   bufio.NewReader(r),
		order: defulatByteOrder,
	}
}

// 根据 tag、require 读取对应数据的 type
// 传入 tag 和是否一定的存在
// 返回读取的结果 type，以及 tag 是否存在，最后是是否存在错误
func (d *Decoder) ReadHead(tag byte, require bool) (t JceEncodeType, have bool, err error) {
	return d.readHeadC(tag, require)
}

// 反序列化一个长度字段
func (d *Decoder) ReadLength() (length uint32, err error) {
	return d.readLength()
}

// 反序列化 int8
func (d *Decoder) ReadInt8(data *int8, tag byte, require bool) (err error) {
	return d.readInt1((*uint8)(unsafe.Pointer(data)), tag, require)
}

// 反序列化 uint8
func (d *Decoder) ReadUint8(data *uint8, tag byte, require bool) (err error) {
	return d.readInt1(data, tag, require)
}

// 反序列化 int16
func (d *Decoder) ReadInt16(data *int16, tag byte, require bool) (err error) {
	return d.readInt2((*uint16)(unsafe.Pointer(data)), tag, require)
}

// 反序列化 uint16
func (d *Decoder) ReadUint16(data *uint16, tag byte, require bool) (err error) {
	return d.readInt2(data, tag, require)
}

// 反序列化 int32
func (d *Decoder) ReadInt32(data *int32, tag byte, require bool) (err error) {
	return d.readInt4((*uint32)(unsafe.Pointer(data)), tag, require)
}

// 反序列化 uint32
func (d *Decoder) ReadUint32(data *uint32, tag byte, require bool) (err error) {
	return d.readInt4(data, tag, require)
}

// 反序列化 int64
func (d *Decoder) ReadInt64(data *int64, tag byte, require bool) (err error) {
	return d.readInt8((*uint64)(unsafe.Pointer(data)), tag, require)
}

// 反序列化 uint64
func (d *Decoder) ReadUint64(data *uint64, tag byte, require bool) (err error) {
	return d.readInt8(data, tag, require)
}

// 反序列化 float32
func (d *Decoder) ReadFloat32(data *float32, tag byte, require bool) (err error) {
	return d.readFloat4(data, tag, require)
}

// 反序列化 float64
func (d *Decoder) ReadFloat64(data *float64, tag byte, require bool) (err error) {
	return d.readFloat8((*float64)(unsafe.Pointer(data)), tag, require)
}

// 反序列化 bool
func (d *Decoder) ReadBool(data *bool, tag byte, require bool) (err error) {
	// [step 1] 读取
	var tmp uint8
	if err = d.readInt1(&tmp, tag, require); err != nil {
		return fmt.Errorf("read bool failed, err: %s", err)
	}

	// [step 2] 如果为 0，则为 false
	if tmp == 0 {
		*data = false
		return
	}

	*data = true
	return
}

// 反序列化 string
func (d *Decoder) ReadString(data *string, tag byte, require bool) (err error) {
	return d.readString(data, tag, require)
}

// 反序列化 []uint8
func (d *Decoder) ReadSliceUint8(data *[]uint8, tag byte, require bool) (err error) {
	return d.readSimpleList(data, tag, require)
}

// 反序列化 []int8
func (d *Decoder) ReadSliceInt8(data *[]int8, tag byte, require bool) (err error) {
	return d.readSimpleList((*[]uint8)(unsafe.Pointer(data)), tag, require)
}

// return reader
func (d *Decoder) Reader() (reader *bufio.Reader) {
	return d.buf
}

// read struct begin type
func (d *Decoder) ReadStructBegin() (err error) {
	return d.readStructBegin()
}

// read struct end type
func (d *Decoder) ReadStructEnd() (err error) {
	return d.readStructEnd()
}
