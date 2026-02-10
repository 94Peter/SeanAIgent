package main

import (
	"context"
	"dagger/aigent/internal/dagger"
	"dagger/aigent/secrets"
	"fmt"
	"time"
)

// func (m *Aigent) DeployRemote(
// 	ctx context.Context,
// 	src *dagger.Directory,
// 	env string,
// 	imageAddr string,
// 	pulumiToken *dagger.Secret,
// 	infisicalClientID string,
// 	infisicalClientSecret *dagger.Secret,
// 	infisicalProjectID string,
// ) (string, error) {

// 	// ---------------------------------------------------------------------
// 	// 1. Load secrets from Infisical
// 	// ---------------------------------------------------------------------
// 	infisicalProvider, err := secrets.NewInfisical(
// 		ctx, dag, infisicalClientID, infisicalClientSecret,
// 	)
// 	if err != nil {
// 		return "", err
// 	}

// 	deploySecrets, err := infisicalProvider.GetSecrets(
// 		ctx, infisicalProjectID, env, "/deploy",
// 	)
// 	if err != nil {
// 		return "", err
// 	}

// 	getSecret := infisicalProvider.GetSecretsRetriever(
// 		ctx, infisicalProjectID, env, "/",
// 	)

// 	cfID, err := getSecret("CF_CLIENT_ID")
// 	if err != nil {
// 		return "", err
// 	}
// 	cfSecret, err := getSecret("CF_CLIENT_SECRET")
// 	if err != nil {
// 		return "", err
// 	}
// 	sshKey, err := getSecret("SSH_PRIVATE_KEY")
// 	if err != nil {
// 		return "", err
// 	}
// 	sshUser, err := getSecret("SSH_USER")
// 	if err != nil {
// 		return "", err
// 	}

// 	// ---------------------------------------------------------------------
// 	// 2. Base container
// 	// ---------------------------------------------------------------------
// 	container := dag.Container().
// 		From("pulumi/pulumi:latest").
// 		WithMountedDirectory("/src", src).
// 		WithWorkdir("/src").
// 		WithMountedSecret("/tmp/id_rsa_ro", sshKey). // immutable
// 		WithSecretVariable("CF_ID", cfID).
// 		WithSecretVariable("CF_SECRET", cfSecret).
// 		WithSecretVariable("SSH_USER", sshUser).
// 		WithSecretVariable("PULUMI_ACCESS_TOKEN", pulumiToken).
// 		WithExec([]string{"sh", "-c", `
//         apt-get update && apt-get install -y netcat-openbsd curl
//         curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o /usr/local/bin/cloudflared
//         chmod +x /usr/local/bin/cloudflared
//     `}).
// 		WithExec([]string{"sh", "-c", `
// set -e

// mkdir -p /root/.ssh
// cp /tmp/id_rsa_ro /root/.ssh/id_rsa
// chmod 700 /root/.ssh
// chmod 600 /root/.ssh/id_rsa
// `}).
// 		WithExec([]string{"sh", "-c", `
// cat > /usr/local/bin/ssh <<'EOF'
// #!/bin/sh
// # 如果參數中包含 "docker"，就把它替換成 Mac 上的絕對路徑
// cmd=""
// for arg in "$@"; do
//     if [ "$arg" = "docker" ]; then
//         cmd="$cmd /usr/local/bin/docker"
//     else
//         cmd="$cmd \"$arg\""
//     fi
// done
// # 呼叫真正的 ssh (通常在 /usr/bin/ssh)
// exec /usr/bin/ssh $cmd
// EOF
// chmod +x /usr/local/bin/ssh
// `}).
// 		WithEnvVariable("PATH", "/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/homebrew/opt/go@1.24/bin")

// 	script := `
// set -e

// # 1. 隧道與配置 (維持不變)
// mkdir -p /root/.ssh
// cat > /root/.ssh/config <<EOF
// Host localhost
//     StrictHostKeyChecking no
//     UserKnownHostsFile /dev/null
//     IdentityFile /root/.ssh/id_rsa
//     Port 2222
// EOF
// chmod 600 /root/.ssh/config

// cloudflared access tcp --hostname ssh.94peter.dev --url localhost:2222 --service-token-id "$CF_ID" --service-token-secret "$CF_SECRET" >/tmp/cf.log 2>&1 &

