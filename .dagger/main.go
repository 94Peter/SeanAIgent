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
	"dagger/aigent/internal/dagger"
	"fmt"
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

func (m *Aigent) BuildWithDockerfile(
	ctx context.Context,
	src *dagger.Directory, // 專案原始碼目錄
	username string, // Docker Hub 帳號
	repoName string, // 映像檔名稱
	tag string, // 標籤
	dockerHubToken *dagger.Secret, // Docker Hub Token
	// +optional
	// +default="Dockerfile"
	dockerfilePath string,
) (string, error) {
	// 1. 定義要支援的平台
	platforms := []dagger.Platform{
		"linux/amd64",
		"linux/arm64",
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
	address := fmt.Sprintf("docker.io/%s/%s:%s", username, repoName, tag)

	// 4. 一次性推送所有平台（Dagger 會自動處理 Manifest）
	fmt.Printf("🚀 正在推送多平台映像檔到 %s...\n", address)

	imageDigest, err := dag.Container().
		WithRegistryAuth("docker.io", username, dockerHubToken).
		Publish(ctx, address, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
		})

	if err != nil {
		return "", fmt.Errorf("多平台推送失敗: %w", err)
	}

	return fmt.Sprintf("✅ 多平台建置成功！\n地址: %s\nDigest: %s", address, imageDigest), nil
}
