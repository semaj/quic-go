// +build !js,!wasm

package main

import (
	"os"
	"strconv"
)

var PayloadSizeMb, _ = strconv.ParseFloat(os.Args[1], 64)
var NumPingPongs, _ = strconv.Atoi(os.Args[2])
