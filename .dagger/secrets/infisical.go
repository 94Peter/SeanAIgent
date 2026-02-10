package secrets

import (
	"context"
	"dagger/aigent/internal/dagger"
	"fmt"

	infisical "github.com/infisical/go-sdk"
)

type Dagger interface {
}

func NewInfisical(
	ctx context.Context,
	dagger *dagger.Client,
	clientID string, clientSecret *dagger.Secret,
) (*infisicalProvider, error) {
	plainSecret, err := clientSecret.Plaintext(ctx)
	if err != nil {
		return nil, err
	}
	client := infisical.NewInfisicalClient(ctx, infisical.Config{
		SiteUrl: "https://app.infisical.com",
	})
	_, err = client.Auth().UniversalAuthLogin(clientID, plainSecret)
	if err != nil {
		return nil, fmt.Errorf("Infisical login failed: %w", err)
	}
	return &infisicalProvider{
		client: client,
		dagger: dagger,
	}, nil
}

type infisicalProvider struct {
	client infisical.InfisicalClientInterface
	dagger *dagger.Client
}

type SecretsRetriever func(key string) (*dagger.Secret, error)

func (p *infisicalProvider) GetSecretsRetriever(
	ctx context.Context, projectID string, env string, path string,
) SecretsRetriever {
	return func(key string) (*dagger.Secret, error) {
		s, err := p.client.Secrets().Retrieve(infisical.RetrieveSecretOptions{
			SecretKey:   key,
			ProjectID:   projectID,
			Environment: env,
			SecretPath:  path,
		})
		if err != nil {
			return nil, err
		}
		return p.dagger.SetSecret(s.SecretKey, s.SecretValue), nil
	}

}

func (p *infisicalProvider) GetSecrets(
	ctx context.Context, projectID string, env string, path string,
) (map[string]*dagger.Secret, error) {
	secrets, err := p.client.Secrets().List(infisical.ListSecretsOptions{
		ProjectID:   projectID,
		Environment: env,
		SecretPath:  path,
	})
	if err != nil {
		return nil, err
	}

	// 4. 將結果封裝回 Dagger Secret Map
	secretMap := make(map[string]*dagger.Secret)
	for _, s := range secrets {
		secretMap[s.SecretKey] = p.dagger.SetSecret(s.SecretKey, s.SecretValue)
	}
	return secretMap, nil
}
