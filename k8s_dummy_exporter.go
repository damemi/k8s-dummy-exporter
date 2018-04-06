/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	flag "github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/runtime"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	kclient "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)

func main() {
	var namespace, podName, metricName, adapterNamespace, adapterServiceName string
	var metricValue int64
	flag.StringVar(&namespace, "namespace", "", "namespace")
	flag.StringVar(&podName, "pod-name", "", "pod name")
	flag.StringVar(&metricName, "metric-name", "foo", "custom metric name")
	flag.Int64Var(&metricValue, "metric-value", 0, "custom metric value")
	flag.StringVar(&adapterNamespace, "adapterNamespace", "custom-metrics", "namespace for custom metric API adapter")
	flag.StringVar(&adapterServiceName, "adapterServiceName", "custom-metrics-apiserver:http", "service name for custom metric API adapter")
	flag.Parse()

	if namespace == "" {
		glog.Fatalf("Namespace required")
	}

	if podName == "" {
		glog.Fatalf("Pod name required")
	}

	clientConfig, err := restclient.InClusterConfig()
	if err != nil {
		glog.Infof("Error creating in-cluster config: %s", err)
	}

	clientConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: Codecs}
	kc := kclient.NewForConfigOrDie(clientConfig)

	servicesProxyRequest := kc.CoreV1().RESTClient().Post().Resource("services").SubResource("proxy")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build custom-metric request
	path := fmt.Sprintf("/write-metrics/namespaces/%s/pods/%s/%s", namespace, podName, metricName)
	glog.Infof("Request path: %s", path)
	body := `{"Value":` + fmt.Sprintf("%d", metricValue) + `}`

	// Make request to custom-metrics service
	req := servicesProxyRequest.Namespace(adapterNamespace).
		Context(ctx).
		Name(adapterServiceName).
		Suffix(path).
		SetHeader("Content-Type", "application/json").
		Body([]byte(body))
	glog.Infof("Request URL: %v", req.URL())
	_, err = req.DoRaw()
	if err != nil {
		glog.Errorf("Error making POST request to custom metrics adapter: %v", err)
	}

	for {
		time.Sleep(5000 * time.Millisecond)
	}
}
