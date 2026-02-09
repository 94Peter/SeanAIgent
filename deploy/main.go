package main

import (
	"github.com/dirien/pulumi-vultr/sdk/v2/go/vultr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// --- 1. 讀取不同 Stack 的配置 ---
		// 這些值會從 Pulumi.prod.yaml 或 Pulumi.dev.yaml 讀取
		cfg := config.New(ctx, "")

		// 預設值設定：如果沒設定就用東京 (nrt) 和 高性能方案 (vhp)
		// 正確寫法：使用 cfg.Get，如果回傳空字串則給予預設值
		region := cfg.Get("region")
		if region == "" {
			region = "nrt" // 預設東京
		}

		plan := cfg.Get("plan")
		if plan == "" {
			plan = "vhp-1c-1gb" // 預設高效能方案
		}
		instanceLabel := cfg.Require("label")

		// --- 2. 防火牆設定 (與之前邏輯相同) ---
		fwGroup, err := vultr.NewFirewallGroup(ctx, "coolify-fw-group", &vultr.FirewallGroupArgs{
			Description: pulumi.Sprintf("Managed by Pulumi: %s", instanceLabel),
		})
		if err != nil {
			return err
		}

		rules := []struct {
			port string
			name string
		}{
			{"22", "ssh"},
			{"80", "http"},
			{"443", "https"},
			{"8000", "coolify-ui"},
		}

		for _, r := range rules {
			_, err = vultr.NewFirewallRule(ctx, "fw-rule-"+r.name, &vultr.FirewallRuleArgs{
				FirewallGroupId: fwGroup.ID(),
				Protocol:        pulumi.String("tcp"),
				IpType:          pulumi.String("v4"),
				Subnet:          pulumi.String("0.0.0.0"),
				SubnetSize:      pulumi.Int(0),
				Port:            pulumi.String(r.port),
			})
			if err != nil {
				return err
			}
		}

		// --- 3. 啟動腳本 ---
		startupScript := `#!/bin/bash
set -e
if [ ! -f /swapfile ]; then
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile append swap sw 0 0' >> /etc/fstab
fi
curl -fsSL https://get.coollabs.io/coolify/install.sh | bash
`

		script, err := vultr.NewStartupScript(ctx, "coolify-init", &vultr.StartupScriptArgs{
			Script: pulumi.String(startupScript),
		})
		if err != nil {
			return err
		}

		// --- 4. 建立 Vultr 執行個體 ---
		server, err := vultr.NewInstance(ctx, "seanaigent-server", &vultr.InstanceArgs{
			Region:          pulumi.String(region),
			Plan:            pulumi.String(plan),
			OsId:            pulumi.Int(1743), // Ubuntu 24.04
			Label:           pulumi.String(instanceLabel),
			FirewallGroupId: fwGroup.ID(),
			ScriptId:        script.ID(),
			EnableIpv6:      pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		// 輸出
		ctx.Export("Server_IP", server.MainIp)
		ctx.Export("Environment", pulumi.String(ctx.Stack()))
		return nil
	})
}
