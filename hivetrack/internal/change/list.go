package change

// List tracks which fields were changed on a model instance.
// T is a model-specific change constant type (iota).
type List[T comparable] struct {
	changes map[T]struct{}
}

func NewList[T comparable]() List[T] {
	return List[T]{changes: make(map[T]struct{})}
}

func (l *List[T]) TrackChange(c T) {
	if l.changes == nil {
		l.changes = make(map[T]struct{})
	}
	l.changes[c] = struct{}{}
}

func (l *List[T]) HasChange(c T) bool {
	if l.changes == nil {
		return false
	}
	_, ok := l.changes[c]
	return ok
}

func (l *List[T]) HasChanges() bool {
	return len(l.changes) > 0
}

func (l *List[T]) GetChanges() []T {
	result := make([]T, 0, len(l.changes))
	for c := range l.changes {
		result = append(result, c)
	}
	return result
}

func (l *List[T]) ClearChanges() {
	l.changes = make(map[T]struct{})
}
