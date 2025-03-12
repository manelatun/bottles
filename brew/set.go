package brew

type set[T comparable] map[T]struct{}

func (s set[T]) Add(keys ...T) {
	for _, key := range keys {
		s[key] = struct{}{}
	}
}

func (s set[T]) Remove(keys ...T) {
	for _, key := range keys {
		delete(s, key)
	}
}

func (s set[T]) Contains(key T) bool {
	_, exists := s[key]
	return exists
}

func (s set[T]) ContainsAll(keys ...T) bool {
	for _, key := range keys {
		_, exists := s[key]
		if !exists {
			return false
		}
	}
	return true
}

func (s set[T]) ContainsAny(keys ...T) bool {
	for _, key := range keys {
		_, exists := s[key]
		if exists {
			return true
		}
	}
	return false
}
