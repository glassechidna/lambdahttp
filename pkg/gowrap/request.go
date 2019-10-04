package gowrap

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func NewHttpRequest(inner events.ALBTargetGroupRequest) *http.Request {
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

func NewLambdaResponse(httpResp *http.Response) (events.ALBTargetGroupResponse, error) {
	rawBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return events.ALBTargetGroupResponse{}, errors.WithStack(err)
	}

	b64 := base64.StdEncoding.EncodeToString(rawBody)
	response := events.ALBTargetGroupResponse{
		StatusCode:        httpResp.StatusCode,
		StatusDescription: httpResp.Status,
		MultiValueHeaders: httpResp.Header,
		Headers:           singleValueHeaders(httpResp.Header),
		Body:              b64,
		IsBase64Encoded:   true,
	}
	return response, nil
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

func apiGwToAlb(r events.APIGatewayProxyRequest) events.ALBTargetGroupRequest {
	h := map[string]string{}
	for k, v := range r.Headers {
		h[strings.ToLower(k)] = v
	}

	mvh := map[string][]string{}
	for k, vs := range r.MultiValueHeaders {
		mvh[strings.ToLower(k)] = vs
	}

	return events.ALBTargetGroupRequest{
		HTTPMethod:                      r.HTTPMethod,
		Path:                            r.Path,
		QueryStringParameters:           r.QueryStringParameters,
		MultiValueQueryStringParameters: r.MultiValueQueryStringParameters,
		Headers:                         h,
		MultiValueHeaders:               mvh,
		IsBase64Encoded:                 r.IsBase64Encoded,
		Body:                            r.Body,
	}
}

func singleValueHeaders(h http.Header) map[string]string {
	m := map[string]string{}
	for k, vs := range h {
		m[k] = vs[0]
	}
	return m
}
