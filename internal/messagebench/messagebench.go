package messagebench

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.golang/paho"

	"tcc-mqtt-benchmark/internal/buildinfo"
	"tcc-mqtt-benchmark/internal/mqttclient"
	"tcc-mqtt-benchmark/internal/runfiles"
	"tcc-mqtt-benchmark/internal/stats"
)

type Config struct {
	PublishPattern string
	ScenarioID     string
	Repetition     int
	BrokerURL      string
	BrokerName     string
	Clients        int
	QoS            byte
	Rate           float64
	Duration       time.Duration
	Warmup         time.Duration
	Drain          time.Duration
	PayloadBytes   int
	TopicPrefix    string
	ConnectRate    float64
	Username       string
	Password       string
	OutputRoot     string
	TLSInsecure    bool
}

type payload struct {
	RunID       string `json:"run_id"`
	ClientIndex int    `json:"client_index"`
	Sequence    uint64 `json:"sequence"`
	SentAtNS    int64  `json:"sent_at_ns"`
	Phase       string `json:"phase"`
	Padding     string `json:"padding"`
}

type messageRecord struct {
	ClientIndex  int
	Sequence     uint64
	SentAtNS     int64
	ReceivedAtNS int64
	LatencyNS    int64
	Duplicate    bool
}

type collector struct {
	runID        string
	mu           sync.Mutex
	seen         map[string]struct{}
	records      []messageRecord
	duplicates   uint64
	decodeErrors uint64
}

type Summary struct {
	ScenarioID               string             `json:"scenario_id"`
	Repetition               int                `json:"repetition"`
	PublishPattern           string             `json:"publish_pattern"`
	ToolVersion              string             `json:"tool_version"`
	Mode                     string             `json:"mode"`
	RunID                    string             `json:"run_id"`
	BrokerName               string             `json:"broker_name"`
	BrokerURL                string             `json:"broker_url"`
	MQTTVersion              string             `json:"mqtt_version"`
	QoS                      byte               `json:"qos"`
	ClientsRequested         int                `json:"clients_requested"`
	ClientsConnected         int                `json:"clients_connected"`
	PublisherConnectFailures int                `json:"publisher_connect_failures"`
	RatePerClient            float64            `json:"rate_per_client_msg_s"`
	TargetMessageRate        float64            `json:"target_message_rate_msg_s"`
	PayloadBytes             int                `json:"payload_bytes"`
	WarmupSeconds            float64            `json:"warmup_seconds"`
	DurationSeconds          float64            `json:"duration_seconds"`
	DrainSeconds             float64            `json:"drain_seconds"`
	PublishAttempts          uint64             `json:"publish_attempts"`
	PublishSuccess           uint64             `json:"publish_success"`
	PublishErrors            uint64             `json:"publish_errors"`
	ReceivedUnique           uint64             `json:"received_unique"`
	Duplicates               uint64             `json:"duplicates"`
	DecodeErrors             uint64             `json:"decode_errors"`
	Lost                     int64              `json:"lost"`
	DeliveryRatePct          float64            `json:"delivery_rate_pct"`
	PublishThroughput        float64            `json:"publish_throughput_msg_s"`
	ReceiveThroughput        float64            `json:"receive_throughput_msg_s"`
	Latency                  stats.Distribution `json:"latency"`
	PublisherConnectLatency  stats.Distribution `json:"publisher_connect_latency"`
	StartedAtUTC             string             `json:"started_at_utc"`
	FinishedAtUTC            string             `json:"finished_at_utc"`
	GeneratorOS              string             `json:"generator_os"`
	GeneratorArch            string             `json:"generator_arch"`
	GoVersion                string             `json:"go_version"`
	ResultDirectory          string             `json:"result_directory"`
}

