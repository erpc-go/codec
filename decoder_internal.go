package jce

import (
	"fmt"
	"io"
)

// ---------------------------------------------------------------------------
// 内部函数
// ---------------------------------------------------------------------------

// 读取一个字节
//
//go:nosplit
func (d *Decoder) readByte() (data uint8, err error) {
	return d.buf.ReadByte()
}

// 读取两个字节
//
//go:nosplit
func (d *Decoder) readByte2() (data uint16, err error) {
	// [step 1] 建立缓冲区
	b := make([]byte, 2)

	// [step 2] 开始读
	if _, err = io.ReadFull(d.buf, b); err != nil {
		return
	}

	// [step 3] 转换字节序
	return d.order.Uint16(b), nil
}

// 读取 4 个字节
//
//go:nosplit
func (d *Decoder) readByte4() (data uint32, err error) {
	// [step 1] 建立缓冲区
	b := make([]byte, 4)

	// [step 2] 开始读
	if _, err = io.ReadFull(d.buf, b); err != nil {
		return
	}

	// [step 3] 转换字节序
	return d.order.Uint32(b), nil
}

// 读取 8 个字节
//
//go:nosplit
func (d *Decoder) readByte8() (data uint64, err error) {
	// [step 1] 建立缓冲区
	b := make([]byte, 8)

	// [step 2] 开始读
	if _, err = io.ReadFull(d.buf, b); err != nil {
		return
	}

	// [step 3] 转换字节序
	return d.order.Uint64(b), nil
}

// readByteN 读取下 n 个字节
//
//go:nosplit
func (d *Decoder) readByteN(n int) (data []byte, err error) {
	// [step 1] 建立缓冲区
	data = make([]byte, n)

	// [step 2] 开始读
	if _, err = io.ReadFull(d.buf, data); err != nil {
		return nil, fmt.Errorf("read n bytes failed, err:%s", err)
	}

	return
}

// 读一个 type,tag
//
//go:nosplit
func (d *Decoder) readHead() (ty JceEncodeType, tag byte, err error) {
	// [step 1] 先读一字节，前 4bit 必然是 type
	data, err := d.readByte()
	if err != nil {
		return 0, 0, err
	}

	// [step 2] 读前 4b 作为 type
	ty = JceEncodeType((data & 0xf0) >> 4)

	// [step 3] 然后读取剩下 4b，根据是否等于 15，来判断 tag 是这个值，还是后面一个字节
	tag = data & 0x0f

	// [step 4] 如果等于 15，说明这个值就是 tag
	if tag != 15 {
		return
	}

	// [step 5] 不然的话，就再读一个字节作为 tag
	tag, err = d.readByte()

	return
}

// unreadHead 回退一个head byte， curTag 为当前读到的tag信息，当tag超过4位时则回退两个head byte
//
//go:nosplit
func (d *Decoder) unreadHead(curTag byte) {
	_ = d.buf.UnreadByte()
	if curTag >= 15 {
		_ = d.buf.UnreadByte()
	}
}

// 跳过 type 类型个字节, 不包括 head 部分
//
// go:nosplit
func (d *Decoder) skipField(ty JceEncodeType) (err error) {
	switch ty {
	case Int1:
		return d.skip(1)
	case Int2:
		return d.skip(2)
	case Int4:
		return d.skip(4)
	case Int8:
		return d.skip(8)
	case Float4:
		return d.skip(4)
	case Float8:
		return d.skip(8)
	case String:
		return d.skipFieldString()
	case Map:
		return d.skipFieldMap()
	case List:
		return d.skipFieldList()
	case SimpleList:
		return d.skipFieldSimpleList()
	case StructBegin:
		return d.skipToStructEnd()
	case StructEnd:
		return
	case Zero:
		return
	default:
		return fmt.Errorf("skip fialed, invalid type")
	}
}

// skip 跳过 n 个字节
//
//go:nosplit
func (d *Decoder) skip(n int) (err error) {
	_, err = d.buf.Discard(n)
	return
}

// 跳过 string 个字节
//
//go:nosplit
func (d *Decoder) skipFieldString() (err error) {
	// [step 1] 读长度
	length, err := d.ReadLength()
	if err != nil {
		return
	}
	// [step 2] 跳
	return d.skip(int(length))
}

// 跳过 map 数据部分字节
//
//go:nosplit
func (d *Decoder) skipFieldMap() (err error) {
	// [step 1] 读 item 的 长度
	length, err := d.ReadLength()
	if err != nil {
		return err
	}

	// [step 2]  扫描 k-v 对,一共 2*length 个
	for i := uint32(0); i < length*2; i++ {
		var t JceEncodeType

		// [step 2.1] 读 head
		t, _, err = d.readHead()
		if err != nil {
			return
		}

		// [step 2.2] 跳数据
		if err = d.skipField(t); err != nil {
			return
		}

	}

	return
}

// 跳过 list 数据部分
//
//go:nosplit
func (d *Decoder) skipFieldList() (err error) {
	// [step 1] 读长度
	length, err := d.ReadLength()
	if err != nil {
		return err
	}

	// [step 2] 跳数据
	for i := uint32(0); i < length; i++ {
		// [step 2.1] 读 head
		var t JceEncodeType
		t, _, err = d.readHead()
		if err != nil {
			return
		}

		// [step 2.2] 跳 data
		_ = d.skipField(t)
	}

	return
}

// 跳过 SimpleList 数据部分
//
//go:nosplit
func (d *Decoder) skipFieldSimpleList() error {
	// [step 1] 读数据长度
	length, err := d.ReadLength()
	if err != nil {
		return err
	}

	// [step 2] 读 item type
	t, err := d.readByte()
	if err != nil {
		return err
	}

	if JceEncodeType(t) != Int1 {
		return fmt.Errorf("simple list need byte head. but get %d", t)
	}

	// [step 3] 跳数据
	return d.skip(int(length))
}

//go:nosplit
func (d *Decoder) skipToStructEnd() (err error) {
	for {
		ty, _, err := d.readHead()
		if err != nil {
			return err
		}

		err = d.skipField(ty)
		if err != nil {
			return err
		}
		if ty == StructEnd {
			break
		}
	}

	return
}
