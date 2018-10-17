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

package main

import (
	"flag"
	"github.com/golang/glog"

	"github.com/dashpole/example-gpu-monitor/pkg/gpustats"
	"github.com/dashpole/example-gpu-monitor/pkg/kubeletdevices"
	"github.com/dashpole/example-gpu-monitor/pkg/metrics"
)

var (
	socket             = flag.String("socket", "", "location of the kubelet's podresources service")
	port               = flag.Int("port", 8080, "port on which to listen")
	prometheusEndpoint = flag.String("prometheus_endpoint", "/metrics", "Endpoint to expose Prometheus metrics on")
)

func main() {
	defer glog.Flush()
	flag.Parse()

	glog.V(1).Infof("Starting example-gpu-monitor")

	metrics.Register(gpustats.NewGPUStatsProvider(), kubeletdevices.NewDeviceProvider())

}
