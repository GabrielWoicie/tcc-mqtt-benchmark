package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tcc-mqtt-benchmark/internal/buildinfo"
	"tcc-mqtt-benchmark/internal/connectionbench"
	"tcc-mqtt-benchmark/internal/messagebench"
)

func main() {
	args, err := resolveArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch args[0] {
	case "message":
		err = runMessage(ctx, args[1:])
	case "connections":
		err = runConnections(ctx, args[1:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %s\n\n", args[0])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func runMessage(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("message", flag.ContinueOnError)
	broker := fs.String("broker", "", "Broker URL, e.g. mqtt://1.2.3.4:1883")
	scenario := fs.String("scenario-id", "", "Experiment identifier")
	repetition := fs.Int("repetition", 1, "Experiment repetition (>= 1)")
	pattern := fs.String("publish-pattern", "staggered", "Publication pattern (message only)")
	name := fs.String("broker-name", "broker", "Broker label used in result files, e.g. emqx or mosquitto")
	clients := fs.Int("clients", 100, "Number of publisher connections")
	qos := fs.Int("qos", 0, "MQTT QoS: 0, 1 or 2")
	rate := fs.Float64("rate", 1, "Messages per second per publisher")
	duration := fs.Duration("duration", 60*time.Second, "Measurement duration")
	warmup := fs.Duration("warmup", 10*time.Second, "Warm-up before measurement")
	drain := fs.Duration("drain", 5*time.Second, "Time to wait for in-flight messages after measurement")
	payloadBytes := fs.Int("payload-bytes", 256, "Approximate JSON payload target size in bytes; minimum 64")
	prefix := fs.String("topic-prefix", "tcc/benchmark", "Base topic")
	connectRate := fs.Float64("connect-rate", 100, "Publisher connection ramp in connections/s; 0 means no throttle")
	username := fs.String("username", "", "Optional MQTT username")
	password := fs.String("password", "", "Optional MQTT password")
	out := fs.String("out", "results", "Results root directory")
	tlsInsecure := fs.Bool("tls-insecure", false, "Skip TLS certificate validation (only for controlled tests)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *repetition < 1 {
		return fmt.Errorf("repetition must be >= 1")
	}
	if *pattern != "staggered" && *pattern != "synchronized" {
		return fmt.Errorf("publish-pattern must be staggered or synchronized")
	}
	if *broker == "" {
		return fmt.Errorf("--broker is required")
	}
	if *qos < 0 || *qos > 2 {
		return fmt.Errorf("--qos must be 0, 1 or 2")
	}

	fmt.Printf("Starting message benchmark: scenario=%s repetition=%d pattern=%s broker=%s clients=%d qos=%d rate=%.3f msg/s/client duration=%s payload=%d B\n", *scenario, *repetition, *pattern, *name, *clients, *qos, *rate, *duration, *payloadBytes)
	summary, err := messagebench.Run(ctx, messagebench.Config{ScenarioID: *scenario, Repetition: *repetition, PublishPattern: *pattern,
		BrokerURL: *broker, BrokerName: *name, Clients: *clients, QoS: byte(*qos), Rate: *rate,
		Duration: *duration, Warmup: *warmup, Drain: *drain, PayloadBytes: *payloadBytes, TopicPrefix: *prefix,
		ConnectRate: *connectRate, Username: *username, Password: *password, OutputRoot: *out, TLSInsecure: *tlsInsecure,
	})
	if err != nil {
		return err
	}
	printJSON(summary)
	return nil
}

func runConnections(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("connections", flag.ContinueOnError)
	broker := fs.String("broker", "", "Broker URL, e.g. mqtt://1.2.3.4:1883")
	scenario := fs.String("scenario-id", "", "Experiment identifier")
	repetition := fs.Int("repetition", 1, "Experiment repetition (>= 1)")
	fs.String("publish-pattern", "", "Rejected in connections mode")
	name := fs.String("broker-name", "broker", "Broker label used in result files")
	clients := fs.Int("clients", 1000, "Number of simultaneous MQTT client connections")
	connectRate := fs.Float64("connect-rate", 100, "Connection attempts/s; 0 means launch without throttle")
	hold := fs.Duration("hold", 60*time.Second, "How long to keep successful connections open after ramp")
	username := fs.String("username", "", "Optional MQTT username")
	password := fs.String("password", "", "Optional MQTT password")
	out := fs.String("out", "results", "Results root directory")
	tlsInsecure := fs.Bool("tls-insecure", false, "Skip TLS certificate validation (only for controlled tests)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *repetition < 1 {
		return fmt.Errorf("repetition must be >= 1")
	}
	explicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "publish-pattern" {
			explicit = true
		}
	})
	if explicit {
		return fmt.Errorf("publish-pattern is only valid for mode=message")
	}
	if *broker == "" {
		return fmt.Errorf("--broker is required")
	}

	fmt.Printf("Starting connection benchmark: broker=%s clients=%d connect-rate=%.3f/s hold=%s\n", *name, *clients, *connectRate, *hold)
	summary, err := connectionbench.Run(ctx, connectionbench.Config{ScenarioID: *scenario, Repetition: *repetition,
		BrokerURL: *broker, BrokerName: *name, Clients: *clients, ConnectRate: *connectRate, Hold: *hold,
		Username: *username, Password: *password, OutputRoot: *out, TLSInsecure: *tlsInsecure,
	})
	if err != nil {
		return err
	}
	printJSON(summary)
	return nil
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func usage() {
	fmt.Printf("mqttbench v%s - benchmark MQTT 5.0 for TCC\n\n", buildinfo.Version)
	fmt.Print(`

Usage:
  mqttbench                         Read ./mqttbench.conf
  mqttbench --config path.conf       Read a configuration file
  mqttbench --config path.conf [options]  Override file values
  mqttbench message [options]
  mqttbench connections [options]

Examples:
  mqttbench message --broker mqtt://IP:1883 --broker-name emqx --clients 100 --qos 0 --rate 1 --duration 60s
  mqttbench connections --broker mqtt://IP:1883 --broker-name emqx --clients 1000 --connect-rate 100 --hold 60s

Use "mqttbench <mode> -h" to see all options.`)
}
