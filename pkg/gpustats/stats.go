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

package gpustats

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/mindprince/gonvml"
)

type GPUStatsProvider interface {
	GetStats(deviceID string) (*Stats, error)
	Stop()
}

type Stats struct {
	Model       string
	ID          string
	MemoryTotal uint64
	MemoryUsed  uint64
	DutyCycle   uint64
}

func NewGPUStatsProvider() GPUStatsProvider {
	m := monitorImpl{
		devices: make(map[int]gonvml.Device),
	}
	if err := gonvml.Initialize(); err != nil {
		// This is under a logging level because otherwise we may cause
		// log spam if the drivers/nvml is not installed on the system.
		glog.V(4).Infof("Could not initialize NVML: %v", err)
		return &m
	}
	m.nvmlInitialized = true
	numDevices, err := gonvml.DeviceCount()
	if err != nil {
		glog.Warningf("GPU metrics would not be available. Failed to get the number of nvidia devices: %v", err)
		return &m
	}
	glog.V(1).Infof("NVML initialized. Number of nvidia devices: %v", numDevices)
	m.devices = make(map[int]gonvml.Device, numDevices)
	for i := 0; i < int(numDevices); i++ {
		device, err := gonvml.DeviceHandleByIndex(uint(i))
		if err != nil {
			glog.Warningf("Failed to get nvidia device handle %d: %v", i, err)
			continue
		}
		minorNumber, err := device.MinorNumber()
		if err != nil {
			glog.Warningf("Failed to get nvidia device minor number: %v", err)
			continue
		}
		m.devices[int(minorNumber)] = device
	}
	return &m
}

type monitorImpl struct {
	nvmlInitialized bool

	// map from device minor number to Device
	devices map[int]gonvml.Device
}

func (m *monitorImpl) Stop() {
	if m.nvmlInitialized {
		gonvml.Shutdown()
	}
}

// GetStats assumes the device id is nvidia[minor number]
func (m *monitorImpl) GetStats(deviceId string) (*Stats, error) {
	i, err := getMinorNumber(deviceId)
	if err != nil {
		return nil, fmt.Errorf("error getting device minor number from path %s: %v", deviceId, err)
	}
	device, found := m.devices[i]
	if !found {
		return nil, fmt.Errorf("device with minor number %d was not found", i)
	}
	model, err := device.Name()
	if err != nil {
		return nil, fmt.Errorf("error while getting gpu name: %v", err)
	}
	uuid, err := device.UUID()
	if err != nil {
		return nil, fmt.Errorf("error while getting gpu uuid: %v", err)
	}
	memoryTotal, memoryUsed, err := device.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("error while getting gpu memory info: %v", err)
	}
	utilizationGPU, err := device.AverageGPUUtilization(10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("error while getting gpu utilization: %v", err)
	}

	return &Stats{
		Model:       model,
		ID:          uuid,
		MemoryTotal: memoryTotal,
		MemoryUsed:  memoryUsed,
		DutyCycle:   uint64(utilizationGPU),
	}, nil
}

var deviceExpr = regexp.MustCompile(`^nvidia([0-9]+)$`)

func getMinorNumber(deviceId string) (int, error) {
	matches := deviceExpr.FindStringSubmatch(deviceId)
	if len(matches) != 2 {
		return 0, fmt.Errorf("%s does not match nvidia[0-9]+", deviceId)
	}
	minorString := matches[1]
	i, err := strconv.ParseInt(minorString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %s, to an int: %v", minorString, err)
	}
	return int(i), nil
}
