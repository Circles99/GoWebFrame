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
			req: &Request{
				HeadLength: 120,
				//BodyLength:
				RequestID:   123,
				Version:     2,
				Compresser:  23,
				Serializer:  14,
				ServiceName: "MY_SERVICE",
				MethodName:  "GetById",
				Meta: map[string]string{
					"track_id": "123456",
					"a/b":      "a",
				},
				//Data: []byte("hello, world"),
			},
		},
	}

	for _, tc := range testCases {
		tc.req.CalculateHeaderLength()
		tc.req.CalculateBodyLength()
		t.Run(tc.name, func(t *testing.T) {
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})
	}

}
