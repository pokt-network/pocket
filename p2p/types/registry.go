package types

import sync "sync"

type Registry struct {
	sync.Mutex
	maxcap   uint32
	elements map[string]interface{} // TODO(derrandz): If not performant, replace with unsafe pointer
	factory  func() interface{}     // generates objects on non existing ids in the registry (check the get method)
}

func (r *Registry) Get(id string) (interface{}, bool) {
	defer r.Unlock()
	r.Lock()

	var obj interface{}
	var exists bool

	obj, exists = r.elements[id]
	if !exists {
		// create a new iopipe
		// TODO(derrandz): add logic to check for maxcap if reached
		// TODO(derrandz): add logic to swap old connections for new one on maxcap reached
		obj = r.factory()
		r.elements[id] = obj
	}

	return obj, exists
}

func (r *Registry) Find(id string) (interface{}, bool) {
	defer r.Unlock()
	r.Lock()

	el, exists := r.elements[id]
	return el, exists
}

func (r *Registry) Peak(id string) bool {
	defer r.Unlock()
	r.Lock()

	_, exists := r.elements[id]
	return exists
}

func (r *Registry) Remove(id string) (bool, error) {
	defer r.Unlock()
	r.Lock()

	panic("Not implemented")
	return false, nil
}

func (r *Registry) Capacity() uint32 {
	return r.maxcap
}

func NewRegistry(cap uint32, factory func() interface{}) *Registry {
	return &Registry{
		maxcap:   cap,
		elements: make(map[string]interface{}),
		factory:  factory,
	}
}
