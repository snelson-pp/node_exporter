package collector

import "testing"

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
