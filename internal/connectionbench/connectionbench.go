package connectionbench

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"tcc-mqtt-benchmark/internal/buildinfo"
	"tcc-mqtt-benchmark/internal/mqttclient"
	"tcc-mqtt-benchmark/internal/runfiles"
	"tcc-mqtt-benchmark/internal/stats"
)

type Config struct {
	ScenarioID  string
	Repetition  int
	BrokerURL   string
	BrokerName  string
	Clients     int
	ConnectRate float64
	Hold        time.Duration
	Username    string
	Password    string
	OutputRoot  string
	TLSInsecure bool
}

type record struct {
	ClientIndex int
	Success     bool
	Latency     time.Duration
	Error       string
}

type Summary struct {
	ScenarioID              string             `json:"scenario_id"`
	Repetition              int                `json:"repetition"`
	PublishPattern          string             `json:"publish_pattern"`
	ToolVersion             string             `json:"tool_version"`
	Mode                    string             `json:"mode"`
	BrokerName              string             `json:"broker_name"`
	BrokerURL               string             `json:"broker_url"`
	MQTTVersion             string             `json:"mqtt_version"`
	ClientsRequested        int                `json:"clients_requested"`
	ClientsConnected        int                `json:"clients_connected"`
	ConnectionFailures      int                `json:"connection_failures"`
	RequestedConnectRate    float64            `json:"requested_connect_rate_per_s"`
	ActualConnectRate       float64            `json:"actual_successful_connect_rate_per_s"`
	ConnectionWindowSeconds float64            `json:"connection_window_seconds"`
	HoldSeconds             float64            `json:"hold_seconds"`
	ConnectionLatency       stats.Distribution `json:"connection_latency"`
	StartedAtUTC            string             `json:"started_at_utc"`
	FinishedAtUTC           string             `json:"finished_at_utc"`
	GeneratorOS             string             `json:"generator_os"`
	GeneratorArch           string             `json:"generator_arch"`
	GoVersion               string             `json:"go_version"`
	ResultDirectory         string             `json:"result_directory"`
}

