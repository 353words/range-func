## Range Over Functions
+++
title = "Range Over Functions"
date = "FIXME"
tags = ["golang"]
categories = ["golang"]
url = "FIXME"
author = "mikit"
+++

Go 1.22 added [`range over function` experiment](https://tip.golang.org/wiki/RangefuncExperiment).
In this blog post we'll discuss the motivation for adding this experiment and see some examples on how to use it.

_NOTE: In order to run the code you need to set the `GOEXPERIMENT` environment variable to `rangefunc`._

Go lacks a standard iterator protocol, and this is an attempt to make one.
Before we dive into details, let's take a look at two common patterns in iteration: separating iterator from container and inversion of control.

_NOTE_: If you are familiar with Python's iterators and generators, this blog post will seem familiar._

### Separating Iterator from Container

Our example container is a `stack` that's  implemented using a linked list.

**Listing 1: Stack**

```go
10 type node[T any] struct {
11     value T
12     next  *node[T]
13 }
14 
15 type Stack[T any] struct {
16     head *node[T]
17 }
18 
19 func (s *Stack[T]) Push(v T) {
20     s.head = &node[T]{v, s.head}
21 }
22 
23 var ErrEmpty = errors.New("empty stack")
24 
25 func (s *Stack[T]) Pop() (T, error) {
26     if s.head == nil {
27         var v T
28         return v, ErrEmpty
29     }
30 
31     n := s.head
32     s.head = s.head.next
33     return n.value, nil
34 }
```

Listing 1 shows a stack implementation.
On lines 10-13 we define a `node` struct with a `value` and `next` field.
On lines 15-17 we define a `Stack` struct that head a `head` node.
On lines 19-21 we define the `Push` function and on lines 23-34 we define `Pop`.

We don't want the `Stack` to keep track of iteration location.
If we have more than one iteration going at the same time, the bookkeeping of where is each iteration can become complex.
Instead, we are going to define a iterator that is responsible for a single iteration.

**Listing 2: StackIterator**

```go
36 func (s *Stack[T]) Items() *StackIterator[T] {
37     return &StackIterator[T]{s.head}
38 }
39 
40 type StackIterator[T any] struct {
41     node *node[T]
42 }
43 
44 func (s *StackIterator[T]) Next() (T, bool) {
45     if s.node == nil {
46         var v T
47         return v, false
48     }
49 
50     n := s.node
51     s.node = s.node.next
52     return n.value, true
53 }
```

Listing 2 shows a stack iterator.
On lines 36-38 we define `Items` that returns a `StackIterator`
On lines 41-43 we define the `StackIterator` struct that holds the current node.
On lines 45-54 we define the `Next` method that will return the next value in the stack.

**Listing 3: Using the Iterator**

```go
128     it := s.Items()
129     for v, ok := it.Next(); ok; v, ok = it.Next() {
130         fmt.Println(v)
131     }
```

Listing 3 shows how to use `Next`.
On line 128 we create an iterator using `Item`.
On lines 129-131 we iterate over the items in the stack using `it.Next`.
When `it.Next()` return `false` for the second value it means there are no more items.

### Inversion of Control

Say you want to print all the items in the stack, you can write the following code:

**Listing 4: PrintItems**

```go
55 func (s *Stack[T]) PrintItems() {
56     for n := s.head; n != nil; n = n.next {
57         fmt.Println(n.value)
58     }
59 }
```

Listing 4 shows the `PrintItems` method.
One line 56 we use a `for` loop to iterate over the nodes and on line 54 we print the values.

Now, what if instead of printing you want to save the values to a file, or maybe send them back in an HTTP request handler?
You're not going to write a different function to each scenario.
The idea is that you pass to the function that does the iteration another function that will handle the values.
Let's call this function `yield`.

**Listing 5: Do**

```go
61 func (s *Stack[T]) Do(yield func(v T)) {
62     for n := s.head; n != nil; n = n.next {
63         yield(n.value)
64     }
65 }
```

Listing 5 shows the `Do` method.
On line 61 we define the method that accepts a `yield` function that will handle the values.
On line 62 we iterate over the nodes and then on line 63 we pass the current value to the `yield` function.

**Listing 6: Using `Do`

```go
134     s.Do(func(n int) {
135         fmt.Println(n)
136     })
```

Listing 6 shows how to use the `Do` method.
On line 134 we call do with an anonymous function that accepts an `int` and prints it.

What happens if we want to stop the iteration after the first 5 values?
Currently, we can't, but we can have the `yield` function return a boolean value indicating that we should stop the iteration.
Which brings us to the topic at hand: range-over function.


### iter.Seq

Once you set the `GOEXPERIMENT` environment variable to `rangefunc`, go exposes an `iter` package that defines two types:

**Listing 7: iter.Seq and iter.Seq2

```go
type Seq[V any] func(yield func(V) bool)
type Seq2[K, V any] func(yield func(K, V) bool)
```

Listing 7 shows the iter.Seq & iter.Seq2 types.

Let's start with `iter.Seq`, it's a function that accepts a `yield` function that accepts a value and return a `bool`.
It's very much like our `Stack.Do` method above, but now the `for` loop supports it.

**Listing 8: Iter**

```go
67 func (s *Stack[T]) Iter() func(func(T) bool) {
68     iter := func(yield func(T) bool) {
69         for n := s.head; n != nil; n = n.next {
70             if !yield(n.value) {
71                 return
72             }
73         }
74     }
75 
76     return iter
77 }
```

Listing 6 shows the `Iter` method.
On line 67 we define `Iter`, it returns a function that matches `iter.Seq` type.
One line 68 we define the function that is returned from `Iter`.
On line 69 we use a `for` loop to iterate over the stack nodes.
On line 70 we pass the current value to the `yield` function, and if `yield` return `false` we stop the iteration on line 71.
Finally, on line 76 we return the iteration function.

**Listing 9: Using Iter**

```go
139     for v := range s.Iter() {
140         fmt.Println(v)
141     }
```

Listing 9 shows how to use the `Iter` method.
One line 139 we use a regular `for` loop with a `range`.
On line 140 we print the values.
The loop we are running goes through all the items in the stack, so `yield` will always return `true`.
If we break inside the `for` loop, say after the second value, `yield` will return `false` and the iteration will terminate.


In some cases, we want two variables on the left side of the `:=`.
For example getting the index of the value as well as the value, or in the case of a map, both the key and the value.
For these cases, we can use `iter.Seq2`

**Listing 10: Iter2**

```go
79 func (s *Stack[T]) Iter2() func(func(int, T) bool) {
80     iter := func(yield func(int, T) bool) {
81         for i, n := 0, s.head; n != nil; i, n = i+1, n.next {
82             if !yield(i, n.value) {
83                 return
84             }
85         }
86     }
```

Listing 10 shows the `Iter2` method.
On line 79 we define `Iter2` that return a function that accepts a function that gets two parameters: an `int` for the index and the value.
On line 80 we define the return function.
On line 81 we run a `for` loop over the nodes and on line 82 we yield the position and the value.

**Listing 11: Using Iter2**

```go
144     for i, v := range s.Iter2() {
145         fmt.Println(i, v)
146     }
```

Listing 11 shows how to use `Iter2`.
On line 144 we run a `for` loop with two values before the `:=`.
On line 145 we print the index and the value.

### Pulling Values

Once we have generic iteration, we'd want to define function that work with them.
Let's try to think about a `Max` function, how will it work?
There are two problems we need to solve with `Max`: Err if the sequence is empty and get the first value.

If you work with slices, these issues are easy.
To check for empty sequence, use `len`.
To get the first values, use `s[0]`.

With sequence, we don't know the length (and they can be potentially infinite).
And, we can't access the first value without some logic inside the `for`.
For these cases, the `iter` package defines the `Pull` and `Pull2` functions.


**Listing 12: Pull and Pull2**

```go
func Pull[V any](seq Seq[V]) (next func() (V, bool), stop func())
func Pull2[K, V any](seq Seq2[K, V]) (next func() (K, V, bool), stop func())
```

Listing 12 shows the `Pull` and `Pull2` functions.
`Pull` get a `Seq` and returns two functions: `next` to pull the next value from `seq` and `stop` to signal stop of iteration.
When you call `stop`, the `yield` passed to `seq` will return `false` - signaling it should stop the iteration.
`Pull2` does that same, only with `Seq2` that return two values.

Let's see how we can use `Pull` to write `Max`.

**Listing 13: Max**

```go
91 func Max[T cmp.Ordered](seq iter.Seq[T]) (T, error) {
92     pull, stop := iter.Pull(seq)
93     defer stop()
94 
95     m, ok := pull()
96     if !ok {
97         return m, fmt.Errorf("Max of empty sequence")
98     }
99 
100     for v, ok := pull(); ok; v, ok = pull() {
101         if v > m {
102             m = v
103         }
104     }
105 
106     return m, nil
107 }
```

Listing 13 shows the `Max` function.
On line 91 we define `Max1 that gets an `iter.Seq` as a parameter and return a value and an error.
On line 92 we use `iter.Pull` to get `pull` that return the next value and `stop` that signals to stop the iteration.
On line 93 we defer `stop` to signal that we want to stop the iteration.
One line 95 we use `pull` to get the first value.
On line 96 we check if there are no more values with `ok` and if it's `false` we return an error on line 97.
On line 100 we use a `for` loop to iterate over the rest of the values.
On line 101 we check if the current value is bigger than `m` (the current maximaum) and if so we update `m` on line 102.
Finally, on line 106 we return the maximal value and `nil` for error.

**Listing 14: Using `Max`**

```go
148     m, err := Max(s.Iter())
149     if err != nil {
150         fmt.Println("ERROR:", err)
151     } else {
152         fmt.Println("max:", m)
153     }
```

Listing 14 show how to use `Max`.
On line 148 we call `Max` on `s.Iter`.
On line 149 we check for error and if so we print the error on line 150.
Otherwise, on line 152 we print the maximal value.


### Conclusion

The range-over function experimenet tries to give Go a general way to provide custom iterations.
The users of `iter.Seq` and `iter.Seq2` will use the familiar `for` loop, the burden of implementing falls on the library writer.

I hope I shed some light on why we have this experiment and also on how to use it.
You can read more about it on [the wiki page](https://go.dev/wiki/RangefuncExperiment).

I'd love to hear from you if you have more ideas on how to use this experiment.
Contact me at [miki@ardanlabs.com](mailto:miki@ardanlabs.com).
