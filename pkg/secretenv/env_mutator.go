package secretenv

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"os"
)

type EnvMutator interface {
	MutateEnv(ctx context.Context, envMap map[string]string) error
}

func MutateEnv(sess *session.Session) error {
	environ := os.Environ()
	envmap := EnvMap(environ)

	mutators := []EnvMutator{
		NewSSM(ssm.New(sess)),
		NewSM(secretsmanager.New(sess)),
	}

	for _, mutator := range mutators {
		err := mutator.MutateEnv(context.Background(), envmap)
		if err != nil {
			return err
		}
	}

	// TODO: doesn't delete keys, can only add or replace
	for k, v := range envmap {
		os.Setenv(k, v)
	}

	return nil
}
