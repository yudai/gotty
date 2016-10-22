package reuseport

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	serverOneResponse   = "1"
	serverTwoResponse   = "2"
	serverThreeResponse = "3"
)

var (
	serverOne   = NewServer(serverOneResponse)
	serverTwo   = NewServer(serverTwoResponse)
	serverThree = NewServer(serverThreeResponse)
)

func NewServer(resp string) *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, resp)
	}))
}

func TestNewReusablePortListener(t *testing.T) {
	listenerOne, err := NewReusablePortListener("tcp4", "localhost:10081")
	if err != nil {
		panic(err)
	}
	defer listenerOne.Close()

	listenerTwo, err := NewReusablePortListener("tcp", "127.0.0.1:10081")
	if err != nil {
		panic(err)
	}
	defer listenerTwo.Close()

	// listenerThree, err := NewReusablePortListener("tcp6", "[::1]:10081")
	// if err != nil {
	// 	panic(err)
	// }
	// defer listenerThree.Close()

	serverOne.Listener = listenerOne
	serverTwo.Listener = listenerTwo
	// serverThree.Listener = listenerThree

	serverOne.Start()
	serverTwo.Start()

	// Server One — First Response
	resp1, err := http.Get(serverOne.URL)
	if err != nil {
		panic(err)
	}
	body1, err := ioutil.ReadAll(resp1.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body1) != serverOneResponse && string(body1) != serverTwoResponse {
		t.Errorf("Expected %#v or %#v, got %#v.", serverOneResponse, serverTwoResponse, string(body1))
	}

	// Server Two — First Response
	resp2, err := http.Get(serverTwo.URL)
	if err != nil {
		panic(err)
	}
	body2, err := ioutil.ReadAll(resp2.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body2) != serverOneResponse && string(body2) != serverTwoResponse {
		t.Errorf("Expected %#v or %#v, got %#v.", serverOneResponse, serverTwoResponse, string(body2))
	}

	serverTwo.Close()

	// Server One — Second Response
	resp3, err := http.Get(serverOne.URL)
	if err != nil {
		panic(err)
	}
	body3, err := ioutil.ReadAll(resp3.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body3) != serverOneResponse {
		t.Errorf("Expected %#v, got %#v.", serverOneResponse, string(body3))
	}

	serverThree.Start()

	// Server Three — First Response
	resp4, err := http.Get(serverThree.URL)
	if err != nil {
		panic(err)
	}
	body4, err := ioutil.ReadAll(resp4.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body4) != serverThreeResponse {
		t.Errorf("Expected %#v, got %#v.", serverThreeResponse, string(body4))
	}

	serverThree.Close()

	// Server One — Third Response
	resp5, err := http.Get(serverOne.URL)
	if err != nil {
		panic(err)
	}
	body5, err := ioutil.ReadAll(resp5.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body5) != serverOneResponse {
		t.Errorf("Expected %#v, got %#v.", serverOneResponse, string(body5))
	}

	serverOne.Close()
}

func BenchmarkNewReusablePortListener(b *testing.B) {
	for i := 0; i < b.N; i++ {
		listener, err := NewReusablePortListener("tcp4", "localhost:10081")
		if err != nil {
			b.Error(err)
		}
		listener.Close()
	}
}

func ExampleNewReusablePortListener(b *testing.B) {
	listener, err := NewReusablePortListener("tcp4", ":8881")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	server := &http.Server{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
		listener.Close()
	})

	panic(server.Serve(listener))
}
