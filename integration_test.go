package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	cases := []struct{ question, reply string }{
		{"ping", "pong"},
		{"hi", "holla"},
	}

	for _, x := range cases {
		err := errors.New("")
		resp := new(http.Response)

		for l := 0; l < 10; l++ {
			resp, err = http.Get("http://" + os.Getenv("SERVER") + ":8080/" + x.question)
			if err == nil {
				break
			}
			fmt.Println("No connection to the server , retring in 2sec")
			time.Sleep(2000 * time.Millisecond)
		}
		if err != nil {
			t.Error(err)
		}
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			t.Error(err)
		} else if string(body) != x.reply {
			t.Error("oops we have a problem: expected reply - ", x.reply, ", but got - ", string(body))
		}
	}
}
