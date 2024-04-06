package component

type LifecycleComponentStack struct {
	Components []LifecycleComponent
}

func (l *LifecycleComponentStack) Start() {
	for _, c := range l.Components {
		c.Start()
	}
}

func (l *LifecycleComponentStack) AddComponent(c LifecycleComponent) {
	l.Components = append(l.Components, c)
}
