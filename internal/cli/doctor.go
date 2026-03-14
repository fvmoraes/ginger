package cli

import (
	"github.com/fvmoraes/ginger/internal/doctor"
)

func runDoctor(_ []string) {
	doctor.Run()
}
