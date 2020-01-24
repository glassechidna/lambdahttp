package secretenv

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/pkg/errors"
	"strings"
)

type SM struct {
	api    secretsmanageriface.SecretsManagerAPI
	prefix string
}

func NewSM(api secretsmanageriface.SecretsManagerAPI) *SM {
	return &SM{api: api, prefix: "{aws-sm}"}
}

func (s *SM) MutateEnv(ctx context.Context, envMap map[string]string) error {
	keys := keysWithPrefixedValue(envMap, s.prefix)

	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		name, field := s.parseName(envMap[key])

		input := &secretsmanager.GetSecretValueInput{SecretId: &name}
		output, err := s.api.GetSecretValueWithContext(ctx, input)
		if err != nil {
			return errors.WithStack(err)
		}

		if field == "" {
			envMap[key] = *output.SecretString
		} else {
			envMap[key], err = extractValue(*output.SecretString, field)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SM) parseName(name string) (string, string) {
	nameAndField := strings.TrimPrefix(name, s.prefix)
	parts := strings.Split(nameAndField, "::")

	if len(parts) > 1 {
		return parts[0], parts[1]
	} else {
		return parts[0], ""
	}
}

func extractValue(secret, field string) (string, error) {
	msi := map[string]interface{}{}
	err := json.Unmarshal([]byte(secret), &msi)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if str, ok := msi[field].(string); ok {
		return str, nil
	} else {
		return "", errors.Errorf("field not a string: %s for secret", msi[field])
	}
}
