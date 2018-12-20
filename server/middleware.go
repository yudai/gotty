package server

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func (server *Server) wrapLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &logResponseWriter{w, 200}
		handler.ServeHTTP(rw, r)
		log.Printf("%s %d %s %s", r.RemoteAddr, rw.status, r.Method, r.URL.Path)
	})
}

func (server *Server) wrapHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// todo add version
		w.Header().Set("Server", "GoTTY")
		handler.ServeHTTP(w, r)
	})
}

func (server *Server) wrapBasicAuth(handler http.Handler, credential string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(token) != 2 || strings.ToLower(token[0]) != "basic" {
			w.Header().Set("WWW-Authenticate", `Basic realm="GoTTY"`)
			http.Error(w, "Bad Request", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(token[1])
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if credential != string(payload) {
			w.Header().Set("WWW-Authenticate", `Basic realm="GoTTY"`)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		log.Printf("Basic Authentication Succeeded: %s", r.RemoteAddr)
		handler.ServeHTTP(w, r)
	})
}

func (server *Server) wrapBasicAuthenticator(handler http.Handler, script string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip verification for static files
		if r.URL.Path != "/" {
			handler.ServeHTTP(w, r)
			return
		}

		// authentication failed
		failed := func() {
			w.Header().Set("WWW-Authenticate", `Basic realm="GoTTY"`)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		}

		token := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(token) != 2 || strings.ToLower(token[0]) != "basic" {
			failed()
			return
		}

		payload, err := base64.StdEncoding.DecodeString(token[1])
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// slplit username and password
		parts := strings.SplitN(string(payload), ":", 2)

		// username or password is empty
		if parts[0] == "" || parts[1] == "" {
			failed()
			return
		}

		cmd := exec.Command(script, parts...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		cmd.Stdin = &requestInfoReader{Request: r}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Start()
		if err != nil {
			log.Printf("ERROR: Authenticator start failed: %v", err)
			http.Error(w, "Internal Server Error: Authenticator start failed", http.StatusInternalServerError)
			return
		}

		// Use a channel to signal completion so we can use a select statement
		done := make(chan error)
		go func() { done <- cmd.Wait() }()

		// Start a timer
		timeout := time.After(3 * time.Second)

		// The select statement allows us to execute based on which channel
		// we get a message from first.
		select {
		case <-timeout:
			// Timeout happened first, kill the process and print a message.
			cmd.Process.Kill()
			log.Printf("ERROR: Authenticator timed out")
			http.Error(w, "Internal Server Error: Authenticator start failed", http.StatusInternalServerError)
			return
		case err := <-done:
			if err != nil {
				status := err.(*exec.ExitError).Sys().(syscall.WaitStatus)
				if status.Exited() {
					if status.ExitStatus() == 1 {
						failed()
						return
					}
				}
				log.Printf("ERROR: Authenticator failed: %v", err)
				return
			}
		}

		log.Printf("Basic Authentication Succeeded: %s user=%q", r.RemoteAddr, parts[0])
		handler.ServeHTTP(w, r)
	})
}
