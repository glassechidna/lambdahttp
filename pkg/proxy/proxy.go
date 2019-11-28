package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/glassechidna/lambdahttp/pkg/gowrap"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Proxy struct {
	runtimeBaseUrl string
	localPort      int
	runtimeClient  *http.Client
	webClient      *http.Client
}

func New(runtimeBaseUrl string, localPort int, runtimeClient *http.Client, webClient *http.Client) *Proxy {
	return &Proxy{
		runtimeBaseUrl: runtimeBaseUrl,
		localPort:      localPort,
		runtimeClient:  runtimeClient,
		webClient:      webClient,
	}
}

func (p *Proxy) Next(ctx context.Context) error {
	albReq, requestId, err := p.getNextRequest(ctx)
	if err != nil {
		return err
	}

	unixMillis, _ := strconv.ParseInt(albReq.Headers["Lambda-Runtime-Deadline-Ms"], 10, 64)
	deadline := time.Unix(0, unixMillis * 1_000_000)
	ctx, _ = context.WithDeadline(ctx, deadline)

	childReq := gowrap.NewHttpRequest(*albReq).WithContext(ctx)
	childReq.URL.Host = fmt.Sprintf("127.0.0.1:%d", p.localPort)
	childReq.URL.Scheme = "http"

	childResponse, err := p.webClient.Do(childReq)
	if err != nil {
		p.postLambdaError(ctx, err, requestId)
		return err
	}

	err = p.postLambdaResponse(ctx, childResponse, requestId)
	if err != nil {
		return err
	}

	return nil
}

func (p *Proxy) postLambdaError(ctx context.Context, err error, requestId string) {
	respUrl := fmt.Sprintf("%s/runtime/invocation/%s/error", p.runtimeBaseUrl, requestId)
	payload := strings.NewReader(fmt.Sprintf("%+v", err))

	req, _ := http.NewRequestWithContext(ctx, "POST", respUrl, payload)
	_, _ = p.runtimeClient.Do(req)
}

func (p *Proxy) postLambdaResponse(ctx context.Context, childResponse *http.Response, requestId string) error {
	respUrl := fmt.Sprintf("%s/runtime/invocation/%s/response", p.runtimeBaseUrl, requestId)

	albResp, err := gowrap.NewLambdaResponse(childResponse)
	if err != nil {
		return errors.WithStack(err)
	}

	responseJson, err := json.Marshal(albResp)
	if err != nil {
		return errors.WithStack(err)
	}

	runtimeReq, err := http.NewRequestWithContext(ctx, "POST", respUrl, bytes.NewReader(responseJson))
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = p.runtimeClient.Do(runtimeReq)
	return errors.WithStack(err)
}

func (p *Proxy) getNextRequest(ctx context.Context) (*events.ALBTargetGroupRequest, string, error) {
	nextUrl := fmt.Sprintf("%s/runtime/invocation/next", p.runtimeBaseUrl)

	runtimeReq, err := http.NewRequestWithContext(ctx, "GET", nextUrl, nil)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	runtimeResponse, err := p.runtimeClient.Do(runtimeReq)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	payload, err := ioutil.ReadAll(runtimeResponse.Body)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	albReq := events.ALBTargetGroupRequest{}
	err = json.Unmarshal(payload, &albReq)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	for _, name := range LambdaHeaders {
		val := runtimeResponse.Header.Get(name)
		if albReq.Headers != nil {
			albReq.Headers[name] = val
		}
		if albReq.MultiValueHeaders != nil {
			albReq.MultiValueHeaders[name] = append(albReq.MultiValueHeaders[name], val)
		}
	}

	requestId := runtimeResponse.Header.Get("Lambda-Runtime-Aws-Request-Id")
	return &albReq, requestId, nil
}

var LambdaHeaders = []string{
	"Lambda-Runtime-Aws-Request-Id",
	"Lambda-Runtime-Deadline-Ms",
	"Lambda-Runtime-Invoked-Function-Arn",
	"Lambda-Runtime-Trace-Id",
	"Lambda-Runtime-Client-Context",
	"Lambda-Runtime-Cognito-Identity",
}
