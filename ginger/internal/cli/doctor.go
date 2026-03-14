package cli

import (
	"github.com/ginger-framework/ginger/internal/doctor"
)

func runDoctor(_ []string) {
	doctor.Run()
}
