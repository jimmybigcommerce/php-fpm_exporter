// Copyright © 2018 Enrico Stahn <enrico.stahn@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package phpfpm provides convenient access to PHP-FPM pool data
package phpfpm

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	hashids "github.com/speps/go-hashids"
)

const (
	namespace = "phpfpm"
)

// Exporter configures and exposes PHP-FPM metrics to Prometheus.
type Exporter struct {
	mutex       sync.Mutex
	PoolManager PoolManager

	CountProcessState            bool
	AggregateProcessLevelMetrics bool

	up                                *prometheus.Desc
	scrapeFailues                     *prometheus.Desc
	startSince                        *prometheus.Desc
	acceptedConnections               *prometheus.Desc
	listenQueue                       *prometheus.Desc
	maxListenQueue                    *prometheus.Desc
	listenQueueLength                 *prometheus.Desc
	idleProcesses                     *prometheus.Desc
	activeProcesses                   *prometheus.Desc
	totalProcesses                    *prometheus.Desc
	maxActiveProcesses                *prometheus.Desc
	maxChildrenReached                *prometheus.Desc
	slowRequests                      *prometheus.Desc
	processRequests                   *prometheus.Desc
	processLastRequestMemory          *prometheus.Desc
	processLastRequestCPU             *prometheus.Desc
	processRequestDuration            *prometheus.Desc
	processState                      *prometheus.Desc
	processTotalRequests              *prometheus.Desc
	processAggregateRequests          *prometheus.Desc
	processAggregateLastRequestMemory *prometheus.Desc
	processAggregateLastRequestCPU    *prometheus.Desc
	processAggregateRequestDuration   *prometheus.Desc
	processAggregateState             *prometheus.Desc
}

// NewExporter creates a new Exporter for a PoolManager and configures the necessary metrics.
func NewExporter(pm PoolManager) *Exporter {
	return &Exporter{
		PoolManager: pm,

		CountProcessState:            false,
		AggregateProcessLevelMetrics: false,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could PHP-FPM be reached?",
			[]string{"pool"},
			nil),

		scrapeFailues: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_failures"),
			"The number of failures scraping from PHP-FPM.",
			[]string{"pool"},
			nil),

		startSince: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "start_since"),
			"The number of seconds since FPM has started.",
			[]string{"pool"},
			nil),

		acceptedConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "accepted_connections"),
			"The number of requests accepted by the pool.",
			[]string{"pool"},
			nil),

		listenQueue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "listen_queue"),
			"The number of requests in the queue of pending connections.",
			[]string{"pool"},
			nil),

		maxListenQueue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_listen_queue"),
			"The maximum number of requests in the queue of pending connections since FPM has started.",
			[]string{"pool"},
			nil),

		listenQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "listen_queue_length"),
			"The size of the socket queue of pending connections.",
			[]string{"pool"},
			nil),

		idleProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "idle_processes"),
			"The number of idle processes.",
			[]string{"pool"},
			nil),

		activeProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "active_processes"),
			"The number of active processes.",
			[]string{"pool"},
			nil),

		totalProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_processes"),
			"The number of idle + active processes.",
			[]string{"pool"},
			nil),

		maxActiveProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_active_processes"),
			"The maximum number of active processes since FPM has started.",
			[]string{"pool"},
			nil),

		maxChildrenReached: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_children_reached"),
			"The number of times, the process limit has been reached, when pm tries to start more children (works only for pm 'dynamic' and 'ondemand').",
			[]string{"pool"},
			nil),

		slowRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "slow_requests"),
			"The number of requests that exceeded your 'request_slowlog_timeout' value.",
			[]string{"pool"},
			nil),

		processRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_requests"),
			"The number of requests the process has served.",
			[]string{"pool", "pid_hash"},
			nil),

		processLastRequestMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_memory"),
			"The max amount of memory the last request consumed.",
			[]string{"pool", "pid_hash"},
			nil),

		processLastRequestCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_cpu"),
			"The %cpu the last request consumed.",
			[]string{"pool", "pid_hash"},
			nil),

		processRequestDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_request_duration"),
			"The duration in microseconds of the requests.",
			[]string{"pool", "pid_hash"},
			nil),

		processState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_state"),
			"The state of the process (Idle, Running, ...).",
			[]string{"pool", "pid_hash", "state"},
			nil),

		processTotalRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_requests_total"),
			"The number of requests the process has served.",
			[]string{"pool"},
			nil),

		processAggregateRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_requests_average"),
			"The avg number of requests per process.",
			[]string{"pool"},
			nil),

		processAggregateLastRequestMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_request_memory_average"),
			"The avg amount of memory the last request consumed.",
			[]string{"pool"},
			nil),

		processAggregateLastRequestCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_request_cpu_average"),
			"The avg %cpu of last requests.",
			[]string{"pool"},
			nil),

		processAggregateRequestDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_request_duration_average"),
			"The avg duration in microseconds of the requests.",
			[]string{"pool"},
			nil),

		processAggregateState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_state_totals"),
			"The total count for each request state in the pool (Idle, Running, ...).",
			[]string{"pool", "state"},
			nil),
	}
}

