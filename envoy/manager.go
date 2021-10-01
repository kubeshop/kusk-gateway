package envoy

type Manager struct{}

func NewManager() *Manager {
	return new(Manager)
}
