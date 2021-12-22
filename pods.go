package kubereasoning

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/blues/jsonata-go"
	"github.com/dyluth/kube-reasoning/kubeloader"
)

type PodSet struct {
	Data         interface{}
	QueryFilters []string // properties of `items` field, eg "[metadata.name~>`/kube/i`]"
}

// With returns a NEW PodSet with this additional filter
// filter needs to be a valid jsonata filter for items field, eg: "[metadata.name~>`/kube/i`]"
func (ps *PodSet) With(filter string) *PodSet {
	if !strings.HasPrefix(filter, "[") {
		filter = fmt.Sprintf("[%v]", filter)
	}

	ps2 := PodSet{
		Data:         ps.Data,
		QueryFilters: append(ps.QueryFilters, filter),
	}
	return &ps2
}

func (ps *PodSet) WithNamespace(namespace string) *PodSet {
	return ps.With(fmt.Sprintf("metadata.namespace='%v'", namespace))
}

// kind can be `Pod`, `DaemonSet`, `Job` etc
func (ps *PodSet) WithKind(kind string) *PodSet {
	return ps.With(fmt.Sprintf("metadata.ownerReferences.kind='%v'", kind))
}

//
//

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
	q := strings.Join(ps.QueryFilters, "")
	q = fmt.Sprintf("items%v", q)
	e, err := jsonata.Compile(q)
	if err != nil {
		return nil, err
	}
	res, err := e.Eval(ps.Data)
	if err != nil {
		return nil, err
	}

	// if res is a slice, foreach the slices and bang each one into a pod struct, then return that slice
	switch v := res.(type) {
	case []interface{}:
		v2 := res.([]interface{})
		pods := []Pod{}
		for i := range v2 {
			pods = append(pods, Pod{data: v2[i]})
		}
		return pods, nil
	default:
		return nil, fmt.Errorf("unknown type...: %v - %v", reflect.TypeOf(res), v)
	}
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
	e1, _ := jsonata.Compile("status.conditions[]")
	res1, _ := e1.Eval(p.data)

	conditions := []Condition{}
	res1Marshalled, err := json.Marshal(res1)
	if err != nil {
		return condition, err
	}

	err = json.Unmarshal(res1Marshalled, &conditions)
	if err != nil {
		return condition, err
	}
	//[map[lastProbeTime:<nil> lastTransitionTime:2021-12-17T09:00:01Z reason:PodCompleted status:True type:Initialized] map[lastProbeTime:<nil> lastTransitionTime:2021-12-17T09:00:16Z reason:PodCompleted status:False type:Ready] map[lastProbeTime:<nil> lastTransitionTime:2021-12-17T09:00:16Z reason:PodCompleted status:False type:ContainersReady] map[lastProbeTime:<nil> lastTransitionTime:2021-12-17T09:00:01Z status:True type:PodScheduled]]
	if len(conditions) > 0 {
		newest := conditions[0]
		for i := range conditions {
			if conditions[i].LastTransitionTime.After(newest.LastTransitionTime) {
				newest = conditions[i]
			}
		}
		return newest, nil

	}
	return condition, fmt.Errorf("no conditions")
}

type Condition struct {
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Message            string    `json:"message"`
	Reason             string    `json:"reason"`
	Status             string    `json:"status"`
	Type               string    `json:"type"`
}

func LoadPodSetFromFile(filename string) (*PodSet, error) {
	data, err := kubeloader.LoadFromFile(filename)
	ps := PodSet{
		Data:         data,
		QueryFilters: []string{},
	}
	return &ps, err
}

func LoadPodsetFromKubectl() (*PodSet, error) {
	data, err := kubeloader.LoadFromKubectl("pod")
	ps := PodSet{
		Data:         data,
		QueryFilters: []string{},
	}
	return &ps, err
}
