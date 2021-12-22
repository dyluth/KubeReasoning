package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	kubereasoning "github.com/dyluth/kube-reasoning"
	"github.com/dyluth/kube-reasoning/kubeloader"
)

func main() {
	var clusterFilter string
	flag.StringVar(&clusterFilter, "filter", "", "only act on clusters that match to this string")
	flag.Parse()

	// load from kubectl command line and store in a file cache
	kubeloader.LoaderCache = &kubeloader.SimpleFileCache{}

	clusterNames, err := kubeloader.GetClusters(clusterFilter)
	fmt.Printf("===Acting on the following clusters: \n  %v\n", strings.Join(clusterNames, "\n  "))
	if err != nil {
		panic(err)
	}
	for i := range clusterNames {
		reason(clusterNames[i])
	}
}

func reason(clusterName string) {
	err := kubeloader.SetContext(clusterName)
	if err != nil {
		fmt.Printf("error switching context: %v\n", err)
		return
	}

	nodes, err := kubereasoning.LoadNodesetFromKubectl()
	if err != nil {
		panic(err)
	}

	podSet, err := kubereasoning.LoadPodsetFromKubectl()
	if err != nil {
		panic(err)
	}
	allNodes, _ := nodes.Evaluate()
	fmt.Printf("=====================\n %v (%v nodes) \n=====================\n", clusterName, len(allNodes))

	podSummary(podSet, "Pod")
	podSummary(podSet, "Job")
	podSummary(podSet, "DaemonSet")
}

func podSummary(podSet *kubereasoning.PodSet, kind string) {
	all := podSet.WithKind(kind)
	unhealthy := all.WithIsHealthy(false)
	pods, _ := all.Evaluate()                //.Evaluate()
	unhealthyPods, _ := unhealthy.Evaluate() //.Evaluate()
	fmt.Printf("= %v %v (%v unhealthy)\n", kind, len(pods), len(unhealthyPods))
	for i := range unhealthyPods {
		status, err := pods[i].LastStatusChange()
		if err != nil {
			fmt.Printf("%v %v [no conditions]\n", pods[i].NameSpace(), pods[i].Name())
		} else {
			since := time.Since(status.LastTransitionTime)
			fmt.Printf("%v %v [%v:%v for %d hours]\n", pods[i].NameSpace(), pods[i].Name(), status.Type, status.Status, since.Round(time.Hour)/time.Hour)
		}
	}
}
