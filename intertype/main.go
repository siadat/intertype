package main

import (
	"github.com/siadat/intertype"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(intertype.MyAnalyzer)
}
