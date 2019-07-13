// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !noconntrackstats

package collector

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	CONNTRACK_BIN      = "/usr/sbin/conntrack"
	INSERT_FAILED_STAT = "insert_failed"
)

type conntractStatsCollector struct {
	insertFailed *prometheus.Desc
}

func init() {
	registerCollector("conntrack_stats", defaultEnabled, NewConntrackStatsCollector)
}

func NewConntrackStatsCollector() (Collector, error) {

	conntractStats := conntractStatsCollector{
		insertFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "insertfailed"),
			"Number of failed conntrack insert calls.",
			[]string{"cpu"}, nil,
		),
	}

	return &conntractStats, nil
}

func (c *conntractStatsCollector) Update(ch chan<- prometheus.Metric) error {
	binExists, err := conntrackBinExists()
	if err != nil {
		log.Warnf("Error checking if %q exists: %v; skipping conntrack stats gathering", CONNTRACK_BIN, err)
		return nil
	}

	if !binExists {
		log.Infof("Binary %q does not exist; skipping conntrack stats gathering", CONNTRACK_BIN)
		return nil
	}

	conntrackCmd := exec.Command(CONNTRACK_BIN, "-S")
	var stdout bytes.Buffer
	conntrackCmd.Stdout = &stdout
	err = conntrackCmd.Run()
	if err != nil {
		log.Errorf("Error running %q: %v", CONNTRACK_BIN, err)
		return err
	}

	cpuStats, err := ConntrackStatsParseReader(&stdout)
	if err != nil {
		log.Errorf("Error parsing conntrack stream: %v", err)
		return err
	}

	for _, cpuStat := range cpuStats {
		insertFailedValue, ok := cpuStat.Stats[INSERT_FAILED_STAT]
		if !ok {
			log.Warnf("required stat %q missing from cpuid %v", INSERT_FAILED_STAT, cpuStat.Cpu)
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.insertFailed, prometheus.GaugeValue, float64(insertFailedValue), string(cpuStat.Cpu))
	}

	return nil
}

func ConntrackStatsParseReader(reader io.Reader) ([]CpuStat, error) {
	var allStats []CpuStat

	conntrackScanner := bufio.NewScanner(reader)

	for conntrackScanner.Scan() {
		line := conntrackScanner.Text()

		// Skip blank lines
		if len(line) == 0 {
			continue
		}

		cpuStats, err := ConntrackStatsParseLine(line)
		if err != nil {
			log.Warnf("error parsing line %q: %v", line, err)
			continue
		}

		allStats = append(allStats, *cpuStats)
	}

	if err := conntrackScanner.Err(); err != nil {
		log.Errorf("error getting output from %q: %v", CONNTRACK_BIN, err)
		return allStats, err
	}

	return allStats, nil
}

func conntrackBinExists() (bool, error) {
	_, err := os.Stat(CONNTRACK_BIN)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, errors.Wrapf(err, "error checking for %q", CONNTRACK_BIN)
		}
	}

	return true, nil

}

type CpuStat struct {
	Cpu   int
	Stats map[string]int
}

func ConntrackStatsParseLine(line string) (*CpuStat, error) {
	cpuStat := new(CpuStat)

	split := strings.SplitN(line, " ", 2)

	cpuidStr, valuesStr := split[0], split[1]
	cpuidStr = cpuidStr[4:]
	cpuid, err := strconv.Atoi(cpuidStr)
	if err != nil {
		return nil, errors.Wrapf(err, "error converting cpuid %q to int: %v", cpuidStr, err)
	}
	cpuStat.Cpu = cpuid

	valuesStr = strings.Trim(valuesStr, " ")
	valuePairs := strings.Split(valuesStr, " ")
	cpuStat.Stats = map[string]int{}
	for _, pair := range valuePairs {
		splitPairs := strings.SplitN(pair, "=", 2)
		statName := splitPairs[0]
		statValueStr := splitPairs[1]
		statValue, err := strconv.Atoi(statValueStr)
		if err != nil {
			return nil, errors.Wrapf(err, "error converting %q to int: %v", statValueStr, err)
		}
		cpuStat.Stats[statName] = statValue
	}

	return cpuStat, nil
}
