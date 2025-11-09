//go:build !windows

package resource

import (
	webview "github.com/webview/webview_go"
)

func setAppIcon(w webview.WebView) {}
