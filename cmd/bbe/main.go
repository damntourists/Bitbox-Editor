package main

import (
	_ "bitbox-editor/lib/config"
	"bitbox-editor/ui"

	_ "github.com/silbinarywolf/preferdiscretegpu"
)

func main() {
	editor := ui.NewBitboxEditor()
	editor.Run()
}