func Run(parent context.Context, cfg Config) (Summary, error) {
	if cfg.Repetition < 1 {
		return Summary{}, fmt.Errorf("repetition must be >= 1")
	}
	if cfg.PublishPattern == "" {
		cfg.PublishPattern = "staggered"
	}
	if cfg.OutputRoot == "" {
		cfg.OutputRoot = "results"
	}
	if err := validate(cfg); err != nil {
		return Summary{}, err
	}
	runID := time.Now().UTC().Format("20060102T150405.000000000Z")
	safeRunID := strings.ReplaceAll(strings.ReplaceAll(runID, ":", ""), ".", "")
	resultDir := filepath.Join(cfg.OutputRoot, fmt.Sprintf("%s_message_qos%d_c%d_r%s_%s", sanitize(cfg.BrokerName), cfg.QoS, cfg.Clients, formatRate(cfg.Rate), safeRunID))
	if cfg.ScenarioID != "" {
		resultDir = filepath.Join(cfg.OutputRoot, runfiles.Sanitize(cfg.ScenarioID)+"_"+runfiles.Sanitize(cfg.BrokerName)+"_"+safeRunID)
	}
	if err := os.MkdirAll(resultDir, 0o755); err != nil {
		return Summary{}, err
	}

	if err := runfiles.WriteConfig(filepath.Join(resultDir, "effective-config.conf"), "message", cfg); err != nil {
		return Summary{}, err
	}
	started := time.Now().UTC()
	col := &collector{runID: runID, seen: make(map[string]struct{})}

	// One dedicated collector connection subscribes to all topics in the run.
	collectorID := fmt.Sprintf("tcc-collector-%s", shortID(safeRunID))
	sub, _, err := mqttclient.Connect(parent, mqttclient.Config{
		BrokerURL:   cfg.BrokerURL,
		ClientID:    collectorID,
		Username:    cfg.Username,
		Password:    cfg.Password,
		KeepAlive:   300,
		TLSInsecure: cfg.TLSInsecure,
		OnPublish:   col.onPublish,
	})
	if err != nil {
		return Summary{}, fmt.Errorf("collector connection failed: %w", err)
	}
	defer sub.Disconnect()

	topicRoot := strings.TrimSuffix(cfg.TopicPrefix, "/") + "/" + safeRunID
	subCtx, cancelSub := context.WithTimeout(parent, 10*time.Second)
	_, err = sub.Paho.Subscribe(subCtx, &paho.Subscribe{Subscriptions: []paho.SubscribeOptions{{Topic: topicRoot + "/#", QoS: cfg.QoS}}})
	cancelSub()
	if err != nil {
		return Summary{}, fmt.Errorf("collector subscribe failed: %w", err)
	}

	// Create publisher connections in a controlled ramp so the setup phase does not become the benchmark itself.
	publishers := make([]*mqttclient.Client, cfg.Clients)
	connectLatencies := make([]time.Duration, 0, cfg.Clients)
	var connectMu sync.Mutex
	var connectFailures atomic.Uint64
	var wgConnect sync.WaitGroup

	interval := rateInterval(cfg.ConnectRate)
	var connectTicker *time.Ticker
	if interval > 0 {
		connectTicker = time.NewTicker(interval)
		defer connectTicker.Stop()
	}

	for i := 0; i < cfg.Clients; i++ {
		if connectTicker != nil && i > 0 {
			select {
			case <-parent.Done():
				return Summary{}, parent.Err()
			case <-connectTicker.C:
			}
		}
		idx := i
		wgConnect.Add(1)
		go func() {
			defer wgConnect.Done()
			id := fmt.Sprintf("tcc-pub-%s-%06d", shortID(safeRunID), idx)
			ctx, cancel := context.WithTimeout(parent, 15*time.Second)
			c, latency, err := mqttclient.Connect(ctx, mqttclient.Config{
				BrokerURL:   cfg.BrokerURL,
				ClientID:    id,
				Username:    cfg.Username,
				Password:    cfg.Password,
				KeepAlive:   300,
				TLSInsecure: cfg.TLSInsecure,
			})
			cancel()
			if err != nil {
				connectFailures.Add(1)
				return
			}
			publishers[idx] = c
			connectMu.Lock()
			connectLatencies = append(connectLatencies, latency)
			connectMu.Unlock()
		}()
	}
	wgConnect.Wait()

	connected := 0
	for _, c := range publishers {
		if c != nil {
			connected++
		}
	}
	if connected == 0 {
		return Summary{}, fmt.Errorf("no publisher could connect")
	}

	defer func() {
		var wg sync.WaitGroup
		for _, c := range publishers {
			if c == nil {
				continue
			}
			wg.Add(1)
			go func(cli *mqttclient.Client) {
				defer wg.Done()
				cli.Disconnect()
			}(c)
		}
		wg.Wait()
	}()

	var attempts atomic.Uint64
	var successes atomic.Uint64
	var publishErrors atomic.Uint64

	// Warm-up and measurement use a common global clock, minimizing start-time skew among publishers.
	warmupStart := time.Now()
	measureStart := warmupStart.Add(cfg.Warmup)
	measureEnd := measureStart.Add(cfg.Duration)

	var wgPublish sync.WaitGroup
	for idx, c := range publishers {
		if c == nil {
			continue
		}
		wgPublish.Add(1)
		go func(clientIndex int, cli *mqttclient.Client) {
			defer wgPublish.Done()
			publishLoop(parent, cfg, cli, clientIndex, runID, topicRoot, warmupStart, measureStart, measureEnd, &attempts, &successes, &publishErrors)
		}(idx, c)
	}
	wgPublish.Wait()

	// Keep the full measurement window even when the last slot precedes its end.
	if remaining := time.Until(measureEnd); remaining > 0 {
		timer := time.NewTimer(remaining)
		select {
		case <-parent.Done():
			timer.Stop()
		case <-timer.C:
		}
	}

	// Allow already-published messages to arrive before results are closed.
	if cfg.Drain > 0 {
		timer := time.NewTimer(cfg.Drain)
		select {
		case <-parent.Done():
			timer.Stop()
		case <-timer.C:
		}
	}

	col.mu.Lock()
	records := append([]messageRecord(nil), col.records...)
	duplicates := col.duplicates
	decodeErrors := col.decodeErrors
	col.mu.Unlock()

	latencies := make([]time.Duration, 0, len(records))
	for _, r := range records {
		if !r.Duplicate && r.LatencyNS >= 0 {
			latencies = append(latencies, time.Duration(r.LatencyNS))
		}
	}

	receivedUnique := uint64(len(latencies))
	successCount := successes.Load()
	lost := int64(successCount) - int64(receivedUnique)
	if lost < 0 {
		// It is possible for a QoS1/2 message to be received before Publish() returns an error/timeout.
		// Keep loss non-negative and expose raw counters for interpretation.
		lost = 0
	}
	delivery := 0.0
	if successCount > 0 {
		delivery = float64(receivedUnique) * 100 / float64(successCount)
	}

	summary := Summary{ScenarioID: cfg.ScenarioID, Repetition: cfg.Repetition, PublishPattern: cfg.PublishPattern,
		ToolVersion:              buildinfo.Version,
		Mode:                     "message",
		RunID:                    runID,
		BrokerName:               cfg.BrokerName,
		BrokerURL:                runfiles.RedactURL(cfg.BrokerURL),
		MQTTVersion:              "5.0",
		QoS:                      cfg.QoS,
		ClientsRequested:         cfg.Clients,
		ClientsConnected:         connected,
		PublisherConnectFailures: int(connectFailures.Load()),
		RatePerClient:            cfg.Rate,
		TargetMessageRate:        float64(connected) * cfg.Rate,
		PayloadBytes:             cfg.PayloadBytes,
		WarmupSeconds:            cfg.Warmup.Seconds(),
		DurationSeconds:          cfg.Duration.Seconds(),
		DrainSeconds:             cfg.Drain.Seconds(),
		PublishAttempts:          attempts.Load(),
		PublishSuccess:           successCount,
		PublishErrors:            publishErrors.Load(),
		ReceivedUnique:           receivedUnique,
		Duplicates:               duplicates,
		DecodeErrors:             decodeErrors,
		Lost:                     lost,
		DeliveryRatePct:          delivery,
		PublishThroughput:        float64(successCount) / cfg.Duration.Seconds(),
		ReceiveThroughput:        float64(receivedUnique) / cfg.Duration.Seconds(),
		Latency:                  stats.FromDurations(latencies),
		PublisherConnectLatency:  stats.FromDurations(connectLatencies),
		StartedAtUTC:             started.Format(time.RFC3339Nano),
		FinishedAtUTC:            time.Now().UTC().Format(time.RFC3339Nano),
		GeneratorOS:              runtime.GOOS,
		GeneratorArch:            runtime.GOARCH,
		GoVersion:                runtime.Version(),
		ResultDirectory:          resultDir,
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].ClientIndex == records[j].ClientIndex {
			return records[i].Sequence < records[j].Sequence
		}
		return records[i].ClientIndex < records[j].ClientIndex
	})

	if err := writeMessageCSV(filepath.Join(resultDir, "messages.csv"), records); err != nil {
		return Summary{}, err
	}
	if err := writeJSON(filepath.Join(resultDir, "summary.json"), summary); err != nil {
		return Summary{}, err
	}
	if err := writeSummaryCSV(filepath.Join(resultDir, "summary.csv"), summary); err != nil {
		return Summary{}, err
	}

	return summary, nil
}

