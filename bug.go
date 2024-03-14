package main

import (
	"cmp"
	"fmt"
	"iter"
)

func Ints(n int) func(func(int) bool) {
	fn := func(yield func(int) bool) {
		for i := range n {
			if !yield(i) {
				return
			}
		}
	}

	return fn
}

func Max[T cmp.Ordered](seq iter.Seq[T]) (T, error) {
	pull, stop := iter.Pull(seq)
	defer stop()

	m, ok := pull()
	if !ok {
		return m, fmt.Errorf("Max of empty sequence")
	}
	fmt.Println(">>> m:", m)

	//for v := range seq {
	for v, ok := pull(); ok; v, ok = pull() {
		fmt.Println(">>> v:", v)
		if v > m {
			m = v
		}
	}

	return m, nil
}

func main() {
	fmt.Println(Max(Ints(3)))
}
