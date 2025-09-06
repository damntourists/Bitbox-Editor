package main

import (
	"bitbox-editor/ui"

	_ "github.com/silbinarywolf/preferdiscretegpu"
)

func main() {
	editor := ui.NewBitboxEditor()
	editor.Run()
}
