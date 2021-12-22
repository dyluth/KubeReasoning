package main

import (
	"fmt"

	kubereasoning "github.com/dyluth/kube-reasoning"
	"github.com/dyluth/kube-reasoning/kubeloader"
	//"github.com/pkg/profile" // for memory profiling
)

func main() {
	// defer profile.Start(profile.MemProfile).Stop()

	// load from a specific file
	//podSet, err := kubereasoning.LoadPodSetFromFile("cmd/example/getpods-Aexample.json")

	// load from kubectl command line and store in a file cache
	kubeloader.LoaderCache = &kubeloader.SimpleFileCache{}

	nodeSet, err := kubereasoning.LoadNodesetFromKubectl()
	if err != nil {
		panic(err)
	}
	nodes, err := nodeSet.Evaluate()
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %v nodes\n", len(nodes))
	for i := range nodes {
		status, err := nodes[i].LastStatusChange()
		if err != nil {
			fmt.Printf("  %v - [no conditions]\n", nodes[i].Name())

		} else {
			fmt.Printf("  %v - [%v:%v for %d hours]\n", nodes[i].Name(), status.Type, status.Status, status.HoursSince())
		}
	}

	podSet, err := kubereasoning.LoadPodsetFromKubectl()
	if err != nil {
		panic(err)
	}
	podSummary(podSet, "Pod")
	podSummary(podSet, "Job")
	podSummary(podSet, "DaemonSet")
}

func podSummary(podSet *kubereasoning.PodSet, kind string) {
	kindSet := podSet.WithKind(kind)
	unhealthy := kindSet.WithIsHealthy(false)
	pods, _ := kindSet.Evaluate()
	unhealthyPods, _ := unhealthy.Evaluate()
	//pods, err := podSet.With("metadata.name~>/kube/i").WithNamespace("kube-system").WithKind("DaemonSet").Evaluate()
	fmt.Printf("============== %v %v (%v unhealthy) ==============\n", kind, len(pods), len(unhealthyPods))
	for i := range pods {
		status, err := pods[i].LastStatusChange()
		if err != nil {
			fmt.Printf("%v %v [no conditions]\n", pods[i].NameSpace(), pods[i].Name())
		} else {
			fmt.Printf("%v %v [%v:%v for %d hours]\n", pods[i].NameSpace(), pods[i].Name(), status.Type, status.Status, status.HoursSince())
		}
	}
}
