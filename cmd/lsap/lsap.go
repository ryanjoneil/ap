package main

import (
	"encoding/json"
	"os"

	"github.com/ryanjoneil/ap/lsap"
)

func main() {
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
