package types

type WindowLayoutBuilder interface {
	Menu()
	Layout()
	Style() func()
}
