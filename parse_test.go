package parse

import (
	"github.com/influxdata/influxdb/pkg/testing/assert"
	"testing"
)

func TestParse(t *testing.T) {
	//测试用例，input是redis协议字符串，output为解析后的输出，如果output为nil说明input不合法
	tests := []struct {
		input  string
		output interface{}
	}{
		{
			"+ok\r\n", "ok",
		},
		{
			":1000\r\n", 1000,
		},
		{
			"-something wrong\r\n", "something wrong",
		},
		{
			"$11\r\nhello world\r\n", "hello world",
		},
		{
			"*4\r\n$4\r\nname\r\n$2\r\nxc\r\n$4\r\nfrom\r\n$-1\r\n", []interface{}{"name", "xc", "from", nil},
		},
		{
			"ok\r\n", nil,
		},
		{
			"+ok\r", nil,
		},
		{
			"*2\r\n$3\r\nkey\r\n", nil,
		},
	}
	for i, test := range tests {
		output, err := parse(test.input)
		if test.output == nil{
			assert.NotEqual(t, nil, err, i)
		}else{
			assert.Equal(t, test.output, output, i)
		}
	}
}
