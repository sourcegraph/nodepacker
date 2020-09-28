package types

import (
	"fmt"
	"sort"

	"github.com/dustin/go-humanize"
)

type Resource struct {
	Name string

	// unit is MB
	Memory int64
	// unit
	CPU int64
	// unit is GB
	Storage int64
}

func (r Resource) String() string {
	mem := humanize.Ftoa(float64(r.Memory) / 1000.0)
	cpu := humanize.Ftoa(float64(r.CPU) / 1000.0)

	if r.Name != "" {
		return fmt.Sprintf("{%s, cpu: %s, mem: %s GB}", r.Name, cpu, mem)
	} else {
		return fmt.Sprintf("{cpu: %s, mem: %s GB}", cpu, mem)
	}
}

func AddResources(a, b Resource) Resource {
	return Resource{
		Memory:  a.Memory + b.Memory,
		CPU:     a.CPU + b.CPU,
		Storage: a.Storage + b.Storage,
	}
}

func SumResourceSlice(rs []Resource) Resource {
	var res Resource

	for _, r := range rs {
		res.Storage += r.Storage
		res.CPU += r.CPU
		res.Memory += r.Memory
	}
	return res
}

func SumResourceMap(ms map[string]Resource) Resource {
	var res Resource

	for _, r := range ms {
		res.Storage += r.Storage
		res.CPU += r.CPU
		res.Memory += r.Memory
	}
	return res
}

type ResourceFilter interface {
	Pass(r Resource) bool
}

type NoopResourceFilter struct{}

func (noopF *NoopResourceFilter) Pass(_ Resource) bool {
	return true
}

type BlockResourceFilter struct{}

func (blockF *BlockResourceFilter) Pass(_ Resource) bool {
	return false
}

// machines by zone and name
type Machines map[string]map[string]Resource

type ResourceComparator func(Resource, Resource) bool

type resourceSorter struct {
	vals  map[string]Resource
	order []string
	by    ResourceComparator
}

func (s *resourceSorter) Len() int {
	return len(s.order)
}

func (s *resourceSorter) Swap(i, j int) {
	s.order[i], s.order[j] = s.order[j], s.order[i]
}

func (s *resourceSorter) Less(i, j int) bool {
	return s.by(s.vals[s.order[i]], s.vals[s.order[j]])
}

func SortResources(resources map[string]Resource, by ResourceComparator) []string {
	keys := make([]string, 0, len(resources))
	for key := range resources {
		keys = append(keys, key)
	}
	rs := &resourceSorter{
		vals:  resources,
		by:    by,
		order: keys,
	}

	sort.Sort(rs)
	return rs.order
}

func HumanReadableMemCPU(r Resource) string {
	mem := float64(r.Memory) / 1000.0
	cpu := float64(r.CPU) / 1000.0

	return fmt.Sprintf("CPU %s, Mem %s", humanize.Ftoa(cpu), humanize.Ftoa(mem))
}
