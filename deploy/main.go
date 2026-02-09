package main

import (
	"encoding/base64"

	"github.com/dirien/pulumi-vultr/sdk/v2/go/vultr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		// --- 1. 配置讀取 ---
		region := cfg.Get("region")
		if region == "" {
			region = "nrt"
		}
		plan := cfg.Get("plan")
		if plan == "" {
			plan = "vhp-1c-1gb"
		}
		instanceLabel := cfg.Require("label")

		fwGroup, err := vultr.NewFirewallGroup(ctx, "cloudflare-secure-fw", &vultr.FirewallGroupArgs{
			Description: pulumi.Sprintf("Cloudflare Protected for %s", instanceLabel),
		})
		if err != nil {
			return err
		}

		// --- 3. 核心優化：合併 Port 80 與 443 ---
		// 使用 "80:443" 範圍格式，一條規則同時搞定 HTTP 與 HTTPS

		vultr.NewFirewallRule(ctx, "cf-web-range", &vultr.FirewallRuleArgs{
			FirewallGroupId: fwGroup.ID(),
			Protocol:        pulumi.String("tcp"),
			IpType:          pulumi.String("v4"),
			Source:          pulumi.String("cloudflare"),
			Port:            pulumi.String("80:443"), // 合併關鍵點
		})

		// --- 4. 啟動腳本 (修正 HOME 變數問題) ---
		rawScript := `#!/bin/bash
set -e
# 設定 Swap
if [ ! -f /swapfile ]; then
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile append swap sw 0 0' >> /etc/fstab
fi

# 補上環境變數避免 Coolify 安裝崩潰
export HOME=/root
export USER=root

# 安裝 Coolify
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
`
		encodedScript := base64.StdEncoding.EncodeToString([]byte(rawScript))
		script, err := vultr.NewStartupScript(ctx, "coolify-init", &vultr.StartupScriptArgs{
			Script: pulumi.String(encodedScript),
		})
		if err != nil {
			return err
		}

		// --- 5. 建立執行個體 ---
		server, err := vultr.NewInstance(ctx, "seanaigent-server", &vultr.InstanceArgs{
			Region:          pulumi.String(region),
			Plan:            pulumi.String(plan),
			OsId:            pulumi.Int(1743),
			Label:           pulumi.String(instanceLabel),
			FirewallGroupId: fwGroup.ID(),
			ScriptId:        script.ID(),
			EnableIpv6:      pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		ctx.Export("Server_IP", server.MainIp)
		return nil
	})
}
