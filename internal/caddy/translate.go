package caddy

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/paulrose/hatch/internal/config"
)

// Translate converts a Hatch config into a full Caddy JSON configuration.
// It skips disabled projects and returns a map suitable for JSON marshaling.
// When rootCACert and rootCAKey are non-empty, a PKI app is added so Caddy
// uses the provided root CA for issuing leaf certificates.
func Translate(cfg config.Config, rootCACert, rootCAKey string) map[string]any {
	httpsRoutes := buildRoutes(cfg)
	httpRedirectRoutes := buildHTTPRedirectRoutes(cfg)
	tlsConfig := buildTLSConfig(cfg, rootCACert)

	httpsPort := fmt.Sprintf(":%d", cfg.Settings.HTTPSPort)
	httpPort := fmt.Sprintf(":%d", cfg.Settings.HTTPPort)

	apps := map[string]any{
		"http": map[string]any{
			"servers": map[string]any{
				"hatch_https": map[string]any{
					"listen":                 []string{httpsPort},
					"routes":                 httpsRoutes,
					"tls_connection_policies": []map[string]any{{}},
					"automatic_https":         map[string]any{},
				},
				"hatch_http": map[string]any{
					"listen": []string{httpPort},
					"routes": httpRedirectRoutes,
				},
			},
		},
		"tls": tlsConfig,
	}

	if rootCACert != "" {
		apps["pki"] = buildPKIConfig(rootCACert, rootCAKey)
	}

	return map[string]any{
		"admin": map[string]any{
			"listen": DefaultAdminAddr,
		},
		"apps": apps,
	}
}

// routeInfo holds metadata for sorting routes by specificity.
type routeInfo struct {
	domain  string
	service config.Service
}

// buildRoutes builds HTTPS routes for all enabled projects, sorted by specificity.
func buildRoutes(cfg config.Config) []map[string]any {
	var infos []routeInfo

	for _, proj := range cfg.Projects {
		if !proj.Enabled {
			continue
		}
		for _, svc := range proj.Services {
			domain := proj.Domain
			if svc.Subdomain != "" {
				domain = svc.Subdomain + "." + proj.Domain
			}
			infos = append(infos, routeInfo{
				domain:  domain,
				service: svc,
			})
		}
	}

	sort.SliceStable(infos, func(i, j int) bool {
		ti := routeTier(infos[i])
		tj := routeTier(infos[j])
		if ti != tj {
			return ti < tj
		}
		// Within same tier, longer paths first.
		if len(infos[i].service.Route) != len(infos[j].service.Route) {
			return len(infos[i].service.Route) > len(infos[j].service.Route)
		}
		// Alphabetical path tiebreaker.
		if infos[i].service.Route != infos[j].service.Route {
			return infos[i].service.Route < infos[j].service.Route
		}
		// Alphabetical domain tiebreaker.
		return infos[i].domain < infos[j].domain
	})

	routes := make([]map[string]any, 0, len(infos))
	for _, info := range infos {
		routes = append(routes, buildRoute(info.domain, info.service))
	}
	return routes
}

// routeTier returns a sorting priority: 0 = subdomain, 1 = path, 2 = catch-all.
func routeTier(info routeInfo) int {
	if info.service.Subdomain != "" {
		return 0
	}
	if info.service.Route != "" {
		return 1
	}
	return 2
}

// buildRoute builds a single HTTPS route with host matcher, optional path matcher,
// and reverse_proxy handler.
func buildRoute(domain string, svc config.Service) map[string]any {
	match := map[string]any{
		"host": []string{domain},
	}
	if svc.Route != "" {
		match["path"] = []string{svc.Route}
	}

	handler := buildReverseProxyHandler(svc.Proxy, svc.WebSocket)

	return map[string]any{
		"match":    []map[string]any{match},
		"handle":   []map[string]any{handler},
		"terminal": true,
	}
}

// buildReverseProxyHandler builds a reverse_proxy handler with the dial address
// extracted from proxyURL. If websocket is true, it adds flush_interval: -1 and
// Connection/Upgrade header forwarding.
func buildReverseProxyHandler(proxyURL string, websocket bool) map[string]any {
	handler := map[string]any{
		"handler":   "reverse_proxy",
		"upstreams": []map[string]any{{"dial": extractDialAddress(proxyURL)}},
	}

	if websocket {
		handler["flush_interval"] = -1
		handler["headers"] = map[string]any{
			"request": map[string]any{
				"set": map[string]any{
					"Connection": []string{"{http.request.header.Connection}"},
					"Upgrade":    []string{"{http.request.header.Upgrade}"},
				},
			},
		}
	}

	return handler
}

// buildHTTPRedirectRoutes builds HTTP→HTTPS redirect routes using a static_response
// handler with a 302 redirect for all project domains.
func buildHTTPRedirectRoutes(cfg config.Config) []map[string]any {
	domains := collectDomains(cfg)
	if len(domains) == 0 {
		return []map[string]any{}
	}

	return []map[string]any{
		{
			"match": []map[string]any{
				{"host": domains},
			},
			"handle": []map[string]any{
				{
					"handler":     "static_response",
					"status_code": "302",
					"headers": map[string]any{
						"Location": []string{"https://{http.request.host}{http.request.uri}"},
					},
				},
			},
		},
	}
}

// buildTLSConfig builds the TLS automation config with internal issuer.
// It adds *.domain wildcards only for projects that use subdomains.
// When rootCACert is non-empty, the issuer references the "hatch" CA.
func buildTLSConfig(cfg config.Config, rootCACert string) map[string]any {
	domains := collectDomains(cfg)

	issuer := map[string]any{"module": "internal"}
	if rootCACert != "" {
		issuer["ca"] = "hatch"
	}

	policies := make([]map[string]any, 0)
	if len(domains) > 0 {
		policies = append(policies, map[string]any{
			"subjects": domains,
			"issuers":  []map[string]any{issuer},
		})
	}

	return map[string]any{
		"automation": map[string]any{
			"policies": policies,
		},
	}
}

// buildPKIConfig returns the Caddy PKI app configuration that registers
// a "hatch" certificate authority backed by the given root CA files.
func buildPKIConfig(rootCACert, rootCAKey string) map[string]any {
	return map[string]any{
		"certificate_authorities": map[string]any{
			"hatch": map[string]any{
				"name": "Hatch Local CA",
				"root": map[string]any{
					"certificate": rootCACert,
					"private_key": rootCAKey,
				},
			},
		},
	}
}

// collectDomains returns all unique domains across enabled projects, sorted.
// It adds *.domain wildcards for projects that have subdomain services.
func collectDomains(cfg config.Config) []string {
	domainSet := make(map[string]bool)

	for _, proj := range cfg.Projects {
		if !proj.Enabled {
			continue
		}
		domainSet[proj.Domain] = true

		for _, svc := range proj.Services {
			if svc.Subdomain != "" {
				domainSet["*."+proj.Domain] = true
				break
			}
		}
	}

	domains := make([]string, 0, len(domainSet))
	for d := range domainSet {
		domains = append(domains, d)
	}
	sort.Strings(domains)
	return domains
}

// extractDialAddress parses a proxy URL and returns the host:port dial address.
// For URLs without an explicit port, it defaults to :80 for http and :443 for https.
func extractDialAddress(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return proxyURL
	}
	host := u.Host
	if u.Port() == "" {
		switch u.Scheme {
		case "https":
			host += ":443"
		default:
			host += ":80"
		}
	}
	return host
}
