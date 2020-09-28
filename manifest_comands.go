package nodepacker

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"nodepacker/types"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
)

func makeAbs(paths []string) ([]string, error) {
	var pas []string

	for _, path := range paths {
		pa, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		pas = append(pas, pa)
	}
	return pas, nil
}

func parseResourceString(s string, isCPUValue bool) int64 {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return 0
	}
	if isCPUValue {
		return q.MilliValue()
	} else {
		return q.Value() / 1000000
	}
}

func parseResource(name string, from map[string]interface{}) int64 {
	if text, ok := from[name].(string); ok {
		return parseResourceString(text, name == "cpu")
	}
	return 0
}

func resourceFrom(trec map[string]interface{}) types.Resource {
	return types.Resource{
		Memory:  parseResource("memory", trec),
		CPU:     parseResource("cpu", trec),
		Storage: parseResource("storage", trec),
	}
}

func extractTotalResource(rec map[string]interface{}) types.Resource {
	var r types.Resource

	for k, v := range rec {
		vmap, ok := v.(map[string]interface{})
		if !ok {
			vlist, ok := v.([]interface{})
			if !ok {
				continue
			}
			for _, vmember := range vlist {
				if vmap, ok := vmember.(map[string]interface{}); ok {
					r = types.AddResources(r, extractTotalResource(vmap))
				}
			}
		} else {
			if k == "resources" {
				if requests, ok := vmap["requests"].(map[string]interface{}); ok {
					r = types.AddResources(r, resourceFrom(requests))
				}
			} else {
				r = types.AddResources(r, extractTotalResource(vmap))
			}
		}
	}
	return r
}

func extractName(res map[string]interface{}) (string, error) {
	meta, ok := res["metadata"].(map[string]interface{})
	if !ok {
		return "", errors.New("no metadata")
	}

	name, ok := meta["name"].(string)
	if !ok {
		return "", errors.New("no name")
	}
	return name, nil
}

func extractNumReplicas(res map[string]interface{}) (int, error) {
	spec, ok := res["spec"].(map[string]interface{})
	if !ok {
		return 0, errors.New("no spec")
	}

	nr, ok := spec["replicas"].(int)
	if !ok {
		return 0, errors.New("no replicas")
	}
	return nr, nil
}

func loadManifest(path string) ([]types.Resource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	decoder := yaml.NewDecoder(br)

	var res map[string]interface{}
	err = decoder.Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode yaml file: %s: %v", path, err)
	}

	kind, ok := res["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("resource %s is missing a kind field", path)
	}

	switch kind {
	case "Deployment":
		fallthrough
	case "StatefulSet":
		r := extractTotalResource(res)
		name, err := extractName(res)
		if err != nil {
			return nil, err
		}
		r.Name = name
		numReplicas, err := extractNumReplicas(res)
		if err != nil {
			return nil, err
		}
		var rs []types.Resource

		if numReplicas == 1 {
			rs = append(rs, r)
		} else {
			for i := 0; i < numReplicas; i++ {
				rs = append(rs, types.Resource{
					Memory:  r.Memory,
					CPU:     r.CPU,
					Storage: r.Storage,
					Name:    fmt.Sprintf("%s-%d", r.Name, i),
				})
			}
		}
		return rs, nil
	default:
		return nil, nil
	}
}

func loadManifests(inputs []string) (map[string]types.Resource, error) {
	pas, err := makeAbs(inputs)
	if err != nil {
		return nil, err
	}

	podsByName := make(map[string]types.Resource)
	for _, input := range pas {
		err = filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
				pods, err := loadManifest(path)
				if err != nil {
					return err
				}

				for _, pod := range pods {
					podsByName[pod.Name] = pod
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return podsByName, nil
}

func readManifestsCommand(cctx *CommandContext, args []string) {
	mfs, err := loadManifests(args)
	if err != nil {
		fmt.Printf("failed to load manifests from %v: %v", args, err)
	}

	cctx.pods = mfs
	fmt.Println("got the manifests")
}
