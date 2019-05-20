package lambdahttp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"io"
	"net/http"
	"net/url"
)

func newRequest(inner events.ALBTargetGroupRequest) *http.Request {
	u := urlForRequest(inner)

	var body io.Reader = bytes.NewReader([]byte(inner.Body))
	if inner.IsBase64Encoded {
		body = base64.NewDecoder(base64.StdEncoding, body)
	}

	req, _ := http.NewRequest(inner.HTTPMethod, u.String(), body)

	for k, v := range inner.Headers {
		req.Header.Set(k, v)
	}

	return req
}

func urlForRequest(request events.ALBTargetGroupRequest) *url.URL {
	proto := request.Headers["x-forwarded-proto"]
	host := request.Headers["host"]
	path := request.Path

	query := url.Values{}
	for k, vs := range request.MultiValueQueryStringParameters {
		query[k] = vs
	}
	for k, v := range request.QueryStringParameters {
		query[k] = append(query[k], v)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s?%s", proto, host, path, query.Encode()))
	return u
}
