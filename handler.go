package lambdahttp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type Handler struct {
	http.Handler
	Log bool
}

func (h *Handler) ApplicationLoadBalancer(ctx context.Context, request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	if h.Log {
		log(request)
	}

	w := httptest.NewRecorder()
	r := NewHttpRequest(request)
	r = r.WithContext(ctx)
	h.ServeHTTP(w, r)

	httpResp := w.Result()
	response, err := NewLambdaResponse(httpResp)
	if h.Log {
		log(response)
	}
	return response, err
}

func (h *Handler) ApiGateway(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if h.Log {
		log(request)
	}

	w := httptest.NewRecorder()
	r := NewHttpRequest(apiGwToAlb(request))
	r = r.WithContext(ctx)
	h.ServeHTTP(w, r)

	httpResp := w.Result()
	rawBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	b64 := base64.StdEncoding.EncodeToString(rawBody)
	response := events.APIGatewayProxyResponse{
		StatusCode:        httpResp.StatusCode,
		MultiValueHeaders: httpResp.Header,
		Headers:           singleValueHeaders(httpResp.Header),
		Body:              b64,
		IsBase64Encoded:   true,
	}
	if h.Log {
		log(response)
	}
	return response, nil
}

func apiGwToAlb(r events.APIGatewayProxyRequest) events.ALBTargetGroupRequest {
	return events.ALBTargetGroupRequest{
		HTTPMethod:                      r.HTTPMethod,
		Path:                            r.Path,
		QueryStringParameters:           r.QueryStringParameters,
		MultiValueQueryStringParameters: r.MultiValueQueryStringParameters,
		Headers:                         r.Headers,
		MultiValueHeaders:               r.MultiValueHeaders,
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

func log(in interface{}) {
	by, _ := json.Marshal(in)
	fmt.Println(string(by))
}
