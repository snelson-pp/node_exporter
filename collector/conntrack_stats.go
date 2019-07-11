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
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ti-mo/conntrack"
)

type conntractStatsCollector struct {
	conntrack    ConntrackStatsInterface
	insertFailed *prometheus.Desc
}

type ConntrackStatsInterface interface {
	Stats() ([]conntrack.Stats, error)
}

func init() {
	registerCollector("conntrack_stats", true, NewConntrackStatsCollector)
}

func NewConntrackStatsCollector() (Collector, error) {
	conntrack, err := conntrack.Dial(nil)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to conntrack")
	}

	conntractStats := conntractStatsCollector{
		conntrack: conntrack,
		insertFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "insertfailed"),
			"Number of failed conntrack insert calls.",
			[]string{"cpu"}, nil,
		),
	}

	return &conntractStats, nil
}

func (c *conntractStatsCollector) Update(ch chan<- prometheus.Metric) error {
	statsPerCpu, err := c.conntrack.Stats()
	if err != nil {
		return errors.Wrap(err, "error getting conntrack stats")
	}

	for _, stats := range statsPerCpu {
		ch <- prometheus.MustNewConstMetric(c.insertFailed, prometheus.GaugeValue, float64(stats.InsertFailed), string(stats.CPUID))
	}

	return nil
}
