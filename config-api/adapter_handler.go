package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/beedevz/hive-api/adapters"
	"gopkg.in/yaml.v3"
)

// ─── Config structs ───────────────────────────────────────────────

type serviceItem struct {
	Name          string                 `yaml:"name" json:"name"`
	URL           string                 `yaml:"url"  json:"url"`
	Adapter       string                 `yaml:"adapter"        json:"adapter"`
	AdapterConfig map[string]interface{} `yaml:"adapter_config" json:"adapter_config"`
}

type serviceCategory struct {
	Category string        `yaml:"category"`
	Items    []serviceItem `yaml:"items"`
}

// ─── Config helpers ───────────────────────────────────────────────

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func unmarshalConfig(format string, data []byte, dest interface{}) error {
	if format == "yaml" {
		return yaml.Unmarshal(data, dest)
	}
	return json.Unmarshal(data, dest)
}

// serviceSectionKeys returns the top-level config keys that hold service
// categories. It reads the "sections" metadata (type: "services") and falls
// back to the legacy "services" key when the metadata is absent.
func serviceSectionKeys(raw map[string]interface{}) []string {
	sections, ok := raw["sections"]
	if !ok {
		return []string{"services"}
	}
	list, ok := sections.([]interface{})
	if !ok {
		return []string{"services"}
	}
	var keys []string
	for _, s := range list {
		m, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		if m["type"] == "services" {
			if key, ok := m["key"].(string); ok && key != "" {
				keys = append(keys, key)
			}
		}
	}
	if len(keys) == 0 {
		return []string{"services"}
	}
	return keys
}

// loadAllServiceItems parses the config and returns every service item across
// all service sections (supports dynamic section keys like "test", "home", …).
func loadAllServiceItems(data []byte, format string) ([]serviceItem, error) {
	var raw map[string]interface{}
	if err := unmarshalConfig(format, data, &raw); err != nil {
		return nil, err
	}
	var items []serviceItem
	for _, key := range serviceSectionKeys(raw) {
		val, ok := raw[key]
		if !ok {
			continue
		}
		// Re-marshal the section value so we can unmarshal into typed structs.
		b, err := yaml.Marshal(val)
		if err != nil {
			continue
		}
		var cats []serviceCategory
		if err := yaml.Unmarshal(b, &cats); err != nil {
			continue
		}
		for _, cat := range cats {
			items = append(items, cat.Items...)
		}
	}
	return items, nil
}

// findServiceByName loads config and returns the first service whose name
// matches case-insensitively. Used by /probe/{name}/* endpoints.
func findServiceByName(name string) (*serviceItem, error) {
	format := detectFormat()
	data, err := readFile(configPath(format))
	if err != nil {
		return nil, err
	}
	items, err := loadAllServiceItems(data, format)
	if err != nil {
		return nil, err
	}
	nameLower := strings.ToLower(name)
	for i, svc := range items {
		if strings.ToLower(svc.Name) == nameLower {
			return &items[i], nil
		}
	}
	return nil, nil
}

// findService loads config and returns the service matching name+adapterType.
// Prefers the entry whose adapter field matches adapterType to resolve duplicate names.
func findService(name, adapterType string) (*serviceItem, error) {
	format := detectFormat()
	data, err := readFile(configPath(format))
	if err != nil {
		return nil, err
	}
	items, err := loadAllServiceItems(data, format)
	if err != nil {
		return nil, err
	}
	var fallback *serviceItem
	for i, svc := range items {
		if svc.Name != name {
			continue
		}
		if svc.Adapter == adapterType {
			return &items[i], nil
		}
		if fallback == nil {
			fallback = &items[i]
		}
	}
	return fallback, nil
}

// ─── HTTP handler ─────────────────────────────────────────────────

// handleAdapter godoc
// @Summary     Run adapter for a service
// @Description Fetches live stats from the named adapter for the given service. Results are cached for 60 s.
// @Tags        adapters
// @Param       type    path  string true "Adapter type (e.g. pihole, proxmox, sonarr)"
// @Param       service query string true "Service name as defined in config"
// @Produce     json
// @Success     200 {object} object "AdapterResult with stats array"
// @Failure     400 {string} string "Missing adapter type or service parameter"
// @Failure     405 {string} string "Method not allowed"
// @Router      /adapters/{type} [get]
func handleAdapter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adapterType := strings.TrimPrefix(r.URL.Path, "/adapters/")
	adapterType = strings.Trim(adapterType, "/")
	if adapterType == "" {
		http.Error(w, "adapter type required in path", http.StatusBadRequest)
		return
	}

	rawName := r.URL.Query().Get("service")
	if rawName == "" {
		http.Error(w, "service parameter required", http.StatusBadRequest)
		return
	}
	serviceName, err := url.QueryUnescape(rawName)
	if err != nil {
		http.Error(w, "invalid service name", http.StatusBadRequest)
		return
	}

	cacheKey := adapterType + ":" + serviceName

	if cached, ok := adapters.GetCached(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cached)
		return
	}

	svc, err := findService(serviceName, adapterType)
	if err != nil {
		writeJSON(w, adapters.ErrResult(adapterType, "config read error: "+err.Error()))
		return
	}
	if svc == nil {
		writeJSON(w, adapters.ErrResult(adapterType, "service not found: "+serviceName))
		return
	}

	cfg := adapters.ExpandEnvVars(svc.AdapterConfig, loadSecrets())

	baseURL := svc.URL
	if u, ok := cfg["api_url"].(string); ok && u != "" {
		baseURL = u
	}
	baseURL = strings.TrimRight(baseURL, "/")

	result := adapters.Run(adapterType, cfg, baseURL)

	if result.Ok {
		adapters.SetCached(cacheKey, result)
	}

	writeJSON(w, result)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
