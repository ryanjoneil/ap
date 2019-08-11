package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ryanjoneil/ap/lsap"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "%s solves linear sum assignment problems, given a square cost matrix\n", os.Args[0])
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "\t%s < input.json > output.json\n", os.Args[0])
		fmt.Fprintf(out, "\tcat <<EOF | %s | jq\n", os.Args[0])
		fmt.Fprintf(out, "\t[\n")
		fmt.Fprintf(out, "\t\t[  90,  76,  75,  70 ],\n")
		fmt.Fprintf(out, "\t\t[  35,  85,  55,  65 ],\n")
		fmt.Fprintf(out, "\t\t[ 125,  95,  90, 105 ],\n")
		fmt.Fprintf(out, "\t\t[  45, 110,  95, 115 ]\n")
		fmt.Fprintf(out, "\t]\n")
		fmt.Fprintf(out, "\tEOF\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var c [][]int64
	if err := json.NewDecoder(os.Stdin).Decode(&c); err != nil {
		panic(err)
	}

	a := lsap.New(c)
	out := map[string]interface{}{
		"assignment": a.Assign(),
		"cost":       a.Cost(),
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}
