package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

type AllocationMetrics struct {
	num_gpus float64
}

func ParseAllocationMetrics(input []byte) *AllocationMetrics {
	var metrics AllocationMetrics
	num_gpus := 0

	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gpu:1") {
			num_gpus += 1
		}
		if strings.Contains(line, "gpu:2") {
			num_gpus += 2
		}
		if strings.Contains(line, "gpu:3") {
			num_gpus += 3
		}
		if strings.Contains(line, "gpu:4") {
			num_gpus += 4
		}
	}
	metrics.num_gpus = float64(num_gpus)
	return &metrics
}

// Returns the scheduler metrics
func AllocationGetMetrics() *AllocationMetrics {
	return ParseAllocationMetrics(AllocationData())
}

// Execute the squeue command and return its output
func AllocationData() []byte {
	cmd := exec.Command("/usr/bin/sacct", "--allusers", "-X", "--states=r", "--format=reqgres")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	out, _ := ioutil.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return out
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm queue metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewAllocationCollector() *AllocationCollector {
	return &AllocationCollector{
		num_gpus: prometheus.NewDesc("num_gpus", "Number of GPUs currently in use", nil, nil),
	}
}

type AllocationCollector struct {
	num_gpus *prometheus.Desc
}

func (qc *AllocationCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- qc.num_gpus
}

func (qc *AllocationCollector) Collect(ch chan<- prometheus.Metric) {
	qm := AllocationGetMetrics()
	ch <- prometheus.MustNewConstMetric(qc.num_gpus, prometheus.GaugeValue, qm.num_gpus)
}