// for i in $(seq 1 10); do
//   if nc -z localhost 2222; then break; fi
//   sleep 1
// done

// # 2. 設定 DOCKER_HOST 與 DOCKER_SSH_COMMAND
// export DOCKER_HOST=ssh://$SSH_USER@localhost

// # 核心修正：讓 Docker SDK 透過我們的 Wrapper 啟動 SSH
// export DOCKER_SSH_COMMAND="/usr/local/bin/ssh-wrapper"

// # 執行部署
// pulumi up --yes
// `

// 	// ---------------------------------------------------------------------
// 	// 5. Pulumi config
// 	// ---------------------------------------------------------------------
// 	container = container.
// 		WithExec([]string{"pulumi", "stack", "select", env, "--create"}).
// 		WithExec([]string{"pulumi", "config", "set", "appImage", imageAddr}).
// 		WithExec([]string{"pulumi", "config", "set", "seanaigent:isRemote", "true"}).
// 		WithExec([]string{"pulumi", "config", "set", "seanaigent:isContainer", "true"})

// 	for key, val := range deploySecrets {
// 		container = container.
// 			WithSecretVariable(key, val).
// 			WithExec([]string{
// 				"sh", "-c",
// 				fmt.Sprintf("pulumi config set --secret %s $%s", key, key),
// 			})
// 	}

// 	// ---------------------------------------------------------------------
// 	// 6. Pulumi up (remote Docker via SSH)
// 	// ---------------------------------------------------------------------
// 	return container.
// 		WithExec([]string{"sh", "-c", script}).
// 		Stdout(ctx)
// }

func (m *Aigent) DeployRemote(
	ctx context.Context,
	src *dagger.Directory,
	env string,
	imageAddr string,
	pulumiToken *dagger.Secret,
	infisicalClientID string,
	infisicalClientSecret *dagger.Secret,
	infisicalProjectID string,
) (string, error) {

	// ---------------------------------------------------------------------
	// 1. Load secrets from Infisical (維持不變)
	// ---------------------------------------------------------------------
	infisicalProvider, err := secrets.NewInfisical(ctx, dag, infisicalClientID, infisicalClientSecret)
	if err != nil {
		return "", err
	}
	deploySecrets, err := infisicalProvider.GetSecrets(ctx, infisicalProjectID, env, "/deploy")
	if err != nil {
		return "", err
	}
	getSecret := infisicalProvider.GetSecretsRetriever(ctx, infisicalProjectID, env, "/")

	// 這裡拿到的就是 Cloudflare Access 的身分證
	cfID, _ := getSecret("CF_CLIENT_ID")
	cfSecret, _ := getSecret("CF_CLIENT_SECRET")

	// ---------------------------------------------------------------------
	// 2. Base container
	// ---------------------------------------------------------------------
	container := dag.Container().
		From("pulumi/pulumi:latest").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		// 將 Cloudflare Token 注入環境變數，這會被你的 Pulumi Go 程式讀取
		WithSecretVariable("CF_ACCESS_CLIENT_ID", cfID).
		WithSecretVariable("CF_ACCESS_CLIENT_SECRET", cfSecret).
		WithSecretVariable("PULUMI_ACCESS_TOKEN", pulumiToken).
		WithEnvVariable("CACHE_BUST", time.Now().String())

	// ---------------------------------------------------------------------
	// 3. Pulumi config
	// ---------------------------------------------------------------------
	container = container.
		WithExec([]string{"pulumi", "stack", "select", env, "--create"}).
		WithExec([]string{"pulumi", "config", "set", "appImage", imageAddr})
		// 因為現在改跑 Nomad，原本的 isContainer 等邏輯可能要視你的 Pulumi 代碼而定

	// 注入其他業務用的機密環境變數
	for key, val := range deploySecrets {
		container = container.
			WithSecretVariable(key, val).
			WithExec([]string{"sh", "-c", fmt.Sprintf("pulumi config set --secret %s $%s", key, key)})
	}

	// ---------------------------------------------------------------------
	// 4. Run Pulumi Up
	// ---------------------------------------------------------------------
	// 不需要啟動 cloudflared 隧道，因為 Pulumi Provider 會直接發起 HTTPS 請求
	return container.
		WithExec([]string{"pulumi", "up", "--yes"}).
		Stdout(ctx)
}
