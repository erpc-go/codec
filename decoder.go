package jce

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"
)

// TODO: 考虑优化长度字段？string 和 list、map 等的长度字段怎么优化？单独写一个 readLength()、writeLength()
// 还是像现有 string 一样拆分 type？

// Decoder 编码器，用于反序列化
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

// TODO: 这里 data 为什么传指针而不是返回值？因为代码生成的时候，如果是 optional，可能有一个默认值，而默认值是
// 反序列化前设置的，所以如果是做返回值，那么在这里其实是不知道对应的默认值是多少的，那么就需要再更改
// 代码生成的逻辑，比较麻烦，所以这里暂时传指针，这样如果不需要修改时，就不动指针即可，则默认值也不会变

// 反序列化 int8
func (d *Decoder) ReadInt8(data *int8, tag byte, require bool) (err error) {
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
	case ZeroTag: // 类型是 0
		*data = 0
		return nil
	case INT1: // 类型是普通的数据，则读取一个字节
		var tmp uint8
		tmp, err = d.readByte()
		if err != nil {
			return fmt.Errorf("read 'int8' failed, tag:%d error:%v", tag, err)
		}
		*data = int8(tmp)
		return
	default: // 如果不是支持的 type
		return fmt.Errorf("read 'int8' type mismatch, tag:%d, get type:%s", tag, t)
	}
}

// 反序列化 int16
func (d *Decoder) ReadInt16(data *int16, tag byte, require bool) (err error) {
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
	case ZeroTag: // 数据是 0
		*data = 0
		return
	case INT1: // 类型是一个字节
		var tmp uint8
		tmp, err = d.readByte()
		if err != nil {
			return fmt.Errorf("read data failed, when int16'data length is 1byte, err:%s", err)
		}
		*data = int16(int8(tmp))
		return
	case INT2: // 类型是两个字节
		var tmp uint16
		tmp, err = d.readByte2()
		if err != nil {
			return fmt.Errorf("read data failed, when int16'data length is 2byte, err:%s", err)
		}
		*data = int16(tmp)
		return
	default:
		return fmt.Errorf("read 'int16' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 int32
func (d *Decoder) ReadInt32(data *int32, tag byte, require bool) (err error) {
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
	case ZeroTag: // 0
		*data = 0
		return
	case INT1: // 1byte
		var tmp uint8
		tmp, err = d.readByte()
		if err != nil {
			return fmt.Errorf("read data failed, when int32'data length is 1byte, err:%s", err)
		}
		*data = int32(int8(tmp))
		return
	case INT2: // 2byte
		var tmp uint16
		tmp, err = d.readByte2()
		if err != nil {
			return fmt.Errorf("read data failed, when int32'data length is 2byte, err:%s", err)
		}
		*data = int32(int16(tmp))
		return
	case INT4: // 4 byte
		var tmp uint32
		tmp, err = d.readByte4()
		if err != nil {
			return fmt.Errorf("read data failed, when int32'data length is 4byte, err:%s", err)
		}
		*data = int32(tmp)
		return
	default:
		return fmt.Errorf("read 'int32' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 int64
func (d *Decoder) ReadInt64(data *int64, tag byte, require bool) (err error) {
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
	case ZeroTag: // 0
		*data = 0
		return
	case INT1: // 1B
		var tmp uint8
		tmp, err = d.readByte()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 1byte, err:%s", err)
		}
		*data = int64(int8(tmp))
		return
	case INT2: // 2B
		var tmp uint16
		tmp, err = d.readByte2()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 2byte, err:%s", err)
		}
		*data = int64(int16(tmp))
		return
	case INT4: // 4B
		var tmp uint32
		tmp, err = d.readByte4()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 4byte, err:%s", err)
		}
		*data = int64(int32(tmp))
		return
	case INT8: // 8B
		var tmp uint64
		tmp, err = d.readByte8()
		if err != nil {
			return fmt.Errorf("read data failed, when int64'data length is 8byte, err:%s", err)
		}
		*data = int64(tmp)
		return
	default:
		return fmt.Errorf("read 'int64' type mismatch, tag:%d, get type:%s", tag, ty)
	}
}

// 反序列化 uint8
func (d *Decoder) ReadUint8(data *uint8, tag byte, require bool) (err error) {
	n := int8(*data)
	if err := d.ReadInt8(&n, tag, require); err != nil {
		return fmt.Errorf("read uint8 failed, err:%s", err)
	}
	*data = uint8(n)
	return
}

