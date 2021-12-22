package main

import (
	"fmt"
	"time"

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
	podSet, err := kubereasoning.LoadPodsetFromKubectl()
	if err != nil {
		panic(err)
	}
	podSummary(podSet, "Pod")
	podSummary(podSet, "Job")
	podSummary(podSet, "DaemonSet")
}

func podSummary(podSet *kubereasoning.PodSet, kind string) {
	pods, _ := podSet.WithKind(kind).Evaluate() //WithIsHealthy(false).Evaluate()
	//pods, err := podSet.With("metadata.name~>/kube/i").WithNamespace("kube-system").WithKind("DaemonSet").Evaluate()
	fmt.Printf("============== %v %v ==============\n", kind, len(pods))
	for i := range pods {
		status, err := pods[i].LastStatusChange()
		if err != nil {
			fmt.Printf("%v %v [no conditions]\n", pods[i].NameSpace(), pods[i].Name())
		} else {
			since := time.Since(status.LastTransitionTime)
			fmt.Printf("%v %v [%v:%v for %d hours]\n", pods[i].NameSpace(), pods[i].Name(), status.Type, status.Status, since.Round(time.Hour)/time.Hour)
		}
	}
}
