// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
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

package sources

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// string mappings for 'health_check' values
	healthCheckHealthy   string = "1"
	healthCheckUnhealthy string = "0"
)

var (
	// HealthStatusEnabled specifies whether to collect Health metrics
	HealthStatusEnabled string
)

func init() {
	Factories["sysfs"] = newLustreSysSource
}

type lustreSysSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func (s *lustreSysSource) generateHealthStatusTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"": {
			{"health_check", "health_check", "Current health status for the indicated instance: " + healthCheckHealthy + " refers to 'healthy', " + healthCheckUnhealthy + " refers to 'unhealthy'", gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "health", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *lustreSysSource) generateOSTMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"obdfilter/*-OST*": {
			{"degraded", "degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded", gaugeMetric, false, core},
			{"grant_precreate", "grant_precreate_capacity_bytes", "Maximum space in bytes that clients can preallocate for objects", gaugeMetric, false, extended},
			{"lfsck_speed_limit", "lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run", gaugeMetric, false, extended},
			{"precreate_batch", "precreate_batch", "Maximum number of objects that can be included in a single transaction", gaugeMetric, false, extended},
			{"soft_sync_limit", "soft_sync_limit", "Number of RPCs necessary before triggering a sync", gaugeMetric, false, extended},
			{"sync_journal", "sync_journal_enabled", "Binary indicator as to whether or not the journal is set for asynchronous commits", gaugeMetric, false, extended},
		},
		"ldlm/namespaces/filter-*": {
			{"lock_count", "lock_count", "Number of locks", gaugeMetric, false, extended},
			{"lock_timeouts", "lock_timeout", "Number of lock timeouts", counterMetric, false, extended},
			{"contended_locks", "lock_contended", "Number of contended locks", gaugeMetric, false, extended},
			{"contention_seconds", "lock_contention_seconds", "Time in seconds during which locks were contended", gaugeMetric, false, extended},

			{"pool/granted", "lock_granted", "Number of granted locks", gaugeMetric, false, extended},
			{"pool/grant_plan", "lock_grant_plan", "Number of planned lock grants per second", gaugeMetric, false, extended},
			{"pool/grant_rate", "lock_grant_rate", "Lock grant rate", gaugeMetric, false, extended},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "ost", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func newLustreSysSource() LustreSource {
	var l lustreSysSource
	l.basePath = filepath.Join(SysLocation, "fs/lustre")
	if HealthStatusEnabled != disabled {
		l.generateHealthStatusTemplates(HealthStatusEnabled)
	}
	if OstEnabled != disabled {
		l.generateOSTMetricTemplates(OstEnabled)
	}
	return &l
}

func (s *lustreSysSource) Update(ch chan<- prometheus.Metric) (err error) {
	var directoryDepth int

	for _, metric := range s.lustreProcMetrics {
		directoryDepth = strings.Count(metric.filename, "/")
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			switch metric.filename {
			case "health_check":
				err = s.parseTextFile(metric.source, "health_check", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
					ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			default:
				err = s.parseFile(metric.source, single, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{"component", "target", extraLabel}, []string{nodeType, nodeName, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *lustreSysSource) parseTextFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	filename, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	fileString := string(fileBytes[:])
	switch filename {
	case "health_check":
		if strings.TrimSpace(fileString) == "healthy" {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckHealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		} else {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckUnhealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		}
	}
	return nil
}

func (s *lustreSysSource) parseFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	switch metricType {
	case single:
		value, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, promName, helpText, convertedValue, "", "")
	}
	return nil
}
