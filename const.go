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

func (j JceEncodeType) String() string {
	if int(j) < len(typeToStr) {
		return typeToStr[j]
	}
	return "invalidType"
}

// jce type
const (
	INT1 JceEncodeType = iota
	INT2
	INT4
	INT8
	FLOAT4
	FLOAT8
	ZeroTag
	STRING1
	STRING4
	MAP
	SimpleList
	LIST
	StructBegin
	StructEnd
)

var typeToStr = []string{
	"Int1",
	"Int2",
	"Int4",
	"Int8",
	"Float4",
	"Float8",
	"ZeroTag",
	"String1",
	"String4",
	"Map",
	"SimpleList",
	"List",
	"StructBegin",
	"StructEnd",
}
