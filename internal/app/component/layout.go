package component

type Layout []ComponentType

func (l Layout) Layout() {
	for _, c := range l {
		if c != nil {
			c.Layout()
		}
	}
}

func (l Layout) Build() {
	for _, c := range l {
		if c != nil {
			c.Build()
		}
	}
}

func (l Layout) Range(f func(c ComponentType)) {
	for _, c := range l {
		if splits, ok := c.(SplittableType); ok {
			splits.Range(f)
			continue
		}
		f(c)
	}
}
