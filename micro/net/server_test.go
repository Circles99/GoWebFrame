package net

import (
	"fmt"
	"testing"
)

func Test_Server(t *testing.T) {
	go func() {
		s := NewServer("tcp", ":8081")
		err := s.Start()
		if err != nil {
			t.Log(err)
		}
	}()

	for i := 0; i < 10; i++ {
		c := &Client{"tcp", ":8081"}
		resp, err := c.Send("hello Word")
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(resp)
	}

}
