package main

import (
	_ "bitbox-editor/lib/config"
	_ "bitbox-editor/lib/system"

	"bitbox-editor/ui"

	_ "github.com/silbinarywolf/preferdiscretegpu"
)

func main() {
	editor := ui.NewBitboxEditor()
	editor.Run()
}
