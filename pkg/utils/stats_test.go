package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatsCollector(t *testing.T) {
	sc := NewStatsCollector()

	stat1 := &Stats{Name: "Test1", Duration: 1 * time.Second, Status: "success", Operation: "install"}
	stat2 := &Stats{Name: "Test2", Duration: 2 * time.Second, Status: "failed", Operation: "install"}

	sc.AddStat(stat1)
	sc.AddStat(stat2)

	stats := sc.GetStats()

	assert.Len(t, stats, 2)
	assert.Equal(t, stat1, stats[0])
	assert.Equal(t, stat2, stats[1])
}
