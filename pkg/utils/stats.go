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
	cs := iostream.ColorScheme()
	table := tablewriter.NewWriter(iostream.Out)
	table.SetHeader([]string{"Name", "Duration", "Status", "Operation"})
	// Set table color to green
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgGreenColor},
	)

	var totalDuration time.Duration
	for _, stat := range stats {
		table.Append([]string{cs.Green(stat.Name), cs.Green(stat.Duration.String()), cs.Green(stat.Status), cs.Green(stat.Operation)})
		totalDuration += stat.Duration
	}

	table.Append([]string{cs.Green("Total"), cs.Green(totalDuration.String()), ""})
	table.Render()
}
