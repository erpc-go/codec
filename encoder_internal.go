package jce

// ---------------------------------------------------------------------------------
// 内部函数
// ---------------------------------------------------------------------------------

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
