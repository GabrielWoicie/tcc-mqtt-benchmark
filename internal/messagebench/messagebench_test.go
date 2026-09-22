package messagebench

import (
	"github.com/eclipse/paho.golang/paho"
	"testing"
	"time"
)

func TestOffsetsAndHalfOpenWindow(t *testing.T) {
	for _, pattern := range []string{"staggered", "synchronized"} {
		for _, rate := range []float64{1, 5} {
			for _, warmup := range []time.Duration{0, 5 * time.Second, 5500 * time.Millisecond} {
				cfg := Config{Clients: 5, Rate: rate, PublishPattern: pattern}
				count := 0
				for i := 0; i < cfg.Clients; i++ {
					offset := publishOffset(cfg, i)
					want := time.Duration(0)
					if pattern == "staggered" {
						want = time.Duration(float64(time.Second) * float64(i) / (rate * 5))
					}
					if offset != want {
						t.Fatalf("offset %s, want %s", offset, want)
					}
					for slot := offset; slot < warmup+15*time.Second; slot += rateInterval(rate) {
						if slot >= warmup {
							count++
						}
					}
				}
				if count != int(75*rate) {
					t.Fatalf("%s rate=%v warmup=%s: %d attempts", pattern, rate, warmup, count)
				}
			}
		}
	}
}

func TestCollectorExcludesWarmup(t *testing.T) {
	c := &collector{runID: "test", seen: make(map[string]struct{})}
	for _, phase := range []string{"warmup", "measure"} {
		body := makePayload(256, payload{RunID: "test", Phase: phase, Sequence: 1, SentAtNS: time.Now().UnixNano()})
		c.onPublish(paho.PublishReceived{Packet: &paho.Publish{Payload: body}})
	}
	if len(c.records) != 1 || c.duplicates != 0 {
		t.Fatalf("warmup affected records: %+v", c)
	}
}
