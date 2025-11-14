package component

type ComponentType interface {
	Build()
	Destroy()
	Layout()
	SendUpdate(cmd UpdateCmd)

	HandleGlobalUpdate(cmd UpdateCmd) bool
}

type LayoutBuilderType interface {
	Layout()
}

type SplittableType interface {
	Range(func(c ComponentType))
}
