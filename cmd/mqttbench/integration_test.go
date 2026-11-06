package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eclipse/paho.golang/packets"
)

// localBroker is a minimal MQTT 5 QoS 0 test server, not a production broker.
func localBroker(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var subscriber net.Conn
	var conns []net.Conn
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			raw, err := ln.Accept()
			if err != nil {
				return
			}
			c := packets.NewThreadSafeConn(raw)
			mu.Lock()
			conns = append(conns, c)
			mu.Unlock()
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer c.Close()
				for {
					p, err := packets.ReadPacket(c)
					if err != nil {
						return
					}
					switch v := p.Content.(type) {
					case *packets.Connect:
						_, err = (&packets.Connack{Properties: &packets.Properties{}}).WriteTo(c)
					case *packets.Subscribe:
						mu.Lock()
						subscriber = c
						mu.Unlock()
						_, err = (&packets.Suback{PacketID: v.PacketID, Reasons: []byte{0}, Properties: &packets.Properties{}}).WriteTo(c)
					case *packets.Publish:
						mu.Lock()
						sub := subscriber
						mu.Unlock()
						if sub != nil {
							_, err = (&packets.Publish{Topic: v.Topic, Payload: v.Payload, Properties: &packets.Properties{}}).WriteTo(sub)
						}
					case *packets.Pingreq:
						_, err = (&packets.Pingresp{}).WriteTo(c)
					case *packets.Disconnect:
						return
					}
					if err != nil {
						return
					}
				}
			}()
		}
	}()
	t.Cleanup(func() {
		ln.Close()
		mu.Lock()
		for _, c := range conns {
			c.Close()
		}
		mu.Unlock()
		wg.Wait()
	})
	return "mqtt://" + ln.Addr().String()
}

func TestMessageIntegration(t *testing.T) {
	for _, pattern := range []string{"staggered", "synchronized"} {
		t.Run(pattern, func(t *testing.T) {
			broker := localBroker(t)
			root := t.TempDir()
			conf := filepath.Join(root, "test.conf")
			body := fmt.Sprintf("mode=message\nbroker=%s\nclients=2\nrate=1\nwarmup=1s\nduration=15s\ndrain=200ms\nconnect-rate=0\nscenario-id=TESTE-001/unsafe\nrepetition=2\npassword=123456\nout=%s\npublish-pattern=%s\n", broker, root, pattern)
			if err := os.WriteFile(conf, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			args, err := resolveArgs([]string{"--config", conf, "--clients=5"})
			if err != nil {
				t.Fatal(err)
			}
			if err := runMessage(context.Background(), args[1:]); err != nil {
				t.Fatal(err)
			}
			dirs, _ := filepath.Glob(filepath.Join(root, "TESTE-001_unsafe_*"))
			if len(dirs) != 1 {
				t.Fatalf("result directories: %v", dirs)
			}
			dir := dirs[0]
			raw, err := os.ReadFile(filepath.Join(dir, "summary.json"))
			if err != nil {
				t.Fatal(err)
			}
			var s map[string]any
			if err := json.Unmarshal(raw, &s); err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]any{"tool_version": "0.3.0", "scenario_id": "TESTE-001/unsafe", "repetition": float64(2), "publish_pattern": pattern, "clients_requested": float64(5), "publish_attempts": float64(75), "publish_success": float64(75), "received_unique": float64(75), "lost": float64(0), "duplicates": float64(0)} {
				if s[key] != want {
					t.Errorf("%s = %v; want %v", key, s[key], want)
				}
			}
			effective, err := os.ReadFile(filepath.Join(dir, "effective-config.conf"))
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"clients=5\n", "password=<redacted>\n", "repetition=2\n", "tool-version=0.3.0\n", "payload-bytes=256\n", "publish-pattern=" + pattern + "\n"} {
				if !strings.Contains(string(effective), want) {
					t.Errorf("missing %q", want)
				}
			}
			if strings.Contains(string(effective)+string(raw), "123456") {
				t.Fatal("password leaked")
			}
			rows := readCSV(t, filepath.Join(dir, "summary.csv"))
			for _, key := range []string{"tool_version", "scenario_id", "repetition", "publish_pattern", "lost", "latency_p99_ms"} {
				found := false
				for _, col := range rows[0] {
					if col == key {
						found = true
					}
				}
				if !found {
					t.Errorf("missing column %s", key)
				}
			}
			messages := readCSV(t, filepath.Join(dir, "messages.csv"))
			first := map[int]int64{}
			for _, row := range messages[1:] {
				idx, _ := strconv.Atoi(row[0])
				sent, _ := strconv.ParseInt(row[2], 10, 64)
				if first[idx] == 0 || sent < first[idx] {
					first[idx] = sent
				}
			}
			for i := 0; i < 5; i++ {
				offset := time.Duration(first[i] - first[0])
				want := time.Duration(0)
				if pattern == "staggered" {
					want = time.Duration(i) * 200 * time.Millisecond
				}
				if offset < want-100*time.Millisecond || offset > want+100*time.Millisecond {
					t.Errorf("client %d offset %s, want ~%s", i, offset, want)
				}
				t.Logf("client %d first measurement offset: %s", i, offset)
			}
		})
	}
}

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestConnectionsIntegration(t *testing.T) {
	root := t.TempDir()
	err := runConnections(context.Background(), []string{"--broker=" + localBroker(t), "--clients=3", "--hold=0s", "--connect-rate=0", "--scenario-id=CONN", "--repetition=2", "--password=secret", "--out=" + root})
	if err != nil {
		t.Fatal(err)
	}
	dirs, _ := filepath.Glob(filepath.Join(root, "CONN_*"))
	if len(dirs) != 1 {
		t.Fatalf("directories: %v", dirs)
	}
	for _, file := range []string{"summary.json", "summary.csv", "connections.csv", "effective-config.conf"} {
		raw, err := os.ReadFile(filepath.Join(dirs[0], file))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "secret") {
			t.Fatal("password leaked")
		}
	}
	var summary map[string]any
	raw, _ := os.ReadFile(filepath.Join(dirs[0], "summary.json"))
	json.Unmarshal(raw, &summary)
	if summary["clients_connected"] != float64(3) || summary["repetition"] != float64(2) || summary["scenario_id"] != "CONN" {
		t.Fatalf("summary: %v", summary)
	}
}

func TestNewValidation(t *testing.T) {
	for _, args := range [][]string{{"--repetition=0"}, {"--repetition=-1"}, {"--publish-pattern=invalid"}} {
		if err := runMessage(context.Background(), append(args, "--broker=mqtt://127.0.0.1:1")); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	for _, pattern := range []string{"staggered", "synchronized", ""} {
		err := runConnections(context.Background(), []string{"--publish-pattern=" + pattern})
		if err == nil || err.Error() != "publish-pattern is only valid for mode=message" {
			t.Fatalf("pattern %q: %v", pattern, err)
		}
	}
}
