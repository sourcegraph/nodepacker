package nodepacker

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"nodepacker/filter"
	"nodepacker/types"
	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
)

// Returns a map with keys zones and values a map with keys machine name and value Resource
func availableMachines() (types.Machines, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gcloud", "compute", "machine-types", "list")
	cmd.Stderr = os.Stderr
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	zones := make(types.Machines)

	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if len(parts) < 4 {
			continue
		}
		if parts[0] == "NAME" {
			continue
		}

		machines := zones[parts[1]]
		if machines == nil {
			machines = make(map[string]types.Resource)
			zones[parts[1]] = machines
		}
		machines[parts[0]] = types.Resource{
			Name:   parts[0],
			Memory: parseMachineMem(parts[3]),
			CPU:    parseMachineCPU(parts[2]),
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return zones, nil
}

func parseMachineMem(s string) int64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	v = v * 1000
	return int64(v)
}

func parseMachineCPU(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	v = v * 1000
	return v
}

func fetchMachinesCommand(cctx *CommandContext, args []string) {
	machines, err := availableMachines()
	if err != nil {
		fmt.Println("error getting available machines: ", err)
	}

	cctx.machines = machines
	fmt.Println("got the machines")
	err = saveMachines(machines)
	if err != nil {
		fmt.Println("failed to save machines in ~/.nodepacker/machines:", err)
	}
}

func showMachinesCommand(cctx *CommandContext, args []string) {
	ms := cctx.machines[cctx.zone]

	cpuSorter := func(a types.Resource, b types.Resource) bool {
		if a.CPU < b.CPU {
			return true
		}
		if a.CPU == b.CPU && a.Memory < b.Memory {
			return true
		}
		return false
	}

	memSorter := func(a types.Resource, b types.Resource) bool {
		if a.Memory < b.Memory {
			return true
		}
		if a.Memory == b.Memory && a.CPU < b.CPU {
			return true
		}
		return false
	}

	fs := flag.NewFlagSet("showMachinesCommand", flag.ContinueOnError)
	sortOrder := fs.String("sort", "cpu", "sort order")

	err := fs.Parse(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	sorter := cpuSorter
	if *sortOrder == "mem" {
		sorter = memSorter
	}

	sortedMs := types.SortResources(ms, sorter)

	mf, err := filter.Create(fs.Args())
	if err != nil {
		fmt.Println("failed to build filter", err)
		mf = &types.NoopResourceFilter{}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)

	for _, k := range sortedMs {
		v := ms[k]

		if mf.Pass(v) {
			mem := humanize.Ftoa(float64(v.Memory) / 1000.0)
			cpu := humanize.Ftoa(float64(v.CPU) / 1000.0)

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s GB\t\n", k, cpu, mem)
		}
	}
	_ = w.Flush()
}

func saveMachines(machines types.Machines) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(usr.HomeDir, ".nodepacker"), 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(usr.HomeDir, ".nodepacker", "machines.yaml"))
	if err != nil {
		return err
	}
	defer f.Close()

	bf := bufio.NewWriter(f)
	defer bf.Flush()

	e := yaml.NewEncoder(bf)
	return e.Encode(machines)
}

func readMachines() (types.Machines, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filepath.Join(usr.HomeDir, ".nodepacker", "machines.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)

	var machines types.Machines

	e := yaml.NewDecoder(br)
	err = e.Decode(&machines)
	if err != nil {
		return nil, err
	}
	return machines, nil
}

func getSetZoneCommand(cctx *CommandContext, args []string) {
	if len(args) == 0 {
		fmt.Println(cctx.zone)
		return
	}
	if len(args) == 1 {
		_, ok := cctx.machines[args[0]]
		if !ok {
			fmt.Println("unknown zone")
			return
		}
		cctx.zone = args[0]
		fmt.Println("set current zone to ", args[0])
		return
	}

	fmt.Println("expected no args to get current zone or one argument to set current zone")
}
