package jce

import (
	"bytes"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/smartystreets/assertions"
)

func TestHead(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))

	b := NewEncoder(data)
	if err := b.WriteHead(BYTE, 245); err != nil {
		t.Error(err)
	}
	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	d := NewDecoder(data)
	ty, tag, err := d.readHead()
	if err != nil {
		t.Error(err)
	}
	if ty != BYTE && tag != 245 {
		t.Error(err)
	}
}

func TestInt8(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	var tmp int8

	for tag := 0; tag < 250; tag++ {
		for i := math.MinInt8; i <= math.MaxInt8; i++ {
			// [step 1] 编码
			b := NewEncoder(data)
			if err := b.WriteInt8(int8(i), byte(tag)); err != nil {
				t.Error(err)
			}
			if err := b.Flush(); err != nil {
				t.Error(err)
			}

			// [step 2] 解码
			d := NewDecoder(data)
			if err := d.ReadInt8(&tmp, byte(tag), true); err != nil {
				t.Error(err)
			}

			if tmp != int8(i) {
				t.Error("no eq.")
			}
		}
	}
}

func TestInt16(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	var tmp int16

	for tag := 0; tag < 250; tag += 10 {
		for i := math.MinInt16; i <= math.MaxInt16; i++ {
			// [step 1] 编码
			b := NewEncoder(data)
			if err := b.WriteInt16(int16(i), byte(tag)); err != nil {
				t.Error(err)
			}
			if err := b.Flush(); err != nil {
				t.Error(err)
			}

			// [step 2] 解码
			d := NewDecoder(data)
			if err := d.ReadInt16(&tmp, byte(tag), true); err != nil {
				t.Error(err)
			}

			if tmp != int16(i) {
				t.Error("no eq.")
			}

		}
	}
}

func TestInt16_2(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	if err := b.WriteInt16(int16(-1), byte(0)); err != nil {
		t.Error(err)
	}
	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	var tmp int16
	d := NewDecoder(data)
	if err := d.ReadInt16(&tmp, byte(0), true); err != nil {
		t.Error(err)
	}
	if tmp != int16(-1) {
		t.Error("no eq.", tmp)
	}
}

func TestInt32(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	if err := b.WriteInt32(int32(-1), byte(10)); err != nil {
		t.Error(err)
	}

	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	var tmp int32
	d := NewDecoder(data)
	if err := d.ReadInt32(&tmp, byte(10), true); err != nil {
		t.Error(err)
	}
	if tmp != -1 {
		t.Error("no eq.")
	}
}

func TestInt32_2(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	if err := b.WriteInt32(math.MinInt32, byte(10)); err != nil {
		t.Error(err)
	}

	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	var tmp int32
	d := NewDecoder(data)
	if err := d.ReadInt32(&tmp, byte(10), true); err != nil {
		t.Error(err)
	}
	if tmp != math.MinInt32 {
		t.Error("no eq.")
	}
}

func TestInt64(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	if err := b.WriteInt64(math.MinInt64, byte(10)); err != nil {
		t.Error(err)
	}
	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	var tmp int64
	d := NewDecoder(data)
	if err := d.ReadInt64(&tmp, byte(10), true); err != nil {
		t.Error(err)
	}
	if tmp != math.MinInt64 {
		t.Error("no eq.")
	}
}

// test uint8
func TestUint8(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	var tmp uint8

	for tag := 0; tag < 250; tag++ {
		for i := 0; i <= math.MaxUint8; i++ {
			// [step 1] 编码
			b := NewEncoder(data)
			if err := b.WriteUint8(uint8(i), byte(tag)); err != nil {
				t.Error(err)
			}
			if err := b.Flush(); err != nil {
				t.Error(err)
			}

			// [step 2] 解码
			d := NewDecoder(data)
			if err := d.ReadUint8(&tmp, byte(tag), true); err != nil {
				t.Error(err)
			}

			if tmp != uint8(i) {
				t.Error("no eq.")
			}
		}
	}
}

