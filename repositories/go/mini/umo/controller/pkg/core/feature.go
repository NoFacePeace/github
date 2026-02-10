package core

type FeatureManager struct {
}

func (f *FeatureManager) IsEnabled(name string) bool {
	return false
}
