// Command latency measures the memory application handler, not real HTTP/S1 load.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/query"
)

type latency struct {
	P50 int64   `json:"p50_ns"`
	P95 int64   `json:"p95_ns"`
	P99 int64   `json:"p99_ns"`
	Raw []int64 `json:"raw_ns,omitempty"`
}

func summarize(samples []int64, raw bool) latency {
	ordered := append([]int64(nil), samples...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	at := func(p int) int64 { return ordered[(p*len(ordered)+99)/100-1] }
	result := latency{P50: at(50), P95: at(95), P99: at(99)}
	if raw {
		result.Raw = samples
	}
	return result
}
func main() {
	count := flag.Int("samples", 10000, "unique isolated-account samples, then identical replays (1..100000)")
	raw := flag.Bool("raw", false, "include raw chronological latency samples")
	flag.Parse()
	if *count < 1 || *count > 100000 {
		fmt.Fprintln(os.Stderr, "samples must be 1..100000")
		os.Exit(2)
	}
	h := query.NewScoreEventHandler()
	ctx := context.Background()
	samples := map[string][]int64{"unique": make([]int64, *count), "replay": make([]int64, *count)}
	for _, mode := range []string{"unique", "replay"} {
		for i := 0; i < *count; i++ {
			q := query.ScoreEventQuery{EventID: fmt.Sprintf("event-%d", i), Account: fmt.Sprintf("account-%d", i), AmountCents: 100}
			start := time.Now()
			d, err := h.Handle(ctx, q)
			elapsed := time.Since(start).Nanoseconds()
			if err != nil || d.Score != 0 || d.Decision != "ALLOW" {
				fmt.Fprintln(os.Stderr, "unexpected scoring result", err)
				os.Exit(1)
			}
			samples[mode][i] = elapsed
		}
	}
	unique, replay := summarize(samples["unique"], *raw), summarize(samples["replay"], *raw)
	_ = json.NewEncoder(os.Stdout).Encode(struct {
		Mode      string  `json:"mode"`
		GoVersion string  `json:"go_version"`
		OS        string  `json:"os"`
		Arch      string  `json:"arch"`
		Procs     int     `json:"gomaxprocs"`
		Samples   int     `json:"samples_per_cohort"`
		Unique    latency `json:"unique"`
		Replay    latency `json:"replay"`
	}{"memory application handler; not HTTP or real-process S1", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), *count, unique, replay})
	if unique.P99 >= int64(5*time.Millisecond) || replay.P99 >= int64(5*time.Millisecond) {
		os.Exit(1)
	}
}
