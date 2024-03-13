package main

import (
	"fmt"
	"time"
)

type item[V any] struct {
	val  V
	time time.Time
}

type Cache[K comparable, V any] map[K]item[V]

func (c Cache[K, V]) Iter() func(func(K, V) bool) {
	fn := func(yield func(K, V) bool) {
	}

	return fn
}

func main() {
	fmt.Println("Go!")
}
