package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bench.conf")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("\ufeff# comentário\r\nmode=message\r\nclients = 5\r\npassword=a=b#c\r\nwarmup=0s\r\n")
	got, err := resolveArgs([]string{"--config", path, "--clients=7"})
	want := []string{"message", "--clients=5", "--password=a=b#c", "--warmup=0s", "--clients=7"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v; want %v", got, err, want)
	}
	for _, body := range []string{"clients 5", "clients=5\nclients=6", "mode=invalid", "--clients=5"} {
		write(body)
		if _, err := readConfig(path); err == nil {
			t.Errorf("accepted invalid config %q", body)
		}
	}
	write("mode=connections\nhold=2s\n")
	got, err = resolveArgs([]string{"--config=" + path})
	if err != nil || !reflect.DeepEqual(got, []string{"connections", "--hold=2s"}) {
		t.Fatalf("connections: %v, %v", got, err)
	}
	for _, body := range []string{"unknown=1", "duration=bad", "clients=bad", "tls-insecure=bad", "qos=3"} {
		write(body)
		args, err := readConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := runMessage(context.Background(), append(args[1:], "--broker=mqtt://localhost:1883")); err == nil {
			t.Errorf("accepted %q", body)
		}
	}
}

func TestLegacyArgs(t *testing.T) {
	want := []string{"message", "--clients", "5"}
	got, err := resolveArgs(want)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("%v, %v", got, err)
	}
	if _, err := resolveArgs([]string{"--config"}); err == nil {
		t.Fatal("accepted missing path")
	}
	if _, err := readConfig(filepath.Join(t.TempDir(), "missing.conf")); err == nil {
		t.Fatal("accepted missing file")
	}
}

func TestDefaultConfig(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mqttbench.conf"), []byte("mode=message\nscenario-id=DEFAULT\nrepetition=1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(original); err != nil {
			t.Error(err)
		}
	}()
	got, err := resolveArgs(nil)
	want := []string{"message", "--scenario-id=DEFAULT", "--repetition=1"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v", got, err)
	}
}
