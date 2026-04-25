package main

import "encoding/json"

type openAPI struct {
	OpenAPI string                 `json:"openapi"`
	Info    openAPIInfo            `json:"info"`
	Servers []openAPIServer        `json:"servers,omitempty"`
	Paths   map[string]openAPIPath `json:"paths"`
}

type openAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type openAPIServer struct {
	URL string `json:"url"`
}

type openAPIPath map[string]openAPIOperation

type openAPIOperation struct {
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	Tags        []string                   `json:"tags,omitempty"`
	Responses   map[string]openAPIResponse `json:"responses"`
}

type openAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]openAPIMediaType `json:"content,omitempty"`
}

type openAPIMediaType struct {
	Schema  map[string]any `json:"schema,omitempty"`
	Example any            `json:"example,omitempty"`
}

func buildOpenAPISpec(routes map[string]metricRoute) []byte {
	paths := map[string]openAPIPath{
		"/": {
			"get": {
				Summary:     "Route index",
				Description: "Returns a route overview for the bridge.",
				Tags:        []string{"system"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "Route index as JSON"},
				},
			},
		},
		"/ui": {
			"get": {
				Summary:     "Setup console",
				Description: "Shows the browser-based HomeWizard setup page with v1/v2 mode guidance and pairing controls when needed.",
				Tags:        []string{"system"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "HTML page"},
				},
			},
		},
		"/healthz": {
			"get": {
				Summary:     "Health check",
				Description: "Returns a basic bridge health payload and token readiness flag.",
				Tags:        []string{"system"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "Health JSON"},
				},
			},
		},
		"/auth/status": {
			"get": {
				Summary:     "Authentication status",
				Description: "Returns the current pairing status used by the browser pairing UI.",
				Tags:        []string{"system"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "Authentication status JSON"},
				},
			},
		},
		"/pair": {
			"post": {
				Summary:     "Start or retry pairing",
				Description: "Triggers the HomeWizard v2 pairing flow when HTTPS/v2 mode is configured. In HTTP/v1 mode, pairing is not required and this endpoint returns a friendly no-op response.",
				Tags:        []string{"system"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "Pairing succeeded or pairing not required"},
					"409": {Description: "Pairing pending button press"},
					"502": {Description: "Upstream HomeWizard error"},
				},
			},
		},
		"/status": {
			"get": {
				Summary:     "Combined status",
				Description: "Returns raw HomeWizard payloads plus resolved metric values and missing optional routes.",
				Tags:        []string{"system"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "Combined bridge status JSON"},
				},
			},
		},
		"/api/device": {
			"get": {
				Summary:     "Raw device info",
				Description: "Proxies HomeWizard device information from /api.",
				Tags:        []string{"raw"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "Device info JSON"},
				},
			},
		},
		"/api/measurement": {
			"get": {
				Summary:     "Raw measurement",
				Description: "Proxies the most recent HomeWizard measurement JSON.",
				Tags:        []string{"raw"},
				Responses: map[string]openAPIResponse{
					"200": {Description: "Measurement JSON"},
				},
			},
		},
		"/api/telegram": {
			"get": {
				Summary:     "Raw telegram",
				Description: "Returns the most recent P1 telegram as plain text.",
				Tags:        []string{"raw"},
				Responses: map[string]openAPIResponse{
					"200": {
						Description: "Telegram text",
						Content: map[string]openAPIMediaType{
							"text/plain": {Schema: map[string]any{"type": "string"}},
						},
					},
				},
			},
		},
	}

	for _, route := range routes {
		paths[route.Path] = openAPIPath{
			"get": {
				Summary:     route.Path,
				Description: route.Description,
				Tags:        []string{route.Group},
				Responses: map[string]openAPIResponse{
					"200": {
						Description: "Plain-text metric value",
						Content: map[string]openAPIMediaType{
							"text/plain": {
								Schema:  map[string]any{"type": "string"},
								Example: route.Example,
							},
						},
					},
					"404": {Description: "Metric unavailable for this meter"},
				},
			},
		}
	}

	spec := openAPI{
		OpenAPI: "3.0.3",
		Info: openAPIInfo{
			Title:       "HomeWizard Bridge API",
			Description: "Auth-free local bridge for HomeWizard P1 data, tailored for Loxone-style HTTP consumption.",
			Version:     "1.0.0",
		},
		Servers: []openAPIServer{{URL: "/"}},
		Paths:   paths,
	}

	payload, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return []byte(`{"openapi":"3.0.3","info":{"title":"HomeWizard Bridge API","version":"1.0.0"},"paths":{}}`)
	}
	return payload
}
