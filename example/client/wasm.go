// +build js,wasm

package main

import (
	"syscall/js"
)

var PayloadSizeMb = js.Global().Get("payloadSizeMb").Float()
var NumPingPongs = js.Global().Get("numPingPongs").Int()
