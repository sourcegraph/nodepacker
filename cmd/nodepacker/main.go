package main

import (
	"nodepacker"
	"github.com/c-bata/go-prompt"
)

func main() {
	crh := nodepacker.NewCommandReplHandler()

	pr := prompt.New(crh.ExecuteStatement,
		crh.Completer,
		prompt.OptionPrefix("nodepacker> "),
		prompt.OptionTitle("nodepacker prompt"))

	pr.Run()
}
