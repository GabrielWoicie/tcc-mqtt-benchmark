package mqttclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
)

type Config struct {
	BrokerURL     string
	ClientID      string
	Username      string
	Password      string
	KeepAlive     uint16
	DialTimeout   time.Duration
	PacketTimeout time.Duration
	TLSInsecure   bool
	OnPublish     func(paho.PublishReceived) (bool, error)
	OnClientError func(error)
}

type Client struct {
	Paho *paho.Client
}

func Connect(ctx context.Context, cfg Config) (*Client, time.Duration, error) {
	if cfg.KeepAlive == 0 {
		cfg.KeepAlive = 300
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 10 * time.Second
	}
	if cfg.PacketTimeout == 0 {
		cfg.PacketTimeout = 10 * time.Second
	}

	u, err := normalizeURL(cfg.BrokerURL)
	if err != nil {
		return nil, 0, err
	}

	started := time.Now()
	rawConn, err := dial(u, cfg.DialTimeout, cfg.TLSInsecure)
	if err != nil {
		return nil, time.Since(started), fmt.Errorf("TCP connection failed: %w", err)
	}

	conn := packets.NewThreadSafeConn(rawConn)

	clientCfg := paho.ClientConfig{
		ClientID:      cfg.ClientID,
		Conn:          conn,
		PacketTimeout: cfg.PacketTimeout,
		OnClientError: cfg.OnClientError,
	}
	if cfg.OnPublish != nil {
		clientCfg.OnPublishReceived = []func(paho.PublishReceived) (bool, error){cfg.OnPublish}
	}

	c := paho.NewClient(clientCfg)
	expiry := uint32(0)
	connectPacket := &paho.Connect{
		ClientID:   cfg.ClientID,
		KeepAlive:  cfg.KeepAlive,
		CleanStart: true,
		Properties: &paho.ConnectProperties{SessionExpiryInterval: &expiry},
	}
	if cfg.Username != "" {
		connectPacket.Username = cfg.Username
		connectPacket.UsernameFlag = true
	}
	if cfg.Password != "" {
		connectPacket.Password = []byte(cfg.Password)
		connectPacket.PasswordFlag = true
	}

	connack, err := c.Connect(ctx, connectPacket)
	latency := time.Since(started)
	if err != nil {
		_ = rawConn.Close()
		return nil, latency, fmt.Errorf("MQTT CONNECT failed: %w", err)
	}
	if connack == nil || connack.ReasonCode >= 0x80 {
		_ = rawConn.Close()
		if connack == nil {
			return nil, latency, fmt.Errorf("MQTT CONNECT returned no CONNACK")
		}
		return nil, latency, fmt.Errorf("MQTT CONNACK rejected connection with reason code %d", connack.ReasonCode)
	}

	return &Client{Paho: c}, latency, nil
}

func (c *Client) Disconnect() {
	if c == nil || c.Paho == nil {
		return
	}
	_ = c.Paho.Disconnect(&paho.Disconnect{ReasonCode: 0})
}

func normalizeURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("broker URL cannot be empty")
	}
	if !strings.Contains(raw, "://") {
		raw = "mqtt://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid broker URL: %w", err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid broker URL: host not found")
	}
	if u.Port() == "" {
		if u.Scheme == "mqtts" || u.Scheme == "ssl" || u.Scheme == "tls" {
			u.Host = net.JoinHostPort(u.Hostname(), "8883")
		} else {
			u.Host = net.JoinHostPort(u.Hostname(), "1883")
		}
	}
	return u, nil
}

func dial(u *url.URL, timeout time.Duration, insecure bool) (net.Conn, error) {
	d := &net.Dialer{Timeout: timeout}
	switch strings.ToLower(u.Scheme) {
	case "mqtt", "tcp":
		return d.Dial("tcp", u.Host)
	case "mqtts", "ssl", "tls":
		return tls.DialWithDialer(d, "tcp", u.Host, &tls.Config{
			ServerName:         u.Hostname(),
			InsecureSkipVerify: insecure, // optional only for controlled test environments
			MinVersion:         tls.VersionTLS12,
		})
	default:
		return nil, fmt.Errorf("unsupported scheme %q; use mqtt:// or mqtts://", u.Scheme)
	}
}
