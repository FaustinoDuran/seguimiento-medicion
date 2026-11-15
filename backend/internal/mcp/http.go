package mcp

import (
	"errors"
	"net/http"
	"net/url"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPHandler exposes the shared demo dataset without application accounts.
// TLS is terminated by the hosting proxy; publicURL identifies its HTTPS origin.
func HTTPHandler(server *sdk.Server, publicURL string) (http.Handler, error) {
	parsed, err := url.Parse(publicURL)
	if err != nil {
		return nil, errors.New("invalid MCP_PUBLIC_URL")
	}
	validOrigin := parsed.Scheme == "https" && parsed.Host != ""
	hasExtraParts := parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != ""
	validPath := parsed.Path == "" || parsed.Path == "/"
	if !validOrigin || hasExtraParts || !validPath {
		return nil, errors.New("MCP_PUBLIC_URL must be an HTTPS origin without a path")
	}
	origin := parsed.Scheme + "://" + parsed.Host
	handler := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return server },
		&sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if supplied := r.Header.Get("Origin"); supplied != "" && supplied != origin {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
		handler.ServeHTTP(w, r)
	}), nil
}
