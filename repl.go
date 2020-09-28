package nodepacker

import (
	"fmt"
	"sort"
	"strings"

	"nodepacker/types"
	"github.com/c-bata/go-prompt"
)

type CommandContext struct {
	pods     map[string]types.Resource
	nodes    map[string]types.Resource
	machines types.Machines
	zone     string
}

type CommandFn func(*CommandContext, []string)

type ArgsCompleter func(string, *CommandContext) []prompt.Suggest

type CommandReplHandler struct {
	buf           string
	commands      map[string]CommandFn
	cctx          *CommandContext
	cc            *commandCompleter
	argsCompleter map[string][]ArgsCompleter
}

type handlerBuilder struct {
	commands                 map[string]CommandFn
	allCommandsSuggestion    []prompt.Suggest
	commandSuggestionsByName map[string]prompt.Suggest
	sortedCommandNames       []string
	argsCompleter            map[string][]ArgsCompleter
}

func newBuilder() *handlerBuilder {
	commands := make(map[string]CommandFn)
	commandSuggestionsByName := make(map[string]prompt.Suggest)
	argsCompleter := make(map[string][]ArgsCompleter)
	return &handlerBuilder{
		commands:                 commands,
		commandSuggestionsByName: commandSuggestionsByName,
		argsCompleter:            argsCompleter,
	}
}

func (hb *handlerBuilder) add(fn CommandFn, name, description string, acs ...ArgsCompleter) *handlerBuilder {
	hb.commands[name] = fn
	hb.commandSuggestionsByName[name] = prompt.Suggest{
		Text:        name,
		Description: description,
	}
	hb.sortedCommandNames = append(hb.sortedCommandNames, name)
	hb.allCommandsSuggestion = append(hb.allCommandsSuggestion, prompt.Suggest{
		Text:        name,
		Description: description,
	})
	if acs != nil {
		hb.argsCompleter[name] = acs
	}
	return hb
}

func (hb *handlerBuilder) build() *CommandReplHandler {
	sort.Strings(hb.sortedCommandNames)

	machines, err := readMachines()
	if err != nil {
		fmt.Println("couldn't read machines from ~/.nodepacker/machines. please execute command machines_fetch")
	}

	return &CommandReplHandler{
		commands: hb.commands,
		cc: &commandCompleter{
			allCommandsSuggestion:    hb.allCommandsSuggestion,
			sortedCommandNames:       hb.sortedCommandNames,
			commandSuggestionsByName: hb.commandSuggestionsByName,
		},
		cctx:          &CommandContext{zone: "us-central1-a", machines: machines},
		argsCompleter: hb.argsCompleter,
	}
}

func NewCommandReplHandler() *CommandReplHandler {
	hb := newBuilder()

	hb.add(helpCommand, "help", "help [command]", nil)
	hb.add(exitCommand, "exit", "exit nodepacker", nil)

	hb.add(fetchMachinesCommand, "machines_fetch", "fetch available machines from GCP", nil)
	hb.add(getSetZoneCommand, "machines_zone", "get or set current zone", zoneComplete)
	hb.add(showMachinesCommand, "machines_show", "show machines available in current zone", nil)

	hb.add(addNodesCommand, "nodes_add", "add nodes to cluster", machineComplete)
	hb.add(packCommand, "nodes_pack", "pack nodes", nil)

	hb.add(readManifestsCommand, "manifests_read", "read manifests", pathComplete)

	hb.add(showPodsCommand, "pods_show", "show pods", nil)

	return hb.build()
}

func (crh *CommandReplHandler) ExecuteStatement(statement string) {
	parts := strings.Fields(statement)

	n := len(parts)
	if n == 0 {
		return
	}

	if parts[n-1] == ";" {
		parts = parts[0 : n-1]
	} else {
		parts[n-1] = strings.TrimSuffix(parts[n-1], ";")
	}

	command := parts[0]
	args := parts[1:]

	fn, ok := crh.commands[command]
	if !ok {
		fmt.Println("unknown command " + command)
		return
	}

	fn(crh.cctx, args)
}
