package types

type LayoutBuilder interface {
	Menu()
	Layout()
	Style() func()
}
