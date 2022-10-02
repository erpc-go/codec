// 支持 jce2go 的底层库，用于基础类型的序列化
// 高级类型的序列化，由代码生成器，转换为基础类型的序列化

package jce

import (
	"encoding/binary"
)

// 默认序列化字节序为大端
var (
	defulatByteOrder = binary.BigEndian
)

// jce 基础编码类型表，用来编码使用，和语言无关
type JceEncodeType byte

// jce type
const (
	Int1 JceEncodeType = iota
	Int2
	Int4
	Int8
	Float4
	Float8
	Zero
	String
	Map
	SimpleList
	List
	StructBegin
	StructEnd
)

func (j JceEncodeType) String() string {
	switch j {
	case Int1:
		return "Int1"
	case Int2:
		return "Int2"
	case Int4:
		return "Int4"
	case Int8:
		return "Int8"
	case Float4:
		return "Float4"
	case Float8:
		return "Float8"
	case Zero:
		return "Zero"
	case String:
		return "String"
	case Map:
		return "Map"
	case SimpleList:
		return "SimpleList"
	case List:
		return "List"
	case StructBegin:
		return "StructBegin"
	case StructEnd:
		return "StructEnd"
	default:
		return "invalidType"
	}
}
