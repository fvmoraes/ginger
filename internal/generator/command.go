// Package generator — command subcommand generator for --cli projects.
package generator

import (
	"path/filepath"
)

const commandTmpl = generatedGoFileHeader + `package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var {{.Name}}Cmd = &cobra.Command{
	Use:   "{{.Slug}}",
	Short: "{{.NameTitle}} subcommand",
	RunE: func(cmd *cobra.Command, args []string) error {
		// PT-BR: Implemente a lógica do subcomando aqui.
		// EN: Implement subcommand logic here.
		fmt.Println("{{.Slug}}: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand({{.Name}}Cmd)
}
`

const commandTestTmpl = generatedGoFileHeader + `package commands

import (
	"testing"
)

func Test{{.NameTitle}}Cmd(t *testing.T) {
	cmd := {{.Name}}Cmd
	if cmd.Use != "{{.Slug}}" {
		t.Fatalf("expected Use %q, got %q", "{{.Slug}}", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Fatal("expected RunE to be set")
	}
}
`

// Command generates internal/commands/<name>.go and its test for a --cli project.
func Command(name string) error {
	data := newData(name)

	if err := generate(
		filepath.Join("internal", "commands", data.FileName+".go"),
		commandTmpl,
		data,
	); err != nil {
		return err
	}

	return generate(
		filepath.Join("internal", "commands", data.FileName+"_test.go"),
		commandTestTmpl,
		data,
	)
}
