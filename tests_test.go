// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package tests

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
)

var (
	contextType = "text/html; charset=utf-8"
	statusCode  = fasthttp.StatusBadRequest
	respBody    = fasthttp.StatusMessage(fasthttp.StatusBadRequest)
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
		test.Method = param.reqUrl
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
