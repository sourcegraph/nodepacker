package nodepacker

import (
	"fmt"
	"os"
)

func helpCommand(cctx *CommandContext, args []string) {
	fmt.Println("called help")
}

func exitCommand(cctx *CommandContext, args []string) {
	os.Exit(0)
}
