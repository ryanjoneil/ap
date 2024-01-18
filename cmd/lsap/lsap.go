package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ryanjoneil/ap/lsap"
)

func usage() {
	out := flag.CommandLine.Output()
	msg := " solves linear sum assignment problems, given a square cost matrix."
	fmt.Fprint(out, os.Args[0])
	fmt.Fprintln(out, msg)
	fmt.Fprint(out, "\nUsage:\n\n")
	fmt.Fprintf(out, "%s < input.json -dual -rc > output.json\n\n", os.Args[0])
	fmt.Fprintf(out, "cat <<EOF | %s | jq\n", os.Args[0])
	fmt.Fprintf(out, "[\n")
	fmt.Fprintf(out, "\t[  90,  76,  75,  70 ],\n")
	fmt.Fprintf(out, "\t[  35,  85,  55,  65 ],\n")
	fmt.Fprintf(out, "\t[ 125,  95,  90, 105 ],\n")
	fmt.Fprintf(out, "\t[  45, 110,  95, 115 ]\n")
	fmt.Fprintf(out, "]\n")
	fmt.Fprintf(out, "EOF\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	cycles := flag.Bool("cycles", false, "output cyclic assignment form")
	dual := flag.Bool("dual", false, "output dual prices")
	matrix := flag.Bool("matrix", false, "output matrix assignment form")
	rc := flag.Bool("rc", false, "output reduced cost matrix")
	flag.Usage = usage
	flag.Parse()

	var c [][]int64
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		panic(err)
	}

	a := lsap.New(c)
	p := a.Assign()
	out := map[string]any{
		"permutation": p,
		"cost":        a.Cost(),
	}

	if *cycles {
		out["cycles"] = p.Cycles()
	}

	if *dual {
		out["dual"] = a.DualPrices()
	}

	if *matrix {
		out["matrix"] = p.Matrix()
	}

	if *rc {
		rcMatrix := make([][]int64, len(p))
		for u := range p {
			rcRow := make([]int64, len(p))
			for v := range p {
				rcRow[v] = a.ReducedCost(u, v)
			}
			rcMatrix[u] = rcRow
		}
		out["rc"] = rcMatrix
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}
