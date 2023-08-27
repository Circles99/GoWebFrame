package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req:  &Request{
				//HeadLength: 120,
				//BodyLength:
				//RequestID:  123,
				//Version:    2,
				//Compresser: 23,
				//Serializer: 14,
				//ServiceName: "MY_SERVICE",
				//MethodName:  "GetById",
				//Meta: map[string]string{
				//	"track_id": "123456",
				//	"a/b":      "a",
				//},
				//Data: []byte("hello, world"),
			},
		},
	}

	for _, tc := range testCases {
		tc.req.calculateHeaderLength()
		tc.req.calculateBodyLength()
		t.Run(tc.name, func(t *testing.T) {
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})
	}

}

func (req *Request) calculateHeaderLength() {
	// 15是默认计算的 HeadLength + BodyLength + RequestID + Version + Compresser + Serializer
	// 中间的1是为了分隔符留下的
	req.HeadLength = 15 + uint32(len(req.ServiceName)) + 1 + uint32(len(req.MethodName))
}

func (req *Request) calculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}
