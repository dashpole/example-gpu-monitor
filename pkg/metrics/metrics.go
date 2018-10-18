// Copyright 2018 Google Inc. All Rights Reserved.
//
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

package metrics

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dashpole/example-gpu-monitor/pkg/gpustats"
	"github.com/dashpole/example-gpu-monitor/pkg/kubeletdevices"
)

const nvidiaResourceName = "nvidia.com/gpu"

var (
	ContainerGPUMemoryUsedBytesDesc = prometheus.NewDesc(
		"container_gpu_memory_used_bytes",
		"Total accelerator memory allocated in bytes.",
		[]string{"model", "id", "container_name", "pod_name", "pod_namespace"},
		map[string]string{"make": "nvidia"},
	)

	ContainerGPUMemoryTotalBytes = prometheus.NewDesc(
		"container_gpu_memory_total_bytes",
		"Total accelerator memory in bytes.",
		[]string{"model", "id", "container_name", "pod_name", "pod_namespace"},
		map[string]string{"make": "nvidia"},
	)

	ContainerDutyCycle = prometheus.NewDesc(
		"container_gpu_duty_cycle",
		"Percent of time over the past 10s during which the accelerator was actively processing",
		[]string{"model", "id", "container_name", "pod_name", "pod_namespace"},
		map[string]string{"make": "nvidia"},
	)
)

func NewGPUCollector(statsProvider gpustats.GPUStatsProvider, podProvider kubeletdevices.DeviceProvider) *gpuCollector {
	return &gpuCollector{
		podsProvider:  podProvider,
		statsProvider: statsProvider,
	}
}

type gpuCollector struct {
	podsProvider  kubeletdevices.DeviceProvider
	statsProvider gpustats.GPUStatsProvider
}

func (g *gpuCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- ContainerGPUMemoryUsedBytesDesc
	ch <- ContainerGPUMemoryTotalBytes
	ch <- ContainerDutyCycle
}

func (g *gpuCollector) Collect(ch chan<- prometheus.Metric) {
	podResources, err := g.podsProvider.GetDevices()
	if err != nil {
		glog.Errorf("error getting devices from the kubelet: %v", err)
		return
	}
	glog.Infof("got podResources!: %+v", podResources)
	for _, pod := range podResources.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, device := range container.GetDevices() {
				if device.GetResourceName() == nvidiaResourceName {
					for _, id := range device.GetDeviceIds() {
						stats, err := g.statsProvider.GetStats(id)
						if err != nil {
							glog.Warningf("error getting stats for device with id %s: %v", id, err)
							continue
						}
						ch <- prometheus.MustNewConstMetric(
							ContainerGPUMemoryUsedBytesDesc,
							prometheus.GaugeValue,
							float64(stats.MemoryUsed),
							stats.Model,
							stats.ID,
							container.GetName(),
							pod.GetName(),
							pod.GetNamespace(),
						)
						ch <- prometheus.MustNewConstMetric(
							ContainerGPUMemoryTotalBytes,
							prometheus.GaugeValue,
							float64(stats.MemoryTotal),
							stats.Model,
							stats.ID,
							container.GetName(),
							pod.GetName(),
							pod.GetNamespace(),
						)
						ch <- prometheus.MustNewConstMetric(
							ContainerDutyCycle,
							prometheus.GaugeValue,
							float64(stats.DutyCycle),
							stats.Model,
							stats.ID,
							container.GetName(),
							pod.GetName(),
							pod.GetNamespace(),
						)
					}
				}
			}
		}
	}
}