func Run(parent context.Context, cfg Config) (Summary, error) {
	if cfg.Repetition < 1 {
		return Summary{}, fmt.Errorf("repetition must be >= 1")
	}
	if cfg.Clients <= 0 {
		return Summary{}, fmt.Errorf("clients must be > 0")
	}
	if cfg.Hold < 0 {
		return Summary{}, fmt.Errorf("hold cannot be negative")
	}
	if cfg.OutputRoot == "" {
		cfg.OutputRoot = "results"
	}

	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	safeStamp := strings.ReplaceAll(strings.ReplaceAll(stamp, ":", ""), ".", "")
	resultDir := filepath.Join(cfg.OutputRoot, fmt.Sprintf("%s_connections_c%d_cr%s_%s", sanitize(cfg.BrokerName), cfg.Clients, formatRate(cfg.ConnectRate), safeStamp))
	if cfg.ScenarioID != "" {
		resultDir = filepath.Join(cfg.OutputRoot, runfiles.Sanitize(cfg.ScenarioID)+"_"+runfiles.Sanitize(cfg.BrokerName)+"_"+safeStamp)
	}
	if err := os.MkdirAll(resultDir, 0o755); err != nil {
		return Summary{}, err
	}

	if err := runfiles.WriteConfig(filepath.Join(resultDir, "effective-config.conf"), "connections", cfg); err != nil {
		return Summary{}, err
	}
	started := time.Now().UTC()
	clients := make([]*mqttclient.Client, cfg.Clients)
	records := make([]record, cfg.Clients)
	var wg sync.WaitGroup
	interval := rateInterval(cfg.ConnectRate)
	var ticker *time.Ticker
	if interval > 0 {
		ticker = time.NewTicker(interval)
		defer ticker.Stop()
	}

	windowStart := time.Now()
	for i := 0; i < cfg.Clients; i++ {
		if ticker != nil && i > 0 {
			select {
			case <-parent.Done():
				return Summary{}, parent.Err()
			case <-ticker.C:
			}
		}
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("tcc-conn-%s-%06d", shortID(safeStamp), idx)
			ctx, cancel := context.WithTimeout(parent, 15*time.Second)
			cli, latency, err := mqttclient.Connect(ctx, mqttclient.Config{
				BrokerURL:   cfg.BrokerURL,
				ClientID:    id,
				Username:    cfg.Username,
				Password:    cfg.Password,
				KeepAlive:   300,
				TLSInsecure: cfg.TLSInsecure,
			})
			cancel()
			rec := record{ClientIndex: idx, Latency: latency}
			if err != nil {
				rec.Error = err.Error()
				records[idx] = rec
				return
			}
			rec.Success = true
			clients[idx] = cli
			records[idx] = rec
		}()
	}
	wg.Wait()
	window := time.Since(windowStart)

	connected := 0
	latencies := make([]time.Duration, 0, cfg.Clients)
	for _, r := range records {
		if r.Success {
			connected++
			latencies = append(latencies, r.Latency)
		}
	}

	if cfg.Hold > 0 {
		fmt.Printf("%d connections established. Holding them for %s...\n", connected, cfg.Hold)
		timer := time.NewTimer(cfg.Hold)
		select {
		case <-parent.Done():
			timer.Stop()
		case <-timer.C:
		}
	}

	var wgDisc sync.WaitGroup
	for _, cli := range clients {
		if cli == nil {
			continue
		}
		wgDisc.Add(1)
		go func(c *mqttclient.Client) { defer wgDisc.Done(); c.Disconnect() }(cli)
	}
	wgDisc.Wait()

	rate := 0.0
	if window.Seconds() > 0 {
		rate = float64(connected) / window.Seconds()
	}
	summary := Summary{ScenarioID: cfg.ScenarioID, Repetition: cfg.Repetition,
		ToolVersion: buildinfo.Version,
		Mode:        "connections", BrokerName: cfg.BrokerName, BrokerURL: runfiles.RedactURL(cfg.BrokerURL), MQTTVersion: "5.0",
		ClientsRequested: cfg.Clients, ClientsConnected: connected, ConnectionFailures: cfg.Clients - connected,
		RequestedConnectRate: cfg.ConnectRate, ActualConnectRate: rate, ConnectionWindowSeconds: window.Seconds(), HoldSeconds: cfg.Hold.Seconds(),
		ConnectionLatency: stats.FromDurations(latencies), StartedAtUTC: started.Format(time.RFC3339Nano), FinishedAtUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GeneratorOS: runtime.GOOS, GeneratorArch: runtime.GOARCH, GoVersion: runtime.Version(), ResultDirectory: resultDir,
	}
	if err := writeCSV(filepath.Join(resultDir, "connections.csv"), records); err != nil {
		return Summary{}, err
	}
	if err := writeJSON(filepath.Join(resultDir, "summary.json"), summary); err != nil {
		return Summary{}, err
	}
	if err := runfiles.WriteSummaryCSV(filepath.Join(resultDir, "summary.csv"), summary); err != nil {
		return Summary{}, err
	}
	return summary, nil
}

func writeCSV(path string, records []record) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"client_index", "success", "connection_latency_ms", "error"})
	for _, r := range records {
		_ = w.Write([]string{strconv.Itoa(r.ClientIndex), strconv.FormatBool(r.Success), strconv.FormatFloat(float64(r.Latency)/float64(time.Millisecond), 'f', 6, 64), r.Error})
	}
	return w.Error()
}
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
func rateInterval(rate float64) time.Duration {
	if rate <= 0 {
		return 0
	}
	d := time.Duration(float64(time.Second) / rate)
	if d < time.Microsecond {
		return time.Microsecond
	}
	return d
}
func shortID(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[len(s)-12:]
}
func sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-").Replace(s)
	if s == "" {
		return "broker"
	}
	return s
}
func formatRate(v float64) string {
	if v <= 0 {
		return "max"
	}
	return strings.ReplaceAll(strconv.FormatFloat(v, 'f', -1, 64), ".", "p")
}
