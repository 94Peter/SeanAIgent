// A generated module for Aigent functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/aigent/internal/dagger"
	"dagger/aigent/secrets"
)

type Aigent struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Aigent) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Aigent) GrepDir(ctx context.Context, directoryArg *dagger.Directory, pattern string) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directoryArg).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."}).
		Stdout(ctx)
}

func (m *Aigent) BuildWithGoSdk(
	ctx context.Context,
	src *dagger.Directory,
	username string,
	repoName string,
	tag string,
	dockerHubToken *dagger.Secret,
) (string, error) {
	goModCache := dag.CacheVolume("go-mod-cache")
	goBuildCache := dag.CacheVolume("go-build-cache")

	platforms := []dagger.Platform{
		"linux/amd64",
		// "linux/arm64",
	}
	var platformVariants []*dagger.Container

	for _, p := range platforms {
		// 2. 使用 Container API 模擬 Dockerfile 的過程
		// 這樣我們才能在執行 go build 時掛載快取
		builder := dag.Container(dagger.ContainerOpts{Platform: p}).
			From("golang:1.25-alpine").
			WithExec([]string{"apk", "add", "--no-cache", "tzdata"}).
			WithMountedDirectory("/src", src).
			WithWorkdir("/src").
			// 掛載快取路徑
			WithMountedCache("/go/pkg/mod", goModCache).
			WithMountedCache("/root/.cache/go-build", goBuildCache).
			WithEnvVariable("CGO_ENABLED", "0").
			WithExec([]string{"go", "build", "-ldflags", "-s -w", "-o", "bot", "./main.go"})

		// 3. 建立輕量化的運行環境 (Multi-stage build)
		// 這能讓你上傳到 GHCR 的 Image 從 90MB 變成 20MB
		runtime := dag.Container(dagger.ContainerOpts{Platform: p}).
			From("alpine:latest").
			WithDirectory("/usr/share/zoneinfo", builder.Directory("/usr/share/zoneinfo")).
			WithDirectory("/app/assets", src.Directory("assets")).
			WithDirectory("/app/locales", src.Directory("locales")).
			WithFile("/app/bot", builder.File("/src/bot")).
			WithWorkdir("/app").
			WithEntrypoint([]string{"/app/bot"})

		platformVariants = append(platformVariants, runtime)
	}

	// 4. 定義推送到 GHCR 的地址 (轉小寫處理)
	address := fmt.Sprintf("ghcr.io/%s/%s:%s",
		strings.ToLower(username),
		strings.ToLower(repoName),
		tag,
	)

	// 5. 推送所有平台變體
	return dag.Container().
		WithRegistryAuth("ghcr.io", username, dockerHubToken).
		Publish(ctx, address, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
		})
}

func (m *Aigent) BuildWithDockerfile(
	ctx context.Context,
	src *dagger.Directory, // 專案原始碼目錄
	username string, // GHCR 帳號
	repoName string, // 映像檔名稱
	tag string, // 標籤
	dockerHubToken *dagger.Secret, // GHCR Token
	// +optional
	// +default="Dockerfile"
	dockerfilePath string,
) (string, error) {
	// 1. 定義要支援的平台
	platforms := []dagger.Platform{
		"linux/amd64",
		// "linux/arm64",
	}
	// 2. 建立存放各平台 Container 的切片
	var platformVariants []*dagger.Container

	for _, p := range platforms {
		// 針對特定平台進行建置
		fmt.Printf("🛠️ 正在準備 %s 版本的建置...\n", p)

		img := src.DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: dockerfilePath,
			Platform:   p,
		})

		platformVariants = append(platformVariants, img)
	}

	// 3. 設定推送目標

	address := fmt.Sprintf("ghcr.io/%s/%s:%s", strings.ToLower(username), strings.ToLower(repoName), tag)

	// 4. 一次性推送所有平台（Dagger 會自動處理 Manifest）
	fmt.Printf("🚀 正在推送多平台映像檔到 %s...\n", address)

	imageDigest, err := dag.Container().
		WithRegistryAuth("ghcr.io", username, dockerHubToken).
		Publish(ctx, address, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
		})

	if err != nil {
		return "", fmt.Errorf("多平台推送失敗: %w", err)
	}
	fmt.Printf("✅ 多平台建置成功！\n地址: %s\nDigest: %s", address, imageDigest)
	return address, nil
}

