/*
Copyright © 2018 inwinSTACK Inc

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
	goflag "flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/inwinstack/ip-assigner/pkg/operator"

	blended "github.com/inwinstack/blended/generated/clientset/versioned"

	"github.com/golang/glog"
	"github.com/inwinstack/ip-assigner/pkg/config"
	"github.com/inwinstack/ip-assigner/pkg/version"
	flag "github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	cfg        = &config.Config{}
	kubeconfig string
	ver        bool
)

func parserFlags() {
	flag.StringVarP(&kubeconfig, "kubeconfig", "", "", "Absolute path to the kubeconfig file.")
	flag.IntVarP(&cfg.Threads, "threads", "", 2, "Number of worker threads used by the controller.")
	flag.IntVarP(&cfg.SyncSec, "sync-seconds", "", 30, "Seconds for syncing and retrying objects.")
	flag.StringVarP(&cfg.PrivatePool, "private-pool", "", "default", "The default for the private pool.")
	flag.StringVarP(&cfg.PublicPool, "public-pool", "", "internet", "The default for the public pool.")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
}

func restConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	defer glog.Flush()
	parserFlags()

	if ver {
		fmt.Fprintf(os.Stdout, "%s\n", version.GetVersion())
		os.Exit(0)
	}

	k8scfg, err := restConfig(kubeconfig)
	if err != nil {
		glog.Fatalf("Failed to build kubeconfig: %s", err.Error())
	}

	k8sclient, err := kubernetes.NewForConfig(k8scfg)
	if err != nil {
		glog.Fatalf("Failed to build Kubernetes client: %s", err.Error())
	}

	blendedclient, err := blended.NewForConfig(k8scfg)
	if err != nil {
		glog.Fatalf("Failed to build Blended client: %s", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	op := operator.New(cfg, k8sclient, blendedclient)
	if err := op.Run(ctx); err != nil {
		glog.Fatalf("Error serving operator instance: %s.", err)
	}

	<-signalChan
	cancel()
	op.Stop()
	glog.Infof("Shutdown signal received, exiting...")
}
