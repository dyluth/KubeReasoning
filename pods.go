package kubereasoning

import (
	"fmt"

	"github.com/blues/jsonata-go"
	"github.com/dyluth/kube-reasoning/kubeloader"
)

type PodSet struct {
	data *kubeJsonata
}

// With returns a NEW PodSet with this additional filter
// filter needs to be a valid jsonata filter for items field, eg: "[metadata.name~>`/kube/i`]"
func (s *PodSet) With(filter string) *PodSet {
	ps := PodSet{
		data: s.data.With(filter),
	}
	return &ps
}

func (ps *PodSet) WithNamespace(namespace string) *PodSet {
	return ps.With(fmt.Sprintf("metadata.namespace='%v'", namespace))
}

// kind can be `Pod`, `DaemonSet`, `Job` etc
func (ps *PodSet) WithKind(kind string) *PodSet {
	return ps.With(fmt.Sprintf("metadata.ownerReferences.kind='%v'", kind))
}

func (ps *PodSet) WithIsHealthy(healthy bool) *PodSet {
	// status conditions [] status <True | False>
	//"status": "False",
	if !healthy {
		// return any status which is false
		return ps.With("status.conditions[status='False']")
	}
	// return any pod with no status conditions statuses of `False``
	return ps.With("$count(status.conditions[status='False'])=0")
}

func (ps *PodSet) Evaluate() ([]Pod, error) {
	podData, err := ps.data.Evaluate()
	if err != nil {
		return nil, err
	}
	pods := []Pod{}
	for i := range podData {
		pods = append(pods, Pod{data: podData[i].Data})
	}
	return pods, nil
}

type Pod struct {
	data interface{}
}

func (p *Pod) Name() string {
	e1, _ := jsonata.Compile("metadata.name")
	res1, _ := e1.Eval(p.data)
	return fmt.Sprintf("%v", res1)
}
func (p *Pod) NameSpace() string {
	e1, _ := jsonata.Compile("metadata.namespace")
	res1, _ := e1.Eval(p.data)
	return fmt.Sprintf("%v", res1)
}

func (p *Pod) Statuses() string {
	e1, _ := jsonata.Compile("status.conditions[].(type & '-' & status)")
	res1, _ := e1.Eval(p.data)
	return fmt.Sprintf("%v", res1)
}

func (p *Pod) LastStatusChange() (condition Condition, err error) {
	return lastStatusChange(p.data)
}

func LoadPodSetFromFile(filename string) (*PodSet, error) {
	data, err := kubeloader.LoadFromFile(filename)
	ps := PodSet{
		data: &kubeJsonata{Data: data},
	}
	return &ps, err
}

func LoadPodsetFromKubectl() (*PodSet, error) {
	data, err := kubeloader.LoadFromKubectl("pod")
	ps := PodSet{
		data: &kubeJsonata{Data: data},
	}
	return &ps, err
}