func TestUint16(t *testing.T) {
	data := bytes.NewBuffer(make([]byte, 0))
	var tmp uint16

	for tag := 0; tag < 250; tag += 10 {
		for i := 0; i < math.MaxUint16; i++ {
			// [step 1] 编码
			b := NewEncoder(data)
			if err := b.WriteUint16(uint16(i), byte(tag)); err != nil {
				t.Error(err)
			}
			if err := b.Flush(); err != nil {
				t.Error(err)
			}

			// [step 2] 解码
			d := NewDecoder(data)
			if err := d.ReadUint16(&tmp, byte(tag), true); err != nil {
				t.Error(err)
			}

			if tmp != uint16(i) {
				t.Error("no eq.")
			}
		}
	}
}

func TestUint32(t *testing.T) {
	// [step 1] 编码
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	err := b.WriteUint32(uint32(0xffffffff), byte(10))
	if err != nil {
		t.Error(err)
	}
	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	// [step 2] 解码
	var tmp uint32
	d := NewDecoder(data)
	err = d.ReadUint32(&tmp, byte(10), true)
	if err != nil {
		t.Error(err)
	}
	if tmp != 0xffffffff {
		t.Error("no eq.")
	}
}

func TestUint64(t *testing.T) {
	// [step 1] 编码
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	err := b.WriteUint64(uint64(0xffffffffffffffff), byte(10))
	if err != nil {
		t.Error(err)
	}
	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	// [step 2] 解码
	var tmp uint64
	d := NewDecoder(data)
	err = d.ReadUint64(&tmp, byte(10), true)
	if err != nil {
		t.Error(err)
	}
	if tmp != 0xffffffffffffffff {
		t.Error("no eq.")
	}
}

func TestFloat32(t *testing.T) {
	got := float32(0)
	data := bytes.NewBuffer(make([]byte, 0))

	for i := 0; i < 500; i++ {
		b := NewEncoder(data)
		want := rand.Float32()

		if err := b.WriteFloat32(want, 3); err != nil {
			t.Errorf("Test Write_float32 failed. err:%s\n", err)
		}
		if err := b.Flush(); err != nil {
			t.Error(err)
		}

		d := NewDecoder(data)
		if err := d.ReadFloat32(&got, 3, true); err != nil {
			t.Errorf("Test Read_float32 failed. err:%s\n", err)
		}

		if want != got {
			t.Errorf("Test Write_float32 failed. want:%v, got:%v\n", want, got)
		}
	}
}

func TestFloat64(t *testing.T) {
	got := float64(0)
	data := bytes.NewBuffer(make([]byte, 0))

	for i := 0; i < 1; i++ {
		data.Reset()

		b := NewEncoder(data)
		want := rand.Float64()

		if err := b.WriteFloat64(want, 3); err != nil {
			t.Errorf("Test Write_float64 failed. err:%s\n", err)
		}

		if err := b.Flush(); err != nil {
			t.Error(err)
		}

		d := NewDecoder(data)

		if err := d.ReadFloat64(&got, 3, true); err != nil {
			t.Errorf("Test Read_float64 failed. err:%s\n", err)
		}

		if want != got {
			t.Errorf("Test Write_float64 failed. want:%v, got:%v\n", want, got)
		}
	}
}

// 检测 float64 转换失真问题，所以不能优化为 float32
func TestFloat64toFloat32(t *testing.T) {
	a := float64(0.6046602879796196)
	assertions.ShouldBeTrue(a <= math.MaxFloat32)
	assertions.ShouldEqual(float32(a), 0.6046603)
	assertions.ShouldEqual(float64(float32(a)), 0.6046602725982666)
}