// 反序列化 uint16
func (d *Decoder) ReadUint16(data *uint16, tag byte, require bool) (err error) {
	n := int16(*data)
	if err := d.ReadInt16(&n, tag, require); err != nil {
		return fmt.Errorf("read uint16 failed, err:%s", err)
	}
	*data = uint16(n)
	return err
}

// 反序列化 uint32
func (d *Decoder) ReadUint32(data *uint32, tag byte, require bool) (err error) {
	n := int32(*data)
	if err := d.ReadInt32(&n, tag, require); err != nil {
		return fmt.Errorf("read uint32 failed, err:%s", err)
	}
	*data = uint32(n)
	return err
}

// 反序列化 uint64
func (d *Decoder) ReadUint64(data *uint64, tag byte, require bool) (err error) {
	n := int64(*data)
	if err := d.ReadInt64(&n, tag, require); err != nil {
		return fmt.Errorf("read uint64 failed, err:%s", err)
	}
	*data = uint64(n)
	return err
}

// 反序列化 float32
func (d *Decoder) ReadFloat32(data *float32, tag byte, require bool) (err error) {
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
	case ZeroTag: // 0
		*data = 0
		return
	case FLOAT4: // 4B
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

// 反序列化 float64
func (d *Decoder) ReadFloat64(data *float64, tag byte, require bool) (err error) {
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
	case ZeroTag: // 0
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
	case FLOAT8: // 8B
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

// 反序列化 bool
func (d *Decoder) ReadBool(data *bool, tag byte, require bool) (err error) {
	// [step 1] 读取
	var tmp int8
	err = d.ReadInt8(&tmp, tag, require)
	if err != nil {
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

// ReadString reads the string value for the tag and the require or optional sign.
func (d *Decoder) ReadString(data *string, tag byte, require bool) (err error) {
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

// TODO: 这里看是不是要优化一下？把代码生成的逻辑放到基础部分里来
// 反序列化 []uint8
func (d *Decoder) ReadSliceUint8(data *[]uint8, tag byte, require bool) (err error) {
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

	if JceEncodeType(itemType) != INT1 {
		return fmt.Errorf("need BYTE byte when read []uint8, tag:%d", tag)
	}

	// [setp 4] 读数据
	if *data, err = d.readByteN(int(length)); err != nil {
		err = fmt.Errorf("read []uint8 error:%v", err)
	}

	return
}

// 反序列化 []int8
func (d *Decoder) ReadSliceInt8(data *[]int8, tag byte, require bool) (err error) {
	var tmp []uint8
	if err = d.ReadSliceUint8(&tmp, tag, require); err != nil {
		return fmt.Errorf("read []int8 failed, tag:%d, err:%s", tag, err)
	}
	*data = *(*[]int8)(unsafe.Pointer(&tmp))
	return
}

// 反序列化一个长度字段
func (d *Decoder) ReadLength() (length uint32, err error) {
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

func (d *Decoder) ReadByte() (b byte, err error) {
	return d.readByte()
}

// return reader
func (d *Decoder) Reader() (reader *bufio.Reader) {
	return d.buf
}

// read struct begin type
func (d *Decoder) ReadStructBegin() (err error) {
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
func (d *Decoder) ReadStructEnd() (err error) {
	data, err := d.readByte()
	if err != nil {
		return
	}
	if data != uint8(StructEnd) {
		return fmt.Errorf("got type %s, but want %s", JceEncodeType(data), StructEnd)
	}

	return
}

// ---------------------------------------------------------------------------
// 内部函数

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

// 反序列化 head，即 type,tag
// 方案参考 encoder.go WriteHead()
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
// j// go:nosplit
func (d *Decoder) skipField(ty JceEncodeType) (err error) {
	switch ty {
	case INT1:
		return d.skip(1)
	case INT2:
		return d.skip(2)
	case INT4:
		return d.skip(4)
	case INT8:
		return d.skip(8)
	case FLOAT4:
		return d.skip(4)
	case FLOAT8:
		return d.skip(8)
	case String:
		return d.skipFieldString()
	case MAP:
		return d.skipFieldMap()
	case LIST:
		return d.skipFieldList()
	case SimpleList:
		return d.skipFieldSimpleList()
	case StructBegin:
		return d.skipToStructEnd()
	case StructEnd:
		return
	case ZeroTag:
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

	if JceEncodeType(t) != INT1 {
		return fmt.Errorf("simple list need byte head. but get %d", t)
	}

	// [step 3] 跳数据
	return d.skip(int(length))
}

// skipToStructEnd for skip to the StructEnd tag.
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
