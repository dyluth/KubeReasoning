package main

import (
	"fmt"

	kubereasoning "github.com/dyluth/kube-reasoning"
	//"github.com/pkg/profile"
)

func main() {
	// defer profile.Start(profile.MemProfile).Stop()
	podSet, err := kubereasoning.LoadPodSetFromFile("cmd/example/getpods-Aexample.json")
	if err != nil {
		panic(err)
	}
	pods, err := podSet.With(
		"metadata.name~>/kube/i").With(
		"metadata.namespace='kube-system'").With(
		"metadata.ownerReferences.kind='DaemonSet'").Evaluate()

	for i := range pods {
		fmt.Println(pods[i].Name())
	}
	fmt.Printf("ERROR: %v\n", err)
}
