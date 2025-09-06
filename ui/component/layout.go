package component

var (
	_ Splittable    = Layout{}
	_ ComponentType = Layout{}
)

type Layout []ComponentType

func (l Layout) Layout() {
	for _, c := range l {
		if c != nil {
			c.Layout()
		}
	}
}

type Splittable interface {
	Range(func(c ComponentType))
}

func (l Layout) Range(f func(c ComponentType)) {
	for _, c := range l {
		if splits, ok := c.(Splittable); ok {
			splits.Range(f)
			continue
		}
		f(c)
	}
}
