package main

import "fmt"

func Filter[T any](values []T, pred func(T) bool) func(func(T) bool) {
	fn := func(yield func(T) bool) {
		for _, v := range values {
			if !pred(v) {
				continue
			}

			if !yield(v) {
				break
			}
		}
	}

	return fn
}

type Event struct {
	User string
	Kind string
}

func (e Event) Valid() bool {
	if e.User == "" || e.Kind == "" {
		return false
	}

	return true
}

func main() {

	events := []Event{
		{"elliot", "login"},
		{"", "access"},
		{"elliot", "logout"},
	}

	for e := range Filter(events, Event.Valid) {
		fmt.Println(e)
	}
}
