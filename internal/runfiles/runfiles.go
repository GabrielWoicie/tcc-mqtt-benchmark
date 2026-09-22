package runfiles

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"

	"tcc-mqtt-benchmark/internal/buildinfo"
)

func Sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, s)
}

func RedactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid-url>"
	}
	if u.User != nil {
		if _, ok := u.User.Password(); ok {
			u.User = url.UserPassword(u.User.Username(), "<redacted>")
		}
	}
	return u.String()
}

// WriteConfig serializes resolved values, including defaults, without credentials.
func WriteConfig(path, mode string, cfg any) error {
	fields := []struct{ name, key string }{
		{"ScenarioID", "scenario-id"}, {"Repetition", "repetition"},
		{"BrokerURL", "broker"}, {"BrokerName", "broker-name"}, {"Clients", "clients"},
		{"ConnectRate", "connect-rate"}, {"QoS", "qos"}, {"Rate", "rate"},
		{"Warmup", "warmup"}, {"Duration", "duration"}, {"Drain", "drain"},
		{"PayloadBytes", "payload-bytes"}, {"TopicPrefix", "topic-prefix"},
		{"PublishPattern", "publish-pattern"}, {"Hold", "hold"}, {"OutputRoot", "out"},
		{"Username", "username"}, {"Password", "password"}, {"TLSInsecure", "tls-insecure"},
	}
	var b strings.Builder
	fmt.Fprintf(&b, "tool-version=%s\nmode=%s\n", buildinfo.Version, mode)
	v := reflect.ValueOf(cfg)
	for _, field := range fields {
		f := v.FieldByName(field.name)
		if !f.IsValid() {
			continue
		}
		value := fmt.Sprint(f.Interface())
		if field.name == "Password" && value != "" {
			value = "<redacted>"
		}
		if field.name == "BrokerURL" {
			value = RedactURL(value)
		}
		// Prevent CLI values from injecting extra config lines.
		value = strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(value)
		fmt.Fprintf(&b, "%s=%s\n", field.key, value)
	}
	return os.WriteFile(path, []byte(b.String()), 0600)
}

func WriteSummaryCSV(path string, summary any) error {
	v := reflect.ValueOf(summary)
	var header, values []string
	for i := 0; i < v.NumField(); i++ {
		header = append(header, v.Type().Field(i).Tag.Get("json"))
		field := v.Field(i)
		if field.Kind() == reflect.Struct {
			b, err := json.Marshal(field.Interface())
			if err != nil {
				return err
			}
			values = append(values, string(b))
		} else {
			values = append(values, fmt.Sprint(field.Interface()))
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	err = w.WriteAll([][]string{header, values})
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
