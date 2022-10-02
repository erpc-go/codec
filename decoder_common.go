package jce

import (
	"fmt"
	"math"
)

// ---------------------------------------------------------------------------
// 通用类型编码函数
// ---------------------------------------------------------------------------

// readHead
//
//go:nosplit
func (d *Decoder) readHeadC(tag byte, require bool) (t JceEncodeType, have bool, err error) {
	for {
		// [step 1] 读取一个 head
		curType, curTag, err := d.readHead()
		if err != nil {
			return curType, false, err
		}

		// [step 2] 如果读到了 struct 的结尾，或者比需要的 tag 还大，说明需要读取的 tag 不存在
		if curType == StructEnd || curTag > tag {
			// [step 2.1] 如果需要存在，但是却不存在，则返回错误
			if require {
				return curType, false, fmt.Errorf("can not find Tag %d. get tag: %d, get type: %d", tag, curTag, curType)
			}
			// [step 2.2] 如果虽然不存在，但是不是必须的，则只返回读取失败即可
			// 多读了一个head, 退回去.
			d.unreadHead(curTag)
			return curType, false, nil
		}

		// [step 3] 如果找到了对应的 tag
		if curTag == tag {
			return curType, true, nil
		}

		// [step 4]  如果现在的 tag 比需要的 tag 小，则需要继续读取，先跳过当前 tag 余下的数据
		if err = d.skipField(curType); err != nil {
			return curType, false, fmt.Errorf("skip type  %s'data filed, err:%s", curType, err)
		}

		// [step 5] 继续读取下一个 tag
	}
}

// 反序列化一个长度字段
//
//go:nosplit
func (d *Decoder) readLength() (length uint32, err error) {
	// [step 1] 先 peek 一个字节
	t, err := d.buf.Peek(1)
	if err != nil {
		return
	}
	// [step 2] 如果这个字节最高位为 0，则说明长度为 1B
	if t[0] <= 127 {
		data, err := d.readByte()
		return uint32(data), err
	}

	// [step 3] 后者说明长度为 4B
	length, err = d.readByte4()
	if err != nil {
		return
	}

	// [step 4] 高位恢复
	length &= 0x7fffffff

	return
}

// readInt1
//
//go:nosplit
func (d *Decoder) readInt1(data *uint8, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	t, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}
	if !have { // tag 不存在,但是不要求必须存在
		return nil
	}

	// [setp 2] 开始读取数据
	switch t {
	case Zero: // 类型是 0
		*data = 0
		return
	case Int1: // 类型是普通的数据，则读取一个字节
		*data, err = d.readByte()
		return
	default: // 如果不是支持的 type
		return fmt.Errorf("read 'int1' type mismatch, tag:%d, get type:%s", tag, t)
	}
}

