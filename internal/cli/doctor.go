package cli

import (
	"os"

	"github.com/fvmoraes/ginger/internal/doctor"
)

func runDoctor(_ []string) {
	if !doctor.Run() {
		os.Exit(1)
	}
}
