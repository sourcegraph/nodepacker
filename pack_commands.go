package nodepacker

import (
	"fmt"
	"math"
	"strings"

	"nodepacker/types"
)

func relativeCost(a, b int64) float64 {
	rc := float64(a - b) / float64(a + b)
	if rc < 0 {
		return 1.0
	}
	return rc
}

// - find a machine type that can accomodate 2 indexed-search pods
// - multiple by number of indexed-search pods to get cluster node pool
// - place an indexed-search pod in each node
// - use simple binpacking to place the rest and see how much is leftover per node
// - simple binpacking:
//        - sort by largest to smallest CPU (ties largest to smallest mem)
//        - place them in this order in nodes with most space left
//        - add additional nodes for left-overs if needed
func packCommand(cctx *CommandContext, args []string) {
	ms := cctx.machines[cctx.zone]

	// - find a machine type that can accomodate 2 indexed-search pods
	//   we do a linear search for a machine type that minimizes the cost function 'relativeCost'
	idxSearchR := cctx.pods["indexed-search-0"]
	mem := idxSearchR.Memory * 2
	cpu := idxSearchR.CPU * 2

	bestMachineType := ""
	minCost := 1.0

	for mName, mResource := range ms {
		cost := math.Max(relativeCost(mResource.Memory, mem), relativeCost(mResource.CPU, cpu))
		if cost < minCost {
			minCost = cost
			bestMachineType = mName
		}
	}

	// - multiple by number of indexed-search pods to get cluster node pool
	indexedSearchReplicas := 1
	_, ok := cctx.pods[fmt.Sprintf("indexed-search-%d", indexedSearchReplicas)]
	for ok {
		indexedSearchReplicas++
		_, ok = cctx.pods[fmt.Sprintf("indexed-search-%d", indexedSearchReplicas)]
	}
	fmt.Printf("replica count for indexed search is %d\n", indexedSearchReplicas)

	// - place an indexed-search pod in each node
	numNodes := indexedSearchReplicas
	nodeAssign := make(map[string][]string)
	freeSpace := make(map[string]types.Resource)
	for i := 0; i < numNodes; i++ {
		name := fmt.Sprintf("node-%d", i)
		freeSpace[name] = types.Resource{
			Name: name,
			Memory: ms[bestMachineType].Memory - idxSearchR.Memory,
			CPU: ms[bestMachineType].CPU - idxSearchR.CPU,
		}
		nodeAssign[name] = append(nodeAssign[name], fmt.Sprintf("indexed-search-%d", i))
	}

	descendingSorter := func(a types.Resource, b types.Resource) bool {
		if a.CPU > b.CPU {
			return true
		}
		if a.CPU == b.CPU && a.Memory > b.Memory {
			return true
		}
		return false
	}

	// - use simple binpacking to place the rest and see how much is leftover per node
	todo := make(map[string]types.Resource)
	for k, v := range cctx.pods {
		if !strings.HasPrefix(k, "indexed-search") {
			todo[k] = v
		}
	}
	sortedTodo := types.SortResources(todo, descendingSorter)

	numNotAssigned := len(sortedTodo)
	for numNotAssigned > 0 {
		from := 0
		for from < len(sortedTodo) {
			i := from
			for i < len(sortedTodo) && sortedTodo[i] == "" {
				i++
			}
			if i == len(sortedTodo) {
				break
			}
			pod := todo[sortedTodo[i]]

			mostFreeNodeName := ""
			mostFreeMem := int64(0)
			mostFreeCPU := int64(0)

			for nodeName, node := range freeSpace {
				if node.CPU > mostFreeCPU {
					mostFreeNodeName = nodeName
					mostFreeCPU = node.CPU
					mostFreeMem = node.Memory
				} else if node.CPU == mostFreeCPU && node.Memory > mostFreeMem {
					mostFreeNodeName = nodeName
					mostFreeCPU = node.CPU
					mostFreeMem = node.Memory
				}
			}
			if mostFreeNodeName == "" {
				break
			}
			nodeName, node := mostFreeNodeName, freeSpace[mostFreeNodeName]
			if node.Memory-pod.Memory >= 0 && node.CPU-pod.CPU >= 0 {
				freeSpace[nodeName] = types.Resource{
					Memory: node.Memory - pod.Memory,
					CPU:    node.CPU - pod.CPU,
				}
				nodeAssign[nodeName] = append(nodeAssign[nodeName], pod.Name)
				numNotAssigned--
				sortedTodo[i] = ""
			}
			from++
		}

		if numNotAssigned > 0 {
			nodeName := fmt.Sprintf("node-%d", numNodes)
			numNodes++
			freeSpace[nodeName] = types.Resource{
				Name:   nodeName,
				Memory: ms[bestMachineType].Memory,
				CPU:    ms[bestMachineType].CPU,
			}
		}
	}

	fmt.Printf("cluster with %d nodes of machine type %s\n", numNodes, ms[bestMachineType].String())
	fmt.Println("Pod assignment as follows:")
	for i := 0; i < numNodes; i++ {
		nodeName := fmt.Sprintf("node-%d", i)
		fmt.Printf("%s: [%s], free %s\n", nodeName,
			strings.Join(nodeAssign[nodeName], ", "), freeSpace[nodeName].String())
	}
}
