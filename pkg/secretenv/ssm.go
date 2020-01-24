package secretenv

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pkg/errors"
	"strings"
)

type SSM struct {
	api    ssmiface.SSMAPI
	prefix string
}

func NewSSM(api ssmiface.SSMAPI) *SSM {
	return &SSM{api: api, prefix: "{aws-ssm}"}
}

func (s *SSM) MutateEnv(ctx context.Context, envMap map[string]string) error {
	keys := keysWithPrefixedValue(envMap, s.prefix)
	parameterNames := []string{}
	nameMap := map[string]string{}

	for _, key := range keys {
		name := strings.TrimPrefix(envMap[key], s.prefix)
		parameterNames = append(parameterNames, name)
		nameMap[name] = key
	}

	if len(keys) == 0 {
		return nil
	}

	input := &ssm.GetParametersInput{Names: aws.StringSlice(parameterNames), WithDecryption: aws.Bool(true)}
	output, err := s.api.GetParametersWithContext(ctx, input)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, parameter := range output.Parameters {
		envName := nameMap[*parameter.Name]
		envMap[envName] = *parameter.Value
		parameterNames = remove(parameterNames, *parameter.Name)
	}

	if len(parameterNames) > 0 {
		return errors.Errorf("parameters missing from parameter store: %+v", parameterNames)
	}

	return nil
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
