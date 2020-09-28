package nodepacker

import (
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"nodepacker/types"
	"github.com/c-bata/go-prompt"
)

type commandCompleter struct {
	allCommandsSuggestion    []prompt.Suggest
	commandSuggestionsByName map[string]prompt.Suggest
	sortedCommandNames       []string
}

func (cc *commandCompleter) complete(cmdPrefix string) []prompt.Suggest {
	if cmdPrefix == "" {
		return cc.allCommandsSuggestion
	}

	idx := sort.SearchStrings(cc.sortedCommandNames, cmdPrefix)

	var ps []prompt.Suggest

	for i := idx; i < len(cc.sortedCommandNames); i++ {
		if strings.HasPrefix(cc.sortedCommandNames[i], cmdPrefix) {
			ps = append(ps, cc.commandSuggestionsByName[cc.sortedCommandNames[i]])
		} else {
			break
		}
	}

	return ps
}

func (crh *CommandReplHandler) Completer(d prompt.Document) []prompt.Suggest {
	prefix := d.TextBeforeCursor()
	endsWithSpace := strings.HasSuffix(prefix, " ")
	prefix = strings.TrimSpace(prefix)

	if len(prefix) == 0 {
		return crh.cc.complete(prefix)
	}

	parts := strings.Fields(prefix)
	if len(parts) == 1 && !endsWithSpace {
		return crh.cc.complete(prefix)
	}

	acs := crh.argsCompleter[parts[0]]
	if acs == nil {
		return nil
	}
	n := len(parts)
	partial := parts[n-1]
	pos := n - 1
	if endsWithSpace {
		pos += 1
		partial = ""
	}
	if len(acs) <= pos {
		return nil
	}
	return acs[pos](partial, crh.cctx)
}

func pathComplete(prefix string, cctx *CommandContext) []prompt.Suggest {
	dir := filepath.Dir(prefix)
	base := filepath.Base(prefix)

	if base == "." || base == ".." || strings.HasSuffix(prefix, "/") {
		base = ""
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}

	files, err := ioutil.ReadDir(absDir)
	if err != nil {
		return nil
	}

	var res []prompt.Suggest

	for _, file := range files {
		if strings.HasPrefix(file.Name(), base) {
			res = append(res, prompt.Suggest{Text: filepath.Join(dir, file.Name()),
				Description: filepath.Join(absDir, file.Name())})
		}
	}
	return res
}

func machineComplete(prefix string, cctx *CommandContext) []prompt.Suggest {
	var res []prompt.Suggest

	choices := cctx.machines[cctx.zone]
	for k, v := range choices {
		if strings.HasPrefix(k, prefix) {
			res = append(res, prompt.Suggest{Text: k, Description: types.HumanReadableMemCPU(v)})
		}
	}
	return res
}

func zoneComplete(prefix string, cctx *CommandContext) []prompt.Suggest {
	var res []prompt.Suggest

	for k := range cctx.machines {
		if strings.HasPrefix(k, prefix) {
			res = append(res, prompt.Suggest{Text: k})
		}
	}
	return res
}
