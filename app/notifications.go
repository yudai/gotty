package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	EventTypeServerStart   = "server_start"
	EventTypeClientConnect = "client_connect"
	EventTypeCommandStart  = "command_start"
	EventTypeClientClose   = "client_close"
	EventTypeServerClose   = "server_close"
)

type Event struct {
	Type      string   `json:"type"`
	Command   string   `json:"command,omitempty"`
	Urls      []string `json:"urls,omitempty"`
	ClientUrl string   `json:"client_url,omitempty"`
	UserAgent string   `json:"user_agent,omitempty"`
	Args      []string `json:"args,omitempty"`
	PID       int      `json:"pid,omitempty"`
}

type httpSink struct {
	url    string
	mu     sync.Mutex
	client *http.Client
}

type Notifier interface {
	Write(event Event) error
}

func NewHTTPSink(u string, timeout time.Duration) *httpSink {
	return &httpSink{
		url: u,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (hs *httpSink) Write(event Event) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	p, err := json.MarshalIndent(&event, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling event: %v", err)
	}

	body := bytes.NewReader(p)
	resp, err := hs.client.Post(hs.url, "application/json", body)
	if err != nil {
		return fmt.Errorf("error posting: %v", err)
	}

	defer resp.Body.Close()

	// We will treat any 2xx response as accepted by the endpoint.
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	default:
		return fmt.Errorf("response status %v unaccepted", resp.Status)
	}
}
