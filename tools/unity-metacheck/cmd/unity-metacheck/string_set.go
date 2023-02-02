package main

type unit struct{}
type stringSet map[string]unit

func (s stringSet) Add(x string) {
	s[x] = unit{}
}

func (s stringSet) Remove(x string) {
	delete(s, x)
}

func (s stringSet) Contains(x string) bool {
	_, ok := s[x]
	return ok
}