func (m *Aigent) CD_Dev(
	ctx context.Context,
	env string,
	pulumiToken *dagger.Secret,
	infisicalClientID string,
	infisicalClientSecret *dagger.Secret,
	infisicalProjectID string,
) error {
	// 第一步：Build
	// imageAddr, err := m.BuildWithDockerfile(ctx, ...)
	// if err != nil {
	//     return err
	// }
	imageAddr := "94peter/seanaigent:03b7a6e9a84ee460233702d8da9e5666afef9571"
	// 第二步：Deploy，直接把第一步的結果傳進去
	out, err := m.deployDocker(ctx, env, imageAddr, pulumiToken, infisicalClientID, infisicalClientSecret, infisicalProjectID)
	fmt.Println(out)
	return err
}

func (m *Aigent) deployDocker(
	ctx context.Context,
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
	// 1. 根據傳入的 env 去 Infisical 抓對應環境的 Secret
	// 注意：這裡將 "prod" 改成了變數 env
	secrets, err := infisicalProvider.GetSecrets(ctx, infisicalProjectID, env, "/deploy")
	if err != nil {
		return "", err
	}

	// 2. 啟動 Pulumi 容器
	container := dag.Container().
		From("pulumi/pulumi-go:latest").
		WithSecretVariable("PULUMI_ACCESS_TOKEN", pulumiToken).
		WithMountedDirectory("/src", dag.CurrentModule().Source().Directory("deploy")).
		WithWorkdir("/src")

	// 3. 根據環境切換 Pulumi Stack
	// 如果 stack 不存在就建立它
	container = container.
		WithExec([]string{"pulumi", "stack", "select", env, "--create"}).
		WithExec([]string{"pulumi", "config", "set", "appImage", imageAddr})

	// 4. 注入該環境特有的 Secrets
	for key, val := range secrets {

		container = container.
			WithSecretVariable(key, val).
			WithExec([]string{"sh", "-c", fmt.Sprintf("pulumi config set --secret %s $%s", key, key)})
	}

	// 5. 執行部署
	return container.WithExec([]string{"pulumi", "up", "--yes"}).Stdout(ctx)
}

// func (m *Aigent) getSecretsFromInfisical(
// 	ctx context.Context,
// 	env string, // 接收傳入的環境名稱
// 	clientID string,
// 	clientSecret *dagger.Secret,
// 	projectID string,
// ) (map[string]*dagger.Secret, error) {
// 	// 1. 解開 Dagger Secret 取得明文（僅用於 SDK 認證）
// 	plainSecret, err := clientSecret.Plaintext(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. 初始化 Infisical SDK
// 	client := infisical.NewInfisicalClient(ctx, infisical.Config{
// 		SiteUrl: "https://app.infisical.com",
// 	})

// 	_, err = client.Auth().UniversalAuthLogin(clientID, plainSecret)
// 	if err != nil {
// 		return nil, fmt.Errorf("Infisical login failed: %w", err)
// 	}

// 	// 3. 獲取特定路徑下的所有 Secrets
// 	secrets, err := client.Secrets().List(infisical.ListSecretsOptions{
// 		ProjectID:   projectID,
// 		Environment: env,
// 		SecretPath:  "/deploy",
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 4. 將結果封裝回 Dagger Secret Map
// 	secretMap := make(map[string]*dagger.Secret)
// 	for _, s := range secrets {
// 		secretMap[s.SecretKey] = dag.SetSecret(s.SecretKey, s.SecretValue)
// 	}

// 	return secretMap, nil
// }
