package types

type ModalType interface {
	Title() string
	Icon() string

	Open()
	IsOpen() bool
	Close()

	Menu()
	Layout()

	PreBuild()
	PostBuild()
	Build()
}
