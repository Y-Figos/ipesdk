package graph

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	fh "github.com/Y-Figos/ipesdk/internal/file_handler"
)
type RuntimeContext struct {
	Global map[string]any
}
type DAG struct {
	
	Nodes  map[string]*NodeModule
	Edges  map[string][]string
	Sorted [][]*NodeModule
}

func BuildGraph(manifest *fh.Manifest) *DAG {
	dag := DAG{Nodes: make(map[string]*NodeModule),
		Edges: make(map[string][]string)}
	root := filepath.Dir(manifest.ManifestPath)
	ctx := &RuntimeContext{
	Global: make(map[string]any),
	}
	for _, node := range manifest.NodeList {
		var output string
		if node.OutputArgs != nil{
			output = node.OutputArgs["export_as"].(string)
		}
		newModule := &NodeModule{
			ModuleName: node.Id,
			Adapter:    node.Adapter,
			InArgs:     node.InputArgs,
			OutArgs:     node.OutputArgs,
			ScriptPath: filepath.Join(root, "modules", node.Id, "script.lua"),
			DataOutput: output,
			ExportFlag: node.ExportFlag,
			Context: ctx,
		}
		dag.Nodes[node.Id] = newModule
	}

	for _, node := range manifest.NodeList {
		current := dag.Nodes[node.Id]
		if len(node.Depends) > 0 {
			for _, depId := range node.Depends {
				dep, ok := dag.Nodes[depId]
				if !ok {
					log.Printf("Warning: node %s depends on unknown node %s", node.Id, depId)
					continue
				}
				current.Depends = append(current.Depends, dep)

				dep.Children = append(dep.Children, current)
				dag.Edges[depId] = append(dag.Edges[depId], node.Id)
			}
		}
	}
	dag.Sorted = dag.exectutionLayers()
	return &dag
}

func (dag *DAG) validate() error {
	visitState := map[string]int{}

	var dfs func(node *NodeModule) error
	dfs = func(node *NodeModule) error {
		state := visitState[node.ModuleName]

		if state == 1 {
			return fmt.Errorf("cycle detected at module: %s", node.ModuleName)
		}
		if state == 2 {
			return nil
		}

		visitState[node.ModuleName] = 1

		for _, dep := range node.Depends {
			if err := dfs(dep); err != nil {
				return err
			}
		}

		visitState[node.ModuleName] = 2
		return nil
	}

	for _, node := range dag.Nodes {
		if visitState[node.ModuleName] == 0 {
			if err := dfs(node); err != nil {
				return err
			}
		}
	}

	return nil
}

func (dag *DAG) getInDegrees() map[string]int {
	inDegree := make(map[string]int)

	for id := range dag.Nodes {
		inDegree[id] = 0
	}

	for _, node := range dag.Nodes {
		for _, child := range node.Children {
			inDegree[child.ModuleName]++
		}
	}

	return inDegree
}

func (dag *DAG) exectutionLayers() [][]*NodeModule {
	inDegree := dag.getInDegrees()
	queue := []*NodeModule{}
	layers := [][]*NodeModule{}
	for id, degree := range inDegree {

		if degree == 0 {
			queue = append(queue, dag.Nodes[id])
		}
	}

	for len(queue) > 0 {
		currentLayer := queue
		queue = []*NodeModule{}

		for _, node := range currentLayer {
			for _, child := range node.Children {
				inDegree[child.ModuleName]--
				if inDegree[child.ModuleName] == 0 {
					queue = append(queue, child)
				}
			}
		}

		layers = append(layers, currentLayer)
	}
	dag.Sorted = layers
	return layers

}

func (dag *DAG) Run() error {

	if err := dag.validate(); err != nil {
		return fmt.Errorf("DAG not valid, cycle detected: %v", err)
	}

	for _, layer := range dag.Sorted {
		wg := sync.WaitGroup{}
		statusChan := make(chan ModuleStatus, len(layer))
		for _, node := range layer {
			log.Printf("Running Module: %v", node.ModuleName)
			wg.Add(1)
			go func(n *NodeModule) {
				defer wg.Done()
				status := n.Run()
				statusChan <- status
			}(node)
		}
		wg.Wait()
		close(statusChan)
		for status := range statusChan {
			if status == StatusFailed {
				return fmt.Errorf("stopping DAG execution due to module failure")
			}
		}
	}
	return nil
}
func (dag *DAG) RunModule(selectedNode *NodeModule) error {

	if err := dag.validate(); err != nil {
		return fmt.Errorf("DAG not valid, cycle detected: %v", err)
		
	}
	shouldStop := false
	for _, layer := range dag.Sorted {
		wg := sync.WaitGroup{}
		statusChan := make(chan ModuleStatus, len(layer))
		for _, node := range layer {
			log.Printf("Running Module: %v", node.ModuleName)
			wg.Add(1)
			n := node
			go func(n *NodeModule) {
				defer wg.Done()
				status := n.Run()
				statusChan <- status
			}(n)
			if node == selectedNode {
				shouldStop = true
				break
			}
		}
		wg.Wait()
		close(statusChan)
		for status := range statusChan {
			if status == StatusFailed {
				
				return fmt.Errorf("stopping DAG execution due to module failure")
			}
		}
		if shouldStop {
			break // break the outer loop
		}
	}
	return nil
}
