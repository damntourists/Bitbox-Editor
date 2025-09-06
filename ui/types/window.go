package types

type WindowType interface {
	Title() string
	Icon() string

	Open()
	IsOpen() bool
	Close()

	Style() func()
	Menu()
	Layout()

	Build()
}
