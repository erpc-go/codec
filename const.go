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
	BYTE JceEncodeType = iota
	SHORT
	INT
	LONG
	FLOAT
	DOUBLE
	STRING1
	STRING4
	MAP
	LIST
	StructBegin
	StructEnd
	ZeroTag
	SimpleList
)

var typeToStr = []string{
	"Byte",
	"Short",
	"Int",
	"Long",
	"Float",
	"Double",
	"String1",
	"String4",
	"Map",
	"List",
	"StructBegin",
	"StructEnd",
	"ZeroTag",
	"SimpleList",
}
