package proxy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	runtimeMock := httpmock.NewMockTransport()
	websiteMock := httpmock.NewMockTransport()

	proxy := New(
		"http://runtime/base",
		3000,
		&http.Client{Transport: runtimeMock},
		&http.Client{Transport: websiteMock},
	)

	runtimeMock.RegisterResponder("GET", "http://runtime/base/runtime/invocation/next", func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, `
			{
				"requestContext": {
					"elb": {
						"targetGroupArn": "arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/lambda-279XGJDqGZ5rsrHC2Fjr/49e9d65c45c6791a"
					}
				},
				"httpMethod": "GET",
				"path": "/lambda",
				"queryStringParameters": {
					"query": "1234ABCD"
				},
				"headers": {
					"host": "lambda-alb-123578498.us-east-2.elb.amazonaws.com",
					"x-forwarded-proto": "https"
				},
				"body": "",
				"isBase64Encoded": false
			}
		`)

		resp.Header.Set("Lambda-Runtime-Aws-Request-Id", "reqId")
		resp.Header.Set("Lambda-Runtime-Deadline-Ms", fmt.Sprintf("%d", (time.Now().UnixNano() / 1_000_000) + 1_000))
		resp.Header.Set("Lambda-Runtime-Invoked-Function-Arn", "c")
		resp.Header.Set("Lambda-Runtime-Trace-Id", "d")
		resp.Header.Set("Lambda-Runtime-Client-Context", "e")
		resp.Header.Set("Lambda-Runtime-Cognito-Identity", "f")

		return resp, nil
	})

	responseReturned := false

	runtimeMock.RegisterResponder("POST", "http://runtime/base/runtime/invocation/reqId/response", func(req *http.Request) (*http.Response, error) {
		dump, _ := httputil.DumpRequest(req, true)
		expected := `
POST /base/runtime/invocation/reqId/response HTTP/1.1
Host: runtime

{"statusCode":200,"statusDescription":"200","headers":{},"multiValueHeaders":{},"body":"aGVsbG8gd29ybGQh","isBase64Encoded":true}
`
		expected = strings.ReplaceAll(strings.TrimSpace(expected), "\n", "\r\n")
		assert.Equal(t, expected, string(dump))

		responseReturned = true
		return &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Body:       ioutil.NopCloser(&bytes.Buffer{}),
		}, nil
	})

	websiteMock.RegisterResponder("GET", "http://127.0.0.1:3000/lambda?query=1234ABCD", httpmock.NewStringResponder(200, "hello world!"))

	err := proxy.Next(context.Background())
	assert.NoError(t, err)
	assert.True(t, responseReturned)
}