func TestBool(t *testing.T) {
	var got bool
	wants := []bool{true, false}
	data := bytes.NewBuffer(make([]byte, 0))

	for _, want := range wants {
		data.Reset()
		b := NewEncoder(data)
		if err := b.WriteBool(want, 10); err != nil {
			t.Errorf("Test Write_bool failed. err:%s\n", err)
		}

		if err := b.Flush(); err != nil {
			t.Error(err)
		}

		d := NewDecoder(data)

		if err := d.ReadBool(&got, 10, true); err != nil {
			t.Errorf("Test Read_bool failed. err:%s\n", err)
		}

		if got != want {
			t.Errorf("Test Write_bool failed, want:%v, got:%v\n", want, got)
		}
	}
}

func TestString(t *testing.T) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	m := make(map[int]string, 0)

	f := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		return string(b)
	}

	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)

	begin := 0
	end := 9999

	for i := begin; i < end; i++ {
		tmp := f(i)
		m[i] = tmp

		if err := b.WriteString(tmp, byte(i)); err != nil {
			t.Error(err)
		}
	}

	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	d := NewDecoder(data)

	var got string

	for i := begin; i < end; i++ {
		// fmt.Println(i)

		if err := d.ReadString(&got, byte(i), true); err != nil {
			t.Error(err)
			// return
			// panic(err)
		}

		if got != m[i] {
			t.Errorf("want: %s, got: %s", m[i], got)
			// return
			// panic(fmt.Sprintf("want: %s, got: %s", m[i], got))
		}
	}
}

func TestSliceUint8(t *testing.T) {
	var want = []uint8{1, 2, 3, 4, 5}
	var got []uint8

	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	if err := b.WriteSliceUint8(want, byte(len(want))); err != nil {
		t.Errorf("Test Write_slice_uint8 failed. err:%s\n", err)
	}

	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	d := NewDecoder(data)
	if err := d.ReadSliceUint8(&got, byte(len(want)), true); err != nil {
		t.Errorf("Test Read_slice_uint8 failed. err:%s\n", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Test Write_slice_uint8 failed. want:%v, got:%v\n", want, got)
	}
}

func TestSliceInt8(t *testing.T) {
	var want = []int8{1, 2, 3, 4, 5}
	var got []int8

	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)
	if err := b.WriteSliceInt8(want, 0); err != nil {
		t.Errorf("Test Write_slice_int8 failed. err:%s\n", err)
	}

	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	d := NewDecoder(data)

	if err := d.ReadSliceInt8(&got, 0, true); err != nil {
		t.Errorf("Test Read_slice_int8 failed. err:%s\n", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Test Write_slice_int8 failed. want:%v, got:%v\n", want, got)
	}
}

func TestLength(t *testing.T) {

}

// BenchmarkUint32 benchmarks the write and read the uint32 type.
func BenchmarkUint32(t *testing.B) {
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)

	for i := 0; i < 200; i++ {
		err := b.WriteUint32(uint32(0xffffffff), byte(i))
		if err != nil {
			t.Error(err)
		}
	}

	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	d := NewDecoder(data)

	for i := 0; i < 200; i++ {
		var data uint32
		err := d.ReadUint32(&data, byte(i), true)
		if err != nil {
			t.Error(err)
		}
		if data != 0xffffffff {
			t.Error("no eq.")
		}
	}
}

// BenchmarkString benchmark the read and write the string.
func BenchmarkString(t *testing.B) {
	data := bytes.NewBuffer(make([]byte, 0))
	b := NewEncoder(data)

	for i := 0; i < 200; i++ {
		err := b.WriteString("hahahahahahahahahahahahahahahahahahahaha", byte(i))
		if err != nil {
			t.Error(err)
		}
	}

	if err := b.Flush(); err != nil {
		t.Error(err)
	}

	d := NewDecoder(data)

	for i := 0; i < 200; i++ {
		var data string
		err := d.ReadString(&data, byte(i), true)
		if err != nil {
			t.Error(err)
		}
		if data != "hahahahahahahahahahahahahahahahahahahaha" {
			t.Error("no eq.")
		}
	}
}
