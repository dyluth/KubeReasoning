package kubereasoning

import (
	"fmt"
	"strconv"

	"github.com/blues/jsonata-go"
	"github.com/dyluth/kube-reasoning/kubeloader"
)

type JobSet struct {
	data    *kubeJsonata
	podData *PodSet
}

// With returns a NEW JobSet with this additional filter
// filter needs to be a valid jsonata filter for items field, eg: "[metadata.name~>`/kube/i`]"
func (s *JobSet) With(filter string) *JobSet {
	ps := JobSet{
		data:    s.data.With(filter),
		podData: s.podData,
	}
	return &ps
}

func (ps *JobSet) WithNamespace(namespace string) *JobSet {
	return ps.With(fmt.Sprintf("metadata.namespace='%v'", namespace))
}

type JobResultType string

const (
	JobComplete JobResultType = "Complete"
	JobFailed   JobResultType = "Failed"
)

// Type is the result of the job, eg: Complete or Failed
func (ps *JobSet) WithType(t JobResultType) *JobSet {
	return ps.With(fmt.Sprintf("status.conditions[type='%v']", t))
}

func (ps *JobSet) Evaluate() ([]Job, error) {
	JobData, err := ps.data.Evaluate()
	if err != nil {
		return nil, err
	}
	Jobs := []Job{}
	for i := range JobData {
		if ps.podData == nil {
			fmt.Printf("erm..\n")
		}
		Jobs = append(Jobs, Job{data: JobData[i].Data, podData: &PodSet{data: &kubeJsonata{Data: ps.podData.data.Data}}})
	}
	return Jobs, nil
}

//jobs
type Job struct {
	data    interface{}
	podData *PodSet // so we can return the pods for this job
}

func (p *Job) Name() string {
	e1, _ := jsonata.Compile("metadata.name")
	res1, _ := e1.Eval(p.data)
	return fmt.Sprintf("%v", res1)
}

func (p *Job) NameSpace() string {
	e1, _ := jsonata.Compile("metadata.namespace")
	res1, _ := e1.Eval(p.data)
	return fmt.Sprintf("%v", res1)
}

func (p *Job) Counts() (passedCount, failedCount int, err error) {
	//e1, _ := jsonata.Compile("status.conditions[].(type & '-' & status)")
	// jsonata if statement - return the value of status.failed if it exists, else return `0`
	e1, _ := jsonata.Compile("status.failed ? status.failed : 0")
	res1, _ := e1.Eval(p.data)
	failedCount, err = strconv.Atoi(fmt.Sprintf("%v", res1))
	if err != nil {
		return
	}

	e2, _ := jsonata.Compile("status.succeeded ? status.succeeded : 0")
	res2, _ := e2.Eval(p.data)
	passedCount, err = strconv.Atoi(fmt.Sprintf("%v", res2))
	return
}

// return the pods for this jobset
func (p *Job) GetPods() ([]Pod, error) {
	pods := p.podData.With(fmt.Sprintf("metadata.labels.`job-name`='%v'", p.Name()))
	return pods.Evaluate()
}

func (p *Job) LastStatusChange() (condition Condition, err error) {
	return lastStatusChange(p.data)
}

// static load functions
func LoadJobSetFromFile(filename string) (*JobSet, error) {
	data, err := kubeloader.LoadFromFile(filename)
	ps := JobSet{
		data: &kubeJsonata{Data: data},
	}
	return &ps, err
}

// needs pods passed n so can return the pods for the appropriate job
func LoadJobsetFromKubectl(pods *PodSet) (*JobSet, error) {
	data, err := kubeloader.LoadFromKubectl("job")
	js := JobSet{
		data:    &kubeJsonata{Data: data},
		podData: pods,
	}
	return &js, err
}
