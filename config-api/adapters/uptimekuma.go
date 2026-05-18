package adapters

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Uptime Kuma adapter — socket.io polling (v1.23 has no monitor REST API).
// adapter_config: { username: "${UPTIMEKUMA_USER}", password: "${UPTIMEKUMA_PASS}" }
func fetchUptimeKumaStats(cfg map[string]interface{}, baseURL string) AdapterResult {
	username := cfgStr(cfg, "username")
	password := cfgStr(cfg, "password")
	if username == "" || password == "" {
		return errResult("uptime-kuma", "username and password required")
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
		},
	}

	stats, err := sioMonitorStats(client, strings.TrimRight(baseURL, "/"), username, password)
	if err != nil {
		return errResult("uptime-kuma", err.Error())
	}
	return AdapterResult{Adapter: "uptime-kuma", Ok: true, Stats: stats}
}

func sioMonitorStats(client *http.Client, base, username, password string) ([]StatItem, error) {
	sioURL := base + "/socket.io/?EIO=4&transport=polling"

	get := func(sid string) ([]string, error) {
		u := sioURL
		if sid != "" {
			u += "&sid=" + sid
		}
		resp, err := client.Get(u)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return strings.Split(string(b), "\x1e"), nil
	}

	post := func(sid, data string) error {
		u := sioURL + "&sid=" + sid
		req, err := http.NewRequest("POST", u, strings.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		return resp.Body.Close()
	}

	parseEvent := func(raw string) (string, json.RawMessage) {
		if !strings.HasPrefix(raw, "42") {
			return "", nil
		}
		var arr []json.RawMessage
		if json.Unmarshal([]byte(raw[2:]), &arr) != nil || len(arr) < 1 {
			return "", nil
		}
		var name string
		json.Unmarshal(arr[0], &name) //nolint:errcheck
		if len(arr) > 1 {
			return name, arr[1]
		}
		return name, nil
	}

	pollAll := func(sid string) {
		packets, err := get(sid)
		if err != nil {
			return
		}
		for _, p := range packets {
			if p == "2" {
				post(sid, "3") //nolint:errcheck
			}
		}
	}

	// ── 1. Handshake
	packets, err := get("")
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	if len(packets) == 0 || len(packets[0]) < 2 || packets[0][0] != '0' {
		return nil, fmt.Errorf("unexpected handshake — check api_url points directly to Uptime Kuma")
	}
	var hs struct{ Sid string `json:"sid"` }
	if err := json.Unmarshal([]byte(packets[0][1:]), &hs); err != nil || hs.Sid == "" {
		return nil, fmt.Errorf("handshake parse failed")
	}
	sid := hs.Sid

	// ── 2. Namespace connect + drain
	post(sid, "40") //nolint:errcheck
	pollAll(sid)

	// ── 3. Login (socket.io ACK packet)
	loginPkt, _ := json.Marshal([]interface{}{"login", map[string]string{
		"username": username,
		"password": password,
		"token":    "",
	}})
	if err := post(sid, "421"+string(loginPkt)); err != nil {
		return nil, fmt.Errorf("login send failed: %v", err)
	}

	// ── 4. Collect monitorList + heartbeatList
	type monitorDef struct {
		Active bool `json:"active"`
	}
	type heartbeat struct {
		Status int `json:"status"`
	}

	var monitors map[string]monitorDef
	var heartbeats map[string][]heartbeat

	for i := 0; i < 20 && (monitors == nil || heartbeats == nil); i++ {
		packets, err := get(sid)
		if err != nil {
			break
		}
		for _, p := range packets {
			if p == "2" {
				post(sid, "3") //nolint:errcheck
				continue
			}
			if strings.HasPrefix(p, "431") {
				var arr []json.RawMessage
				if json.Unmarshal([]byte(p[3:]), &arr) == nil && len(arr) > 0 {
					var lr struct {
						Ok  bool   `json:"ok"`
						Msg string `json:"msg"`
					}
					if json.Unmarshal(arr[0], &lr) == nil && !lr.Ok {
						msg := lr.Msg
						if msg == "" {
							msg = "invalid credentials"
						}
						return nil, fmt.Errorf("authentication failed: %s", msg)
					}
				}
				continue
			}
			name, data := parseEvent(p)
			switch name {
			case "monitorList":
				if monitors == nil {
					json.Unmarshal(data, &monitors) //nolint:errcheck
				}
			case "heartbeatList":
				if heartbeats == nil {
					json.Unmarshal(data, &heartbeats) //nolint:errcheck
				}
			}
		}
	}

	if monitors == nil {
		return nil, fmt.Errorf("monitor list not received — check api_url")
	}

	// ── 5. Count up/down
	up, down, total := 0, 0, 0
	for id, m := range monitors {
		if !m.Active {
			continue
		}
		total++
		hbs := heartbeats[id]
		if len(hbs) == 0 {
			up++
			continue
		}
		if hbs[len(hbs)-1].Status == 1 {
			up++
		} else {
			down++
		}
	}

	stats := []StatItem{}
	if total == 0 {
		stats = append(stats, StatItem{Label: "Monitors", Value: "none active", Status: "info"})
	} else {
		st, val := "ok", fmt.Sprintf("%d up", up)
		if down > 0 {
			st = "error"
			val = fmt.Sprintf("%d up / %d down", up, down)
		}
		stats = append(stats, StatItem{Label: "Monitors", Value: val, Status: st})
	}
	return stats, nil
}
