package kubereasoning

import (
	"fmt"

	"github.com/blues/jsonata-go"
	"github.com/dyluth/kube-reasoning/kubeloader"
)

type NodeSet struct {
	data *kubeJsonata
}

func (s *NodeSet) With(filter string) *NodeSet {
	ns := NodeSet{
		data: s.data.With(filter),
	}
	return &ns
}

func (s *NodeSet) Evaluate() ([]Node, error) {
	nodeData, err := s.data.Evaluate()
	if err != nil {
		return nil, err
	}
	nodes := []Node{}
	for i := range nodeData {
		nodes = append(nodes, Node{data: nodeData[i].Data})
	}
	return nodes, nil
}

type Node struct {
	data interface{}
}

func (n *Node) Name() string {
	e1, _ := jsonata.Compile("metadata.name")
	res1, _ := e1.Eval(n.data)
	return fmt.Sprintf("%v", res1)
}

func (n *Node) Statuses() string {
	e1, _ := jsonata.Compile("status.conditions[].(type & '-' & status)")
	res1, _ := e1.Eval(n.data)
	return fmt.Sprintf("%v", res1)
}

func (n *Node) LastStatusChange() (condition Condition, err error) {
	return lastStatusChange(n.data)
}

func LoadNodeSetFromFile(filename string) (*NodeSet, error) {
	data, err := kubeloader.LoadFromFile(filename)
	ns := NodeSet{
		data: &kubeJsonata{
			Data:         data,
			QueryFilters: []string{},
		},
	}
	return &ns, err
}

func LoadNodesetFromKubectl() (*NodeSet, error) {
	data, err := kubeloader.LoadFromKubectl("node")
	ns := NodeSet{
		data: &kubeJsonata{
			Data:         data,
			QueryFilters: []string{},
		},
	}
	return &ns, err
}