func publishLoop(parent context.Context, cfg Config, cli *mqttclient.Client, clientIndex int, runID, topicRoot string, warmupStart, measureStart, measureEnd time.Time, attempts, successes, errors *atomic.Uint64) {
	interval := rateInterval(cfg.Rate)
	next := warmupStart.Add(publishOffset(cfg, clientIndex))
	timer := time.NewTimer(time.Until(next))
	defer timer.Stop()
	var seq uint64

	for next.Before(measureEnd) {
		select {
		case <-parent.Done():
			return
		case <-timer.C:
			now := next
			if !time.Now().Before(measureEnd) {
				return
			}
			phase := "warmup"
			if !now.Before(measureStart) {
				phase = "measure"
			}
			if now.Before(warmupStart) {
				continue
			}
			seq++
			sentAt := time.Now().UnixNano()
			body := makePayload(cfg.PayloadBytes, payload{
				RunID:       runID,
				ClientIndex: clientIndex,
				Sequence:    seq,
				SentAtNS:    sentAt,
				Phase:       phase,
			})
			topic := fmt.Sprintf("%s/%06d", topicRoot, clientIndex)

			if phase == "measure" {
				attempts.Add(1)
			}
			pubCtx, cancel := context.WithTimeout(parent, 10*time.Second)
			_, err := cli.Paho.Publish(pubCtx, &paho.Publish{QoS: cfg.QoS, Topic: topic, Payload: body})
			cancel()
			if phase == "measure" {
				if err != nil {
					errors.Add(1)
				} else {
					successes.Add(1)
				}
			}
			next = next.Add(interval)
			// Skip missed slots instead of generating catch-up bursts.
			if late := time.Since(next); late >= interval {
				next = next.Add((late / interval) * interval)
			}
			timer.Reset(time.Until(next))
		}
	}
}

