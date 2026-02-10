package main

import (
	"context"
	"dagger/aigent/internal/dagger"
	"dagger/aigent/secrets"
	"fmt"
)

func (m *Aigent) DeployLocal(
	ctx context.Context,
	src *dagger.Directory,
	dockerSocket *dagger.Socket,
	env string, // 例如 "dev" 或 "prod"
	imageAddr string, // 新增此參數，接收來自 Build 的結果
	pulumiToken *dagger.Secret,
	infisicalClientID string,
	infisicalClientSecret *dagger.Secret,
	infisicalProjectID string,
) (string, error) {

	infisicalProvider, err := secrets.NewInfisical(ctx, dag, infisicalClientID, infisicalClientSecret)
	if err != nil {
		return "", err
	}
	// 1. 取得 Secrets (保持原樣)
	secrets, err := infisicalProvider.GetSecrets(ctx, infisicalProjectID, env, "/deploy")
	if err != nil {
		return "", err
	}

	// 2. 啟動容器並掛載本機 Docker Socket
	container := dag.Container().
		From("pulumi/pulumi-go:latest").
		// --- 新增：切換為 root 並安裝 docker ---
		WithUser("root").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "docker.io"}).
		// ------
		WithSecretVariable("PULUMI_ACCESS_TOKEN", pulumiToken).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithUnixSocket("/var/run/docker.sock", dockerSocket)

		// 3. 根據環境切換 Pulumi Stack
	// 如果 stack 不存在就建立它
	container = container.
		WithExec([]string{"pulumi", "stack", "select", env, "--create"}).
		WithExec([]string{"pulumi", "config", "set", "appImage", imageAddr}).
		WithExec([]string{"pulumi", "config", "set", "seanaigent:isRemote", "false"}).
		WithExec([]string{"pulumi", "config", "set", "seanaigent:isContainer", "true"})

	// 4. 注入該環境特有的 Secrets
	for key, val := range secrets {
		container = container.
			WithSecretVariable("PULUMI_CONFIG_"+key, val).
			WithExec([]string{"sh", "-c", fmt.Sprintf("pulumi config set --secret %s $%s", key, key)})
	}

	// 5. 執行部署
	return container.
		WithExec([]string{"pulumi", "up", "--yes"}).
		WithExec([]string{"docker", "ps"}).
		Stdout(ctx)
}
