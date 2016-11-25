# tests [![GoDoc](https://godoc.org/github.com/go-gem/tests?status.svg)](https://godoc.org/github.com/go-gem/tests) [![Build Status](https://travis-ci.org/go-gem/tests.svg?branch=master)](https://travis-ci.org/go-gem/tests)  [![Go Report Card](https://goreportcard.com/badge/github.com/go-gem/test)](https://goreportcard.com/report/github.com/go-gem/tests) [![Coverage Status](https://coveralls.io/repos/github/go-gem/tests/badge.svg?branch=master)](https://coveralls.io/github/go-gem/tests?branch=master)

a test package for testing you web server, only support fasthttp and similar server,
such as [Gem](https://github.com/go-gem/gem).

## Install

```
go get github.com/go-gem/tests
```


## Example

```
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
```