package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSED_Success(t *testing.T) {
	type args struct {
		msg   string
		input string
		expr  string
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args{"New line can be replaced", "first\nsecond\nthird", `:a;N;$!ba;s/\n/,/g`},
			"first,second,third",
		},
		{
			args{"New line can be replaced", "first\nsecond", `:a;N;$!ba;s/\n/,/g`},
			"first,second",
		},
		{
			args{"New line can be replaced", "first", `:a;N;$!ba;s/\n/,/g`},
			"first",
		},
		{
			args{"New line doesn't add at the end", "Hello sed", `s/sed/world/`},
			"Hello world",
		},
		{
			args{"New line doesn't remove from the end", "Hello sed\n", `s/sed/world/`},
			"Hello world\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.args.msg, func(t *testing.T) {
			t.Parallel()
			actual, _, err := SED(tt.args.expr, []byte(tt.args.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(actual))
		})
	}
}
