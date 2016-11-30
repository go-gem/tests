// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	contextType = "text/html; charset=utf-8"
	statusCode  = fasthttp.StatusBadRequest
	respBody    = fasthttp.StatusMessage(fasthttp.StatusBadRequest)

	// header
	headerKey   = "Custom-Header"
	headerValue = "tests"

	// cookie
	cookie      = &fasthttp.Cookie{}
	cookieKey   = "GOSESSION"
	cookieValue = "GOSESSION_VALUE"

	// Fake server.
	srv = &fasthttp.Server{}

	testParams = make([]param, 0)
)

type param struct {
	expectStatus  int
	expectBody    string
	expectHeaders map[string]string
	expectErr     bool
	expectCustoms []Func

	reqUrl      string
	reqMethod   string
	reqProtocol string
	reqHeaders  map[string]string
	reqPayload  string
}

func init() {
	cookie.SetKey(cookieKey)
	cookie.SetValue(cookieValue)

	srv.Handler = func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Path()) == "/timeout" {
			time.Sleep(time.Millisecond * 200)
		}

		if len(ctx.Request.Header.Peek(headerKey)) > 0 {
			ctx.Response.Header.SetBytesV(headerKey, ctx.Request.Header.Peek(headerKey))
		}

		ctx.SetStatusCode(statusCode)
		ctx.SetContentType(contextType)
		ctx.SetBodyString(respBody)

		ctx.Response.Header.SetCookie(cookie)
	}

	// Correct status.
	testParams = append(testParams, param{
		expectStatus: statusCode,
	})
	// Incorrect status.
	testParams = append(testParams, param{
		expectErr:    true,
		expectStatus: fasthttp.StatusGatewayTimeout,
	})

	// Correct Content-Type.
	testParams = append(testParams, param{
		expectHeaders: map[string]string{
			"Content-Type": contextType,
		},
	})
	// Incorrect Content-Type.
	testParams = append(testParams, param{
		expectErr: true,
		expectHeaders: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		},
	})

	// Correct response body.
	testParams = append(testParams, param{
		expectBody: respBody,
	})
	// Incorrect response body.
	testParams = append(testParams, param{
		expectErr:  true,
		expectBody: "Incorrect response body",
	})

	// Add custom check function to validate cookie.
	testParams = append(testParams, param{
		expectCustoms: []Func{
			func(resp fasthttp.Response) error {
				cookie := &fasthttp.Cookie{}
				cookie.SetKey(cookieKey)
				if !resp.Header.Cookie(cookie) {
					return fmt.Errorf("failed to get cookie")
				}
				if string(cookie.Value()) != cookieValue {
					return fmt.Errorf("Expect cookie named %s: %q, got %q", cookieKey, cookieValue, cookie.Value())
				}
				return nil
			},
		},
	})

	// Test custom request header
	testParams = append(testParams, param{
		reqHeaders: map[string]string{
			headerKey: headerValue,
		},
		expectCustoms: []Func{
			func(resp fasthttp.Response) error {
				bytesHeader := resp.Header.Peek(headerKey)
				if len(bytesHeader) == 0 || string(bytesHeader) != headerValue {
					return fmt.Errorf("Expect header named %s: %q, got %q", headerKey, headerValue, bytesHeader)
				}
				return nil
			},
		},
	})

	// Test timeout
	testParams = append(testParams, param{
		reqUrl:    "/timeout",
		expectErr: true,
	})
}

func TestNew(t *testing.T) {
	var s server

	url := "/user"
	method := "POST"
	protocol := "HTTP/1.0"

	var err error

	test1 := New(s)
	if err = check(test1, defaultUrl, defaultMethod, defaultProtocol); err != nil {
		t.Error(err)
	}

	test2 := New(s, url)
	if err = check(test2, url, defaultMethod, defaultProtocol); err != nil {
		t.Error(err)
	}

	test3 := New(s, url, method)
	if err = check(test3, url, method, defaultProtocol); err != nil {
		t.Error(err)
	}

	test4 := New(s, url, method, protocol)
	if err = check(test4, url, method, protocol); err != nil {
		t.Error(err)
	}
}

func check(t *Test, url, method, protocol string) error {
	if t.Url != url {
		return fmt.Errorf("expected url: %q, got %q", url, t.Url)
	}
	if t.Method != method {
		return fmt.Errorf("expected method: %q, got %q", method, t.Method)
	}
	if t.Protocol != protocol {
		return fmt.Errorf("expected protocol: %q, got %q", protocol, t.Protocol)
	}

	return nil
}

func TestAll(t *testing.T) {
	var err error
	for _, param := range testParams {
		test := New(srv)
		initTest(test, &param)

		err = test.Run()

		if param.expectErr && err == nil {
			t.Error("Expect non-nil error, but got nil error")
			t.Errorf("%+v\n", param)
		} else if !param.expectErr && err != nil {
			t.Error(err)
		}
	}
}

func initTest(test *Test, param *param) {
	// Request
	if param.reqMethod != "" {
		test.Method = param.reqMethod
	}
	if param.reqUrl != "" {
		test.Url = param.reqUrl
	}
	if param.reqProtocol != "" {
		test.Protocol = param.reqProtocol
	}
	if param.reqPayload != "" {
		test.Payload = param.reqPayload
	}
	if len(param.reqHeaders) > 0 {
		test.Headers = param.reqHeaders
	}

	// Expected result
	if param.expectStatus > 0 {
		test.Expect().Status(param.expectStatus)
	}
	if param.expectBody != "" {
		test.Expect().Body(param.expectBody)
	}
	if len(param.expectHeaders) > 0 {
		for k, v := range param.expectHeaders {
			test.Expect().Header(k, v)
		}
	}
	if len(param.expectCustoms) > 0 {
		for _, f := range param.expectCustoms {
			test.Expect().Custom(f)
		}
	}
}

func TestExpect_Rest(t *testing.T) {
	e := new(Expect)
	e.Status(fasthttp.StatusOK)

	e.Rest()
	if len(*e) != 0 {
		t.Error("failed to reset Expect")
	}
}
