package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// resolveArgs translates the external configuration into ordinary CLI flags so
// both entry points share defaults, type checking and benchmark validation.
func resolveArgs(args []string) ([]string, error) {
	if len(args) == 0 {
		return readConfig("mqttbench.conf")
	}
	if args[0] != "--config" && !strings.HasPrefix(args[0], "--config=") {
		return args, nil
	}
	var path string
	var overrides []string
	if args[0] == "--config" {
		if len(args) < 2 || strings.HasPrefix(args[1], "--") {
			return nil, fmt.Errorf("--config requires a file path")
		}
		path, overrides = args[1], args[2:]
	} else {
		path, overrides = strings.TrimPrefix(args[0], "--config="), args[1:]
	}
	if path == "" {
		return nil, fmt.Errorf("--config requires a file path")
	}
	configured, err := readConfig(path)
	if err != nil {
		return nil, err
	}
	return append(configured, overrides...), nil
}

func readConfig(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read configuration %q: %w", path, err)
	}
	defer f.Close()
	mode := "message"
	var flags []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		if line == 1 {
			text = strings.TrimPrefix(text, "\ufeff")
		}
		text = strings.TrimSpace(text)
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		key, value, ok := strings.Cut(text, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || strings.HasPrefix(key, "-") || strings.ContainsAny(key, " \t") {
			return nil, fmt.Errorf("%s:%d: expected field=value (without leading --)", path, line)
		}
		if seen[key] {
			return nil, fmt.Errorf("%s:%d: duplicate field %q", path, line, key)
		}
		seen[key] = true
		if key == "mode" {
			mode = value
		} else {
			flags = append(flags, "--"+key+"="+value)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read configuration %q: %w", path, err)
	}
	if mode != "message" && mode != "connections" {
		return nil, fmt.Errorf("%s: mode must be message or connections", path)
	}
	return append([]string{mode}, flags...), nil
}
