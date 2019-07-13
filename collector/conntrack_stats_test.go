package collector

import (
	"strings"
	"testing"
)

func TestConntrackStatsParseLine(t *testing.T) {
	line := "cpu=0   	found=39 invalid=7 ignore=21270 insert=0 insert_failed=141 drop=141 early_drop=0 error=0 search_restart=5350"

	cpuStat, err := ConntrackStatsParseLine(line)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cpuStat.Cpu != 0 {
		t.Errorf("Expected cpuid %v, got %v", 0, cpuStat.Cpu)
	}

	if insertFailed := cpuStat.Stats["insert_failed"]; insertFailed != 141 {
		t.Errorf("Expected insert_failed %v, got %v", 141, insertFailed)
	}

}

func TestConntrackStatsParseReader(t *testing.T) {
	input := `cpu=0   	found=46 invalid=7 ignore=23270 insert=0 insert_failed=159 drop=159 early_drop=0 error=0 search_restart=5843
cpu=1   	found=376 invalid=11 ignore=38027 insert=0 insert_failed=978 drop=978 early_drop=0 error=0 search_restart=13965
cpu=2   	found=362 invalid=29 ignore=49317 insert=0 insert_failed=943 drop=943 early_drop=0 error=0 search_restart=23222
cpu=3   	found=603 invalid=16 ignore=38484 insert=0 insert_failed=1599 drop=1599 early_drop=0 error=0 search_restart=15615`

	inputReader := strings.NewReader(input)

	cpuStats, err := ConntrackStatsParseReader(inputReader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedStats := []int{159, 978, 943, 1599}

	for cpuId, expectedInsertFailed := range expectedStats {
		cpuStat := cpuStats[cpuId]

		if cpuStat.Cpu != cpuId {
			t.Errorf("Expected cpuid %v, got %v", cpuId, cpuStat.Cpu)
		}

		if insertFailed := cpuStat.Stats["insert_failed"]; insertFailed != expectedInsertFailed {
			t.Errorf("Expected insert_failed %v, got %v", expectedInsertFailed, insertFailed)
		}
	}

}
