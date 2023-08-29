package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRespEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				HeadLength: 120,
				RequestID:  123,
				Version:    2,
				Compresser: 23,
				Serializer: 14,
				Error:      []byte("this is error"),
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "no error",
			resp: &Response{
				HeadLength: 120,
				RequestID:  123,
				Version:    2,
				Compresser: 23,
				Serializer: 14,
				Data:       []byte("hello, world"),
			},
		},
	}

	for _, tc := range testCases {
		tc.resp.CalculateHeaderLength()
		tc.resp.CalculateBodyLength()
		t.Run(tc.name, func(t *testing.T) {
			data := EncodeResp(tc.resp)
			req := DecodeResp(data)
			assert.Equal(t, tc.resp, req)
		})
	}

}
