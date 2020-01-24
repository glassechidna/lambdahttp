package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/glassechidna/lambdahttp/pkg/proxy"
	"github.com/glassechidna/lambdahttp/pkg/secretenv"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	sess := session.Must(session.NewSession())
	err := secretenv.MutateEnv(sess)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	port := port()

	ctx := context.Background()
	cmdch := make(chan error)
	go runCmd(ctx, cmdch)

	readych := make(chan error)
	go waitForHealthy(ctx, readych, port)

	select {
	case err := <-cmdch:
		panic(fmt.Sprintf("%+v", err))
	case err := <-readych:
		if err != nil {
			panic(fmt.Sprintf("%+v", err))
		}
	}

	go runProxy(ctx, port, cmdch)
	for {
		select {
		case <-ctx.Done():
			panic("cancelled")
		case err := <-cmdch:
			panic(err)
		}
	}
}

func waitForHealthy(ctx context.Context, readych chan error, port int) {
	path := strings.TrimPrefix(os.Getenv("HEALTHCHECK_PATH"), "/")
	if path == "" {
		path = "ping"
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/%s", port, path)

	waitUntil(ctx, readych, func() bool {
		resp, err := http.Get(url)
		return err == nil && resp != nil && resp.StatusCode == 200
	})
}

func runProxy(ctx context.Context, port int, cmdch chan error) {
	runtimeBaseUrl := os.ExpandEnv("http://${AWS_LAMBDA_RUNTIME_API}/2018-06-01")

	proxy := proxy.New(runtimeBaseUrl, port, &http.Client{}, &http.Client{})
	for {
		err := proxy.Next(ctx)
		if err != nil {
			cmdch <- err
		}
	}
}

func port() int {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
		os.Setenv("PORT", "8080")
	}
	return port
}

func runCmd(ctx context.Context, ch chan error) {
	fmt.Println(os.Getwd())

	subcmd := os.Getenv("_HANDLER")
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", subcmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	ch <- cmd.Run()
}

func waitUntil(ctx context.Context, done chan error, condition func() bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			done <- ctx.Err()
		case <-ticker.C:
			if condition() {
				done <- nil
			}
		}
	}
}