func publishOffset(cfg Config, index int) time.Duration {
	if cfg.PublishPattern == "synchronized" {
		return 0
	}
	return time.Duration(float64(rateInterval(cfg.Rate)) * float64(index) / float64(cfg.Clients))
}

func (c *collector) onPublish(pr paho.PublishReceived) (bool, error) {
	receivedAt := time.Now().UnixNano()
	var p payload
	if err := json.Unmarshal(pr.Packet.Payload, &p); err != nil {
		c.mu.Lock()
		c.decodeErrors++
		c.mu.Unlock()
		return true, nil
	}
	if p.RunID != c.runID || p.Phase != "measure" {
		return true, nil
	}

	key := strconv.Itoa(p.ClientIndex) + ":" + strconv.FormatUint(p.Sequence, 10)
	c.mu.Lock()
	_, duplicate := c.seen[key]
	if duplicate {
		c.duplicates++
	} else {
		c.seen[key] = struct{}{}
	}
	c.records = append(c.records, messageRecord{
		ClientIndex:  p.ClientIndex,
		Sequence:     p.Sequence,
		SentAtNS:     p.SentAtNS,
		ReceivedAtNS: receivedAt,
		LatencyNS:    receivedAt - p.SentAtNS,
		Duplicate:    duplicate,
	})
	c.mu.Unlock()
	return true, nil
}