// Collect updates the Pools and sends the collected metrics to Prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.PoolManager.Update()

	for _, pool := range e.PoolManager.Pools {
		ch <- prometheus.MustNewConstMetric(e.scrapeFailues, prometheus.CounterValue, float64(pool.ScrapeFailures), pool.Name)

		if pool.ScrapeError != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0, pool.Name)
			log.Errorf("Error scraping PHP-FPM: %v", pool.ScrapeError)
			continue
		}

		active, idle, total := CountProcessState(pool.Processes)
		if !e.CountProcessState && (active != pool.ActiveProcesses || idle != pool.IdleProcesses) {
			log.Error("Inconsistent active and idle processes reported. Set `--fix-process-count` to have this calculated by php-fpm_exporter instead.")
		}

		if !e.CountProcessState {
			active = pool.ActiveProcesses
			idle = pool.IdleProcesses
			total = pool.TotalProcesses
		}

		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1, pool.Name)
		ch <- prometheus.MustNewConstMetric(e.startSince, prometheus.CounterValue, float64(pool.StartSince), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.acceptedConnections, prometheus.CounterValue, float64(pool.AcceptedConnections), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.listenQueue, prometheus.GaugeValue, float64(pool.ListenQueue), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxListenQueue, prometheus.CounterValue, float64(pool.MaxListenQueue), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.listenQueueLength, prometheus.GaugeValue, float64(pool.ListenQueueLength), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.idleProcesses, prometheus.GaugeValue, float64(idle), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.activeProcesses, prometheus.GaugeValue, float64(active), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.totalProcesses, prometheus.GaugeValue, float64(total), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxActiveProcesses, prometheus.CounterValue, float64(pool.MaxActiveProcesses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxChildrenReached, prometheus.CounterValue, float64(pool.MaxChildrenReached), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.slowRequests, prometheus.CounterValue, float64(pool.SlowRequests), pool.Name)

		if !e.AggregateProcessLevelMetrics {
			for _, process := range pool.Processes {
				pidHash := calculateProcessHash(process)
				ch <- prometheus.MustNewConstMetric(e.processState, prometheus.GaugeValue, 1, pool.Name, pidHash, process.State)
				ch <- prometheus.MustNewConstMetric(e.processRequests, prometheus.CounterValue, float64(process.Requests), pool.Name, pidHash)
				ch <- prometheus.MustNewConstMetric(e.processLastRequestMemory, prometheus.GaugeValue, float64(process.LastRequestMemory), pool.Name, pidHash)
				ch <- prometheus.MustNewConstMetric(e.processLastRequestCPU, prometheus.GaugeValue, process.LastRequestCPU, pool.Name, pidHash)
				ch <- prometheus.MustNewConstMetric(e.processRequestDuration, prometheus.GaugeValue, float64(process.RequestDuration), pool.Name, pidHash)
			}
		} else {
			aggregateProcessState(pool, ch, e)
			processAggregateLastRequestCPU(pool, ch, e)
			processAggregateLastRequestMemory(pool, ch, e)
			processAggregateRequestDuration(pool, ch, e)
			processAggregateRequests(pool, ch, e)
		}
	}
}

func aggregateProcessState(pool Pool, ch chan<- prometheus.Metric, e *Exporter) {
	var m = make(map[string]int)
	for _, process := range pool.Processes {
		m[process.State]++
	}
	for k, v := range m {
		ch <- prometheus.MustNewConstMetric(e.processAggregateState, prometheus.CounterValue, float64(v), pool.Name, k)
	}
}

func processAggregateLastRequestCPU(pool Pool, ch chan<- prometheus.Metric, e *Exporter) {
	var total float64
	var count float64
	for _, process := range pool.Processes {
		print(process.LastRequestCPU)
		print(" ... ")
		total += process.LastRequestCPU
		count++
	}
	ch <- prometheus.MustNewConstMetric(e.processAggregateLastRequestCPU, prometheus.CounterValue, total/count, pool.Name)
}

func processAggregateLastRequestMemory(pool Pool, ch chan<- prometheus.Metric, e *Exporter) {
	var total int64
	var count int64
	for _, process := range pool.Processes {
		total += process.LastRequestMemory
		count++
	}
	ch <- prometheus.MustNewConstMetric(e.processAggregateLastRequestMemory, prometheus.CounterValue, float64(total/count), pool.Name)
}

func processAggregateRequestDuration(pool Pool, ch chan<- prometheus.Metric, e *Exporter) {
	var total float64
	var count float64
	for _, process := range pool.Processes {
		total += float64(process.RequestDuration)
		count++
	}
	ch <- prometheus.MustNewConstMetric(e.processAggregateRequestDuration, prometheus.CounterValue, float64(total/count), pool.Name)
}

func processAggregateRequests(pool Pool, ch chan<- prometheus.Metric, e *Exporter) {
	var total float64
	var count float64
	for _, process := range pool.Processes {
		total += float64(process.Requests)
		count++
	}
	ch <- prometheus.MustNewConstMetric(e.processAggregateRequests, prometheus.CounterValue, float64(total/count), pool.Name)
	ch <- prometheus.MustNewConstMetric(e.processTotalRequests, prometheus.CounterValue, float64(total), pool.Name)
}

// Describe exposes the metric description to Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.startSince
	ch <- e.acceptedConnections
	ch <- e.listenQueue
	ch <- e.maxListenQueue
	ch <- e.listenQueueLength
	ch <- e.idleProcesses
	ch <- e.activeProcesses
	ch <- e.totalProcesses
	ch <- e.maxActiveProcesses
	ch <- e.maxChildrenReached
	ch <- e.slowRequests
	ch <- e.processState
	ch <- e.processRequests
	ch <- e.processLastRequestMemory
	ch <- e.processLastRequestCPU
	ch <- e.processRequestDuration
}

// calculateProcessHash generates a unique identifier for a process to ensure uniqueness across multiple systems/containers
func calculateProcessHash(pp PoolProcess) string {
	hd := hashids.NewData()
	hd.Salt = "php-fpm_exporter"
	hd.MinLength = 12
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{int(pp.StartTime), int(pp.PID)})

	return e
}
