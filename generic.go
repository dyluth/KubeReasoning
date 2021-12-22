package kubereasoning

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/blues/jsonata-go"
)

// kubeJsonata is the base structure for handling kubernetes json output with jsonata
// this can then be wrappered by convenience functions
type kubeJsonata struct {
	Data         interface{}
	QueryFilters []string // properties of `items` field, eg "[metadata.name~>`/kube/i`]"
}

func (ps *kubeJsonata) With(filter string) *kubeJsonata {
	if !strings.HasPrefix(filter, "[") {
		filter = fmt.Sprintf("[%v]", filter)
	}

	ps2 := kubeJsonata{
		Data:         ps.Data,
		QueryFilters: append(ps.QueryFilters, filter),
	}
	return &ps2
}

// actually runs the query and breaks the data down into a slice of interfaces that match
func (ps *kubeJsonata) Evaluate() ([]kubeJsonata, error) {
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
		generic := []kubeJsonata{}
		for i := range v2 {
			generic = append(generic, kubeJsonata{Data: v2[i]})
		}
		return generic, nil
	default:
		return nil, fmt.Errorf("unknown type...: %v - %v", reflect.TypeOf(res), v)
	}
}

// Status and conditions

func lastStatusChange(data interface{}) (Condition, error) {
	conditions, err := getConditions(data)
	if err != nil {
		return Condition{}, err
	}
	if len(conditions) > 0 {
		newest := conditions[0]
		for i := range conditions {
			if conditions[i].LastTransitionTime.After(newest.LastTransitionTime) {
				newest = conditions[i]
			}
		}
		return newest, nil

	}
	return Condition{}, fmt.Errorf("no conditions")
}

func getConditions(data interface{}) ([]Condition, error) {
	e1, _ := jsonata.Compile("status.conditions[]")
	res1, _ := e1.Eval(data)

	conditions := []Condition{}
	res1Marshalled, err := json.Marshal(res1)
	if err != nil {
		return conditions, err
	}

	err = json.Unmarshal(res1Marshalled, &conditions)
	return conditions, err
}

type Condition struct {
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Message            string    `json:"message"`
	Reason             string    `json:"reason"`
	Status             string    `json:"status"`
	Type               string    `json:"type"`
}

func (c *Condition) HoursSince() int {
	since := time.Since(c.LastTransitionTime)
	return int(since.Round(time.Hour) / time.Hour)
}