func makePayload(target int, p payload) []byte {
	p.Padding = ""
	base, _ := json.Marshal(p)
	if target <= len(base) {
		return base
	}
	p.Padding = strings.Repeat("x", target-len(base))
	out, _ := json.Marshal(p)
	return out
}

func validate(cfg Config) error {
	if cfg.PublishPattern != "staggered" && cfg.PublishPattern != "synchronized" {
		return fmt.Errorf("publish-pattern must be staggered or synchronized")
	}
	if cfg.Clients <= 0 {
		return fmt.Errorf("clients must be > 0")
	}
	if cfg.QoS > 2 {
		return fmt.Errorf("qos must be 0, 1 or 2")
	}
	if cfg.Rate <= 0 {
		return fmt.Errorf("rate must be > 0")
	}
	if cfg.Duration <= 0 {
		return fmt.Errorf("duration must be > 0")
	}
	if cfg.PayloadBytes < 64 {
		return fmt.Errorf("payload-bytes must be >= 64")
	}
	if cfg.TopicPrefix == "" {
		return fmt.Errorf("topic-prefix cannot be empty")
	}
	if cfg.OutputRoot == "" {
		cfg.OutputRoot = "results"
	}
	return nil
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
	r := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-")
	s = r.Replace(s)
	if s == "" {
		return "broker"
	}
	return s
}

func formatRate(v float64) string {
	return strings.ReplaceAll(strconv.FormatFloat(v, 'f', -1, 64), ".", "p")
}

func writeMessageCSV(path string, records []messageRecord) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"client_index", "sequence", "sent_at_ns", "received_at_ns", "latency_ms", "duplicate"})
	for _, r := range records {
		_ = w.Write([]string{
			strconv.Itoa(r.ClientIndex), strconv.FormatUint(r.Sequence, 10), strconv.FormatInt(r.SentAtNS, 10),
			strconv.FormatInt(r.ReceivedAtNS, 10), strconv.FormatFloat(float64(r.LatencyNS)/1e6, 'f', 6, 64), strconv.FormatBool(r.Duplicate),
		})
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

func writeSummaryCSV(path string, s Summary) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	header := []string{"broker", "qos", "clients_connected", "rate_per_client", "payload_bytes", "duration_s", "publish_success", "received_unique", "lost", "duplicates", "delivery_rate_pct", "publish_throughput_msg_s", "receive_throughput_msg_s", "latency_avg_ms", "latency_p50_ms", "latency_p95_ms", "latency_p99_ms", "latency_max_ms"}
	values := []string{
		s.BrokerName, strconv.Itoa(int(s.QoS)), strconv.Itoa(s.ClientsConnected), strconv.FormatFloat(s.RatePerClient, 'f', -1, 64), strconv.Itoa(s.PayloadBytes),
		strconv.FormatFloat(s.DurationSeconds, 'f', -1, 64), strconv.FormatUint(s.PublishSuccess, 10), strconv.FormatUint(s.ReceivedUnique, 10), strconv.FormatInt(s.Lost, 10), strconv.FormatUint(s.Duplicates, 10),
		strconv.FormatFloat(s.DeliveryRatePct, 'f', 6, 64), strconv.FormatFloat(s.PublishThroughput, 'f', 6, 64), strconv.FormatFloat(s.ReceiveThroughput, 'f', 6, 64),
		strconv.FormatFloat(s.Latency.AvgMS, 'f', 6, 64), strconv.FormatFloat(s.Latency.P50MS, 'f', 6, 64), strconv.FormatFloat(s.Latency.P95MS, 'f', 6, 64), strconv.FormatFloat(s.Latency.P99MS, 'f', 6, 64), strconv.FormatFloat(s.Latency.MaxMS, 'f', 6, 64),
	}
	header = append(header, "tool_version", "scenario_id", "repetition", "publish_pattern")
	values = append(values, s.ToolVersion, s.ScenarioID, strconv.Itoa(s.Repetition), s.PublishPattern)
	_ = w.Write(header)
	_ = w.Write(values)
	return w.Error()
}
