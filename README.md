## Range Over Functions

Go 1.22 added [`range over function` experiment](https://tip.golang.org/wiki/RangefuncExperiment).
Go lacks a standard iterator protocol, and this is an attempt to make one.
The idea is that you will produce the values in a function,
and every time you produce a value you pass it to a `yield` function that will handle the value.
The `yield` function call also lets you know when the iteration is done and there's no need to produce more values by returning `false`.

_NOTE: If you are familiar with Python's generators - it's the same idea._

Say you have a slice of events, and you want to filter out non-valid events.

**Listing 14: Event**

```go
21 type Event struct {
22     User string
23     Kind string
24 }
25 
26 func (e Event) Valid() bool {
27     if e.User == "" || e.Kind == "" {
28         return false
29     }
30 
31     return true
32 }
```

Listing 14 shows the `Event` type and the `Valid` method on the event.

Now you can write a generic `Filter` function:

**Listing 15: Filter**

```go
05 func Filter[T any](values []T, pred func(T) bool) func(func(T) bool) {
06     fn := func(yield func(T) bool) {
07         for _, v := range values {
08             if !pred(v) {
09                 continue
10             }
11 
12             if !yield(v) {
13                 break
14             }
15         }
16     }
17 
18     return fn
19 }
```

Listing 15 shows the `Filter` function.
On line 5 you define a generic function with an `any` constraint.
The function accepts a `values` parameter which is a slice and a `pred` which is a predicate function.
`Filter` returns a function that accepts a `yield` function to handle the value.
On line 06 you define the returned iterator function. 
On line 07 you iterate over the slice and on line 08 you skip invalid values.
On line 12 you send a valid value to the `yield` function and check if you need to continue.
Finally on line 18 you return the iterator functions.

Now let's try it out:

**Listing 16: Testing the Filter**

```go
36     events := []Event{
37         {"elliot", "login"},
38         {"", "access"},
39         {"elliot", "logout"},
40     }
41 
42     for e := range Filter(events, Event.Valid) {
43         fmt.Println(e)
44     }
```

On lines 36-40 you define the `events` slice, the event on line 38 is not valid.
On line 42 you range over `Filter` which is called with `events` and `Event.filter` as the predicate.

To run the program you need to set the `GOEXPERIMENT` environment variable to `rangefunc`

**Listing 17: Running the Program**

```
01 $ GOEXPERIMENT=rangefunc go run filter.go 
02 {elliot login}
03 {elliot logout}
```

Listing 17 shows how to run the program.
On line 01 you set the `GOEXPERIMENT` environment variable and use `go run` to run the program.
On lines 01-02 you can see the output which does not contain the invalid event.