// 反序列化 int2
//
//go:nosplit
func (d *Decoder) readInt2(data *uint16, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	ty, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}

	if !have { // tag 不存在,但是不要求必须存在
		return nil
	}

	// [setp 2] 读取数据
	switch ty {
	case Zero: // 数据是 0
		*data = 0
		return
	case Int1: // 类型是一个字节
		var tmp uint8
		tmp, err = d.readByte()
		if err != nil {
			return fmt.Errorf("read data failed, when int2'data length is 1byte, err:%s", err)
		}
		*data = uint16(tmp)
		return
	case Int2: // 类型是两个字节
		*data, err = d.readByte2()
		return
	default:
		return fmt.Errorf("read 'int2' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 int4
//
//go:nosplit
func (d *Decoder) readInt4(data *uint32, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	ty, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}

	if !have { // tag 不存在,但是不要求必须存在
		return nil
	}

	// [step 2] 读取数据
	switch ty {
	case Zero: // 0
		*data = 0
		return
	case Int1: // 1byte
		var tmp uint8
		tmp, err = d.readByte()
		if err != nil {
			return fmt.Errorf("read data failed, when int32'data length is 1byte, err:%s", err)
		}
		*data = uint32(tmp)
		return
	case Int2: // 2byte
		var tmp uint16
		tmp, err = d.readByte2()
		if err != nil {
			return fmt.Errorf("read data failed, when int32'data length is 2byte, err:%s", err)
		}
		*data = uint32(tmp)
		return
	case Int4: // 4 byte
		var tmp uint32
		tmp, err = d.readByte4()
		if err != nil {
			return fmt.Errorf("read data failed, when int32'data length is 4byte, err:%s", err)
		}
		*data = tmp
		return
	default:
		return fmt.Errorf("read 'int32' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 int8
//
//go:nosplit
func (d *Decoder) readInt8(data *uint64, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	ty, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}

	if !have { // tag 不存在,但是不要求必须存在
		return nil
	}

	// [step 2] 读取数据
	switch ty {
	case Zero: // 0
		*data = 0
		return
	case Int1: // 1B
		var tmp uint8
		tmp, err = d.readByte()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 1byte, err:%s", err)
		}
		*data = uint64(tmp)
		return
	case Int2: // 2B
		var tmp uint16
		tmp, err = d.readByte2()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 2byte, err:%s", err)
		}
		*data = uint64(tmp)
		return
	case Int4: // 4B
		var tmp uint32
		tmp, err = d.readByte4()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 4byte, err:%s", err)
		}
		*data = uint64(tmp)
		return
	case Int8: // 8B
		var tmp uint64
		tmp, err = d.readByte8()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 8byte, err:%s", err)
		}
		*data = tmp
		return
	default:
		return fmt.Errorf("read 'int64' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 float4
//
//go:nosplit
func (d *Decoder) readFloat4(data *float32, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	ty, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}

	if !have { // tag 不存在,但是不要求必须存在
		return nil
	}

	// [step 2] 读取数据
	switch ty {
	case Zero: // 0
		*data = 0
		return
	case Float4: // 4B
		var tmp uint32
		tmp, err = d.readByte4()
		if err != nil {
			return fmt.Errorf("read data failed, when float32'data length is 4byte, err:%s", err)
		}
		*data = math.Float32frombits(tmp)
		return
	default:
		return fmt.Errorf("read 'float' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 float8
//
//go:nosplit
func (d *Decoder) readFloat8(data *float64, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	ty, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}

	if !have { // tag 不存在,但是不要求必须存在
		return nil
	}

	// [step 2] 读取数据
	switch ty {
	case Zero: // 0
		*data = 0
		return

	// -----------------------------------------------
	// tips: 注意，float 64 不能像 int 一样优化成存 float32，因为 IEEE 浮点数的标准，转换会导致失真,故直接写 double 即可
	// -----------------------------------------------
	// case FLOAT: // 4B
	// 	var tmp uint32
	// 	tmp, err = d.readBytes4()
	// 	if err != nil {
	// 		return fmt.Errorf("read data failed, when float64'data length is 4byte, err:%s", err)
	// 	}
	// 	*data = float64(math.Float32frombits(tmp))
	// 	return
	case Float8: // 8B
		var tmp uint64
		tmp, err = d.readByte8()
		if err != nil {
			return fmt.Errorf("read data failed, when float64'data length is 8byte, err:%s", err)
		}
		*data = math.Float64frombits(tmp)
		return
	default:
		return fmt.Errorf("read 'double' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 string
//
//go:nosplit
func (d *Decoder) readString(data *string, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	t, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}

	if !have { // tag 不存在,但是不要求必须存在
		return
	}

	if t != String {
		return fmt.Errorf("want type %s, but got %s", String, JceEncodeType(t))
	}

	// [step 2] 读长度
	length, err := d.ReadLength()
	if err != nil {
		return fmt.Errorf("read string length failed, tag:%d, err:%s", tag, err)
	}

	var buff []byte

	// [step 3] 读具体数据
	if buff, err = d.readByteN(int(length)); err != nil {
		return fmt.Errorf("read string1' data failed, tag,:%d error:%v", tag, err)
	}

	*data = string(buff)
	return
}

// 反序列化 SimpleList
//
//go:nosplit
func (d *Decoder) readSimpleList(data *[]byte, tag byte, require bool) (err error) {
	// [step 1] 读取 head
	t, have, err := d.ReadHead(tag, require)
	if err != nil { // 读取失败
		return fmt.Errorf("read head failed, tag:%d, err:%s", tag, err)
	}

	if !have { // tag 不存在,但是不要求必须存在
		return nil
	}

	if JceEncodeType(t) != SimpleList {
		return fmt.Errorf("need simpleList type, but %s, tag:%d", t, tag)
	}

	// [step 2] 读数据长度
	length, err := d.readByte4()
	if err != nil {
		return fmt.Errorf("read data item length failed, tag:%d, err:%s", tag, err)
	}

	// [setp 3] 读 item type
	itemType, err := d.readByte()
	if err != nil {
		return fmt.Errorf("read item type failed, tag:%d, err:%s", tag, err)
	}

	if JceEncodeType(itemType) != Int1 {
		return fmt.Errorf("need BYTE byte when read []uint8, tag:%d", tag)
	}

	// [setp 4] 读数据
	if *data, err = d.readByteN(int(length)); err != nil {
		err = fmt.Errorf("read []uint8 error:%v", err)
	}

	return
}

// read struct begin type
//
//go:nosplit
func (d *Decoder) readStructBegin() (err error) {
	data, err := d.readByte()
	if err != nil {
		return
	}
	if data != uint8(StructBegin) {
		return fmt.Errorf("got type %s, but want %s", JceEncodeType(data), StructBegin)
	}
	return
}

// read struct end type
//
//go:nosplit
func (d *Decoder) readStructEnd() (err error) {
	data, err := d.readByte()
	if err != nil {
		return
	}
	if data != uint8(StructEnd) {
		return fmt.Errorf("got type %s, but want %s", JceEncodeType(data), StructEnd)
	}
	return
}
