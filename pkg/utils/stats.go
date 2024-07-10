package utils

import (
	"sync"
	"time"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/olekukonko/tablewriter"
)

type Stats struct {
	Name      string
	Duration  time.Duration
	Status    string
	Operation string
}

type StatsCollector struct {
	stats []*Stats
	mu    sync.Mutex
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		stats: make([]*Stats, 0),
	}
}

func (sc *StatsCollector) AddStat(stat *Stats) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.stats = append(sc.stats, stat)
}

func (sc *StatsCollector) GetStats() []*Stats {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.stats
}

func PrintCombinedStats(iostream *iostreams.IOStreams, stats []*Stats) {
	table := tablewriter.NewWriter(iostream.Out)
	table.SetHeader([]string{"Name", "Duration", "Status", "Operation"})

	var totalDuration time.Duration
	for _, stat := range stats {
		table.Append([]string{stat.Name, stat.Duration.String(), stat.Status, stat.Operation})
		totalDuration += stat.Duration
	}

	table.Append([]string{"Total", totalDuration.String(), ""})
	table.Render()
}
