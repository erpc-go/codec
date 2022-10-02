package jce

import (
	"fmt"
	"math"
)

// ---------------------------------------------------------------------------
// 通用类型解码函数
// ---------------------------------------------------------------------------

//go:nosplit
func (e *Encoder) writeHead(t JceEncodeType, tag byte) (err error) {
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

//go:nosplit
func (e *Encoder) writeLength(length uint32) (err error) {
	// [step 1] 如果可以用 1B 表示，则最高位置 0（默认就是）
	if length <= 127 {
		return e.writeByte(uint8(length))
	}

	// [step 2] 后者最高位置 1
	length |= 0x80000000

	// [step 3] 然后写
	return e.writeByte4(length)
}

//go:nosplit
func (e *Encoder) writeInt1(data uint8, tag byte) (err error) {
	// [step 1] 如果值等于 0，则直接写类型 ZeroTag，后面就不用写数据了(数据压缩优化)
	if data == 0 {
		return e.writeHead(Zero, tag)
	}

	// [step 2] 如果不为 0，则先写 type、tag
	if err = e.writeHead(Int1, tag); err != nil {
		return
	}

	// [step 3] 再写数据
	return e.writeByte(data)
}

//go:nosplit
func (e *Encoder) writeInt2(data uint16, tag byte) (err error) {
	// [step 1] 如果值在 int8 的范围内，则写 int8
	if data <= math.MaxUint8 {
		return e.writeInt1((uint8)(data), tag)
	}

	// [step 2] 否则，先写 type、tag
	if err = e.writeHead(Int2, tag); err != nil {
		return
	}

	// [step 3] 再写数据
	return e.writeByte2(uint16(data))
}

//go:nosplit
func (e *Encoder) writeInt4(data uint32, tag byte) (err error) {
	// [step 1] 如果在 int16 范围内,则写入 int16
	if data <= math.MaxUint16 {
		return e.writeInt2(uint16(data), tag)
	}

	// [step 2] 否则先写 type、tag
	if err = e.writeHead(Int4, tag); err != nil {
		return
	}

	// [step 3] 然后写数据
	return e.writeByte4(uint32(data))
}

//go:nosplit
func (e *Encoder) writeInt8(data uint64, tag byte) (err error) {
	// [step 1] 如果在 int32 范围内，则写入 int32
	if data <= math.MaxUint32 {
		return e.writeInt4(uint32(data), tag)
	}

	// [step 2] 佛则写 type、tag
	if err = e.writeHead(Int8, tag); err != nil {
		return
	}

	// [step 3] 写数据
	return e.writeByte8(uint64(data))
}

//go:nosplit
func (e *Encoder) writeFloat4(data float32, tag byte) (err error) {
	// [step 1] 如果值等于 0，则直接写类型 ZeroTag，后面就不用写数据了(数据压缩优化)
	if data == 0 {
		return e.writeHead(Zero, tag)
	}

	// [step 2] 写 type、tag
	if err = e.writeHead(Float4, tag); err != nil {
		return err
	}

	// [step 3] 然后写数据
	return e.writeByte4(math.Float32bits(data))
}

//go:nosplit
func (e *Encoder) writeFloat8(data float64, tag byte) (err error) {
	// [step 1] 如果值等于 0，则直接写类型 ZeroTag，后面就不用写数据了(数据压缩优化)
	if data == 0 {
		return e.writeHead(Zero, tag)
	}

	// -----------------------------------------------
	// tips: 注意，float 64 不能像 int 一样优化成存 float32，因为 IEEE 浮点数的标准，转换会导致失真,故直接写 double 即可
	// -----------------------------------------------
	// [step 2] 如果值在 float32 的范围内，则写 float32
	// if data >= math.SmallestNonzeroFloat32 && data <= math.MaxFloat32 {
	// return e.writeFloat32(float32(data), tag)
	// }

	// [step 2] 否则写 type、tag
	if err = e.writeHead(Float8, tag); err != nil {
		return
	}

	// [step 3] 然后写数据
	return e.writeByte8(math.Float64bits(data))
}

//go:nosplit
func (e *Encoder) writeStringC(data string, tag byte) (err error) {
	// [step 1] 写头部
	if err = e.WriteHead(String, tag); err != nil {
		return
	}

	// [step 2] 写长度
	if err = e.WriteLength(uint32(len(data))); err != nil {
		return
	}

	// [step 3] 写数据
	return e.writeString(data)
}

//go:nosplit
func (e *Encoder) writeSimpleList(data []uint8, tag byte) (err error) {
	// [step 1] 写 simpleList type、tag
	if err = e.WriteHead(SimpleList, tag); err != nil {
		return fmt.Errorf("write head failed, type:%s, tag:%d ,err: %s", SimpleList, tag, err)
	}

	// [step 2] 写数据长度
	if err = e.writeByte4(uint32(len(data))); err != nil {
		return fmt.Errorf("write list length failed, tag:%d ,err: %s", tag, err)
	}

	// [step 3] 写 list 里的类型
	if err = e.writeByte(uint8(Int1)); err != nil {
		return fmt.Errorf("write list item data type failed, type:%s, tag:%d ,err: %s", Int1, tag, err)
	}

	// [step 4] 写数据
	return e.writeByteN(data)
}
