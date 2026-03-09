package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
)

func cond(c bool, t, f string) string {
	if c {
		return t
	}
	return f
}

func getLocalizedURL(ctx context.Context, url string) string {
	locale := i18n.GetLocale(ctx)
	if locale == nil {
		return url
	}
	lang := locale.Code().String()
	if lang == "" || lang == "zh-tw" { // Default or empty
		return url
	}
	// Ensure url starts with /
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	return fmt.Sprintf("/%s%s", lang, url)
}
