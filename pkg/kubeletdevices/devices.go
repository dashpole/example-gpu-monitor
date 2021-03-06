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

package kubeletdevices

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc"

	podresourcesapi "k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/util"
)

const (
	defaultPodResourcesTimeout = 10 * time.Second
	defaultPodResourcesMaxSize = 1024 * 1024 * 16 // 16 Mb
)

type DeviceProvider interface {
	GetDevices() (*podresourcesapi.ListPodResourcesResponse, error)
}

type deviceProvider struct {
	client podresourcesapi.PodResourcesListerClient
}

func NewDeviceProvider(socket string) DeviceProvider {
	addr, dailer, err := util.GetAddressAndDialer(socket)
	if err != nil {
		glog.Fatalf("Error getting address and dialer for socket %s: %v", socket, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultPodResourcesTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithDialer(dailer), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(defaultPodResourcesMaxSize)))
	if err != nil {
		glog.Fatalf("Error dialing socket %s: %v", socket, err)
	}
	return &deviceProvider{
		client: podresourcesapi.NewPodResourcesListerClient(conn),
	}
}

func (d *deviceProvider) GetDevices() (*podresourcesapi.ListPodResourcesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := d.client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("%v.Get(_) = _, %v", d.client, err)
	}
	return resp, nil
}
