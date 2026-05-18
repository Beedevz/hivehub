package adapters

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ─── Core Types ───────────────────────────────────────────────────

type AdapterResult struct {
	Adapter string     `json:"adapter"`
	Ok      bool       `json:"ok"`
	Stats   []StatItem `json:"stats"`
	Error   *string    `json:"error,omitempty"`
}

type StatItem struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Status string `json:"status"` // ok | warn | error | info
}

// ─── Exported Helpers ─────────────────────────────────────────────

// ErrResult builds a failed AdapterResult.
func ErrResult(adapterType, msg string) AdapterResult {
	return AdapterResult{Adapter: adapterType, Ok: false, Stats: []StatItem{}, Error: &msg}
}

// ExpandEnvVars resolves ${VAR} and ${secret:KEY} placeholders in adapter_config values.
// secrets map is optional (pass nil to skip secret resolution).
func ExpandEnvVars(config map[string]interface{}, secrets map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(config))
	for k, v := range config {
		if s, ok := v.(string); ok {
			result[k] = resolveVar(s, secrets)
		} else {
			result[k] = v
		}
	}
	return result
}

// resolveVar expands a single placeholder:
//   - ${secret:KEY} → looked up from secrets map
//   - ${VAR}        → looked up from environment
//   - anything else → returned as-is
func resolveVar(s string, secrets map[string]string) string {
	if strings.HasPrefix(s, "${secret:") && strings.HasSuffix(s, "}") {
		key := s[9 : len(s)-1]
		if secrets != nil {
			if val, ok := secrets[key]; ok {
				return val
			}
		}
		return ""
	}
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		return os.Getenv(s[2 : len(s)-1])
	}
	return s
}

// ─── Package-internal Helpers ─────────────────────────────────────

func errResult(adapterType, msg string) AdapterResult {
	return ErrResult(adapterType, msg)
}

func cfgStr(cfg map[string]interface{}, key string) string {
	if v, ok := cfg[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func newHTTPClient(skipTLS bool) *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: skipTLS, //nolint:gosec
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
}

func doJSON(client *http.Client, req *http.Request, dest interface{}) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

// ─── In-Memory Cache ──────────────────────────────────────────────

type cacheEntry struct {
	result AdapterResult
	at     time.Time
	ttl    time.Duration
}

var (
	adapterCache  = make(map[string]cacheEntry)
	cacheMu       sync.RWMutex
	cacheTTL      = 60 * time.Second
	cacheErrorTTL = 30 * time.Second // errors cached shorter to allow quicker recovery
)

func GetCached(key string) (AdapterResult, bool) {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	e, ok := adapterCache[key]
	if !ok {
		return AdapterResult{}, false
	}
	ttl := e.ttl
	if ttl == 0 {
		ttl = cacheTTL
	}
	if time.Since(e.at) < ttl {
		return e.result, true
	}
	return AdapterResult{}, false
}

// SetCached stores a successful result with the standard 60s TTL.
func SetCached(key string, r AdapterResult) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	adapterCache[key] = cacheEntry{result: r, at: time.Now(), ttl: cacheTTL}
}

// SetCachedError stores a failed result with a shorter 30s TTL so the adapter
// retries sooner once the remote service recovers, without hammering it on
// every request.
func SetCachedError(key string, r AdapterResult) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	adapterCache[key] = cacheEntry{result: r, at: time.Now(), ttl: cacheErrorTTL}
}

// ─── Shared format helpers ────────────────────────────────────────

func formatCount(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatSpeed(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/1024)
	}
	return fmt.Sprintf("%.1f MB/s", float64(bytesPerSec)/1024/1024)
}
