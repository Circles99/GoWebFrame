package round_robin

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"testing"
)

func TestBuilder_Pick(t *testing.T) {
	var testCases = []struct {
		name              string
		b                 *Balancer
		wantErr           error
		wantSubConn       SubConn
		wantBalancerIndex int32
	}{
		{
			name: "start",
			b: &Balancer{
				index: -1,
				connections: []balancer.SubConn{
					SubConn{name: "127.0.0.1：8080"},
					SubConn{name: "127.0.0.1：8081"},
				},
				length: 2,
			},
			wantErr:           nil,
			wantSubConn:       SubConn{name: "127.0.0.1：8080"},
			wantBalancerIndex: 0,
		},
		{
			name: "end",
			b: &Balancer{
				index: 1,
				connections: []balancer.SubConn{
					SubConn{name: "127.0.0.1：8080"},
					SubConn{name: "127.0.0.1：8081"},
				},
				length: 2,
			},
			wantErr:           nil,
			wantSubConn:       SubConn{name: "127.0.0.1：8080"},
			wantBalancerIndex: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.b.Pick(balancer.PickInfo{})
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantSubConn.name, res.SubConn.(SubConn).name)
			assert.NotNil(t, res.Done)
			assert.Equal(t, tc.wantBalancerIndex, tc.b.index)
		})
	}
}

type SubConn struct {
	name string
	balancer.SubConn
}
