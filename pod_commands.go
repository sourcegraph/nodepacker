package nodepacker

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"nodepacker/filter"
	"nodepacker/types"
	"github.com/dustin/go-humanize"
)

func showPodsCommand(cctx *CommandContext, args []string) {
	pods := cctx.pods

	mf, err := filter.Create(args)
	if err != nil {
		fmt.Println("failed to build filter", err)
		mf = &types.NoopResourceFilter{}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)

	podKeys := make([]string, 0, len(pods))
	for k := range pods {
		podKeys = append(podKeys, k)
	}

	sort.Strings(podKeys)

	for _, k := range podKeys {
		v := pods[k]
		if mf.Pass(v) {
			mem := humanize.Ftoa(float64(v.Memory) / 1000.0)
			cpu := humanize.Ftoa(float64(v.CPU) / 1000.0)

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s GB\t\n", k, cpu, mem)
		}
	}
	_ = w.Flush()

	totalRes := types.SumResourceMap(pods)
	mem := humanize.Ftoa(float64(totalRes.Memory) / 1000.0)
	cpu := humanize.Ftoa(float64(totalRes.CPU) / 1000.0)

	fmt.Printf("\ntotal CPU: %s, total mem: %s\n", cpu, mem)
}
