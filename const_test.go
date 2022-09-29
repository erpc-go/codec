// 支持 jce2go 的底层库，用于基础类型的序列化
// 高级类型的序列化，由代码生成器，转换为基础类型的序列化

package jce

import "testing"

func TestJceEncodeType_String(t *testing.T) {
	tests := []struct {
		name string
		j    JceEncodeType
		want string
	}{
		{
			name: "",
			j:    BYTE,
			want: "Byte",
		},
		{
			name: "",
			j:    SimpleList,
			want: "SimpleList",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.j.String(); got != tt.want {
				t.Errorf("JceEncodeType.String() = %s, want %s", got, tt.want)
			}
		})
	}
}
