// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

/*
Package tests provides simple APIs to test you web server,
only support servers which based fasthttp.

Example

	import (
		"testing"

		"github.com/go-gem/tests"
		"github.com/valyala/fasthttp"
	)

	func TestFastHTTP(t *testing.T) {
		contentType := "text/html; charset=utf-8"
		statusCode := fasthttp.StatusBadRequest
		respBody := fasthttp.StatusMessage(fasthttp.StatusBadRequest)

		// Fake server
		srv := &fasthttp.Server{
			Handler: func(ctx *fasthttp.RequestCtx) {
				ctx.SetContentType(contentType)
				ctx.SetStatusCode(statusCode)
				ctx.SetBodyString(respBody)
			},
		}

		// Create a Test instance.
		test := tests.New(srv)

		// Customize request.
		// See Test struct.
		test.Url = "/"

		// Add excepted result.
		test.Expect().
			Status(statusCode).
			Header("Content-Type", contentType).
			Body(respBody)

		// Custom checking function.
		test.Expect().Custom(func(resp fasthttp.Response) error {
			// check response.

			return nil
		})

		// Run test.
		if err := test.Run(); err != nil {
			t.Error(err)
		}
	}
*/
package tests

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	line = "\r\n"
)

// Func customize function to check response.
type Func func(resp fasthttp.Response) error

// Expect a slice of checking functions.
type Expect []Func

// Custom add custom function.
func (e *Expect) Custom(f Func) *Expect {
	*e = append(*e, f)
	return e
}

func (e *Expect) check(resp fasthttp.Response) (err error) {
	for _, f := range *e {
		if err = f(resp); err != nil {
			return
		}
	}

	return
}

// Status add expected status code.
func (e *Expect) Status(code int) *Expect {
	return e.Custom(func(resp fasthttp.Response) error {
		if resp.StatusCode() != code {
			return fmt.Errorf("expected status code %d, got %d", code, resp.StatusCode())
		}

		return nil
	})
}

// Body add expected response body.
func (e *Expect) Body(body string) *Expect {
	return e.Custom(func(resp fasthttp.Response) error {
		if string(resp.Body()) != body {
			return fmt.Errorf("expected response body %q, got %q", body, resp.Body())
		}

		return nil
	})
}

// Header add expected header.
func (e *Expect) Header(key, value string) *Expect {
	return e.Custom(func(resp fasthttp.Response) error {
		v := resp.Header.Peek(key)
		if string(v) != value {
			return fmt.Errorf("expected response header named %s: %q, got %q", key, value, v)
		}

		return nil
	})
}

// Rest reset expected result.
func (e *Expect) Rest() *Expect {
	*e = (*e)[:0]
	return e
}

type server interface {
	ServeConn(net.Conn) error
}

// Test struct
type Test struct {
	server  server
	Timeout time.Duration
	rw      *readWriter

	// Request configuration
	Url      string
	Method   string
	Protocol string
	Headers  map[string]string
	Payload  string

	expect *Expect
}

var (
	defaultMethod   = "GET"
	defaultUrl      = "/"
	defaultProtocol = "HTTP/1.1"

	// DefaultTimeout
	DefaultTimeout = 200 * time.Microsecond
)

// New returns a Test instance with default configuration.
func New(server server, args ...string) *Test {
	t := &Test{
		server:  server,
		rw:      &readWriter{},
		Timeout: DefaultTimeout,

		Url:      defaultUrl,
		Method:   defaultMethod,
		Protocol: defaultProtocol,
		Headers:  make(map[string]string),

		expect: new(Expect),
	}

	argsCount := len(args)
	switch argsCount {
	case 3:
		t.Protocol = args[2]
		fallthrough
	case 2:
		t.Method = args[1]
		fallthrough
	case 1:
		t.Url = args[0]
	}

	return t
}

var (
	errTimeout = errors.New("timeout")
)

// Run run test and return an error,
// return nil if everything is ok.
func (t *Test) Run() (err error) {
	t.initRW()

	br := bufio.NewReader(&t.rw.w)
	var resp fasthttp.Response
	ch := make(chan error)
	go func() {
		ch <- t.server.ServeConn(t.rw)
	}()

	select {
	case err = <-ch:
		if err != nil {
			return
		}
	case <-time.After(t.Timeout):
		return errTimeout
	}

	if err = resp.Read(br); err != nil {
		return fmt.Errorf("unexpected error when reading response: %s", err)
	}
	if err = t.expect.check(resp); err != nil {
		return
	}

	return
}

func (t *Test) initRW() {
	firstPart := fmt.Sprintf("%s %s %s", t.Method, t.Url, t.Protocol)

	secondPart := ""
	for k, v := range t.Headers {
		secondPart += fmt.Sprintf("%s: %s", k, v) + line
	}

	req := firstPart + line +
		secondPart + line +
		t.Payload

	t.rw.r.WriteString(req)
}

// Expect return expect object.
func (t *Test) Expect() *Expect {
	return t.expect
}

type readWriter struct {
	net.Conn
	r bytes.Buffer
	w bytes.Buffer
}

func (rw *readWriter) Close() error {
	return nil
}

// Read
func (rw *readWriter) Read(b []byte) (int, error) {
	return rw.r.Read(b)
}

// Write
func (rw *readWriter) Write(b []byte) (int, error) {
	return rw.w.Write(b)
}
