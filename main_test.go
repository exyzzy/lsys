package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/exyzzy/lsys/lsys"
)

func TestRenderAllLsys(t *testing.T) {
	fmt.Fprintln(os.Stdout, "All LSys Fractals:")
	fmt.Fprintln(os.Stdout, lsys.RenderAllLsys(os.Stdout))
	return
}
