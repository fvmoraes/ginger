package cli

import (
	"fmt"
	"os"
	"os/exec"
)

func runRun(args []string) {
	cmd := exec.Command("go", append([]string{"run", "./cmd/app"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		os.Exit(1)
	}
}

func runBuild(args []string) {
	output := "./bin/app"
	if len(args) > 0 {
		output = args[0]
	}
	cmd := exec.Command("go", "build", "-o", output, "./cmd/app")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Built: %s\n", output)
}
