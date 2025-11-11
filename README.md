# golazy

golazy is a tiny, generic Go helper for lazy-loading values.
It provides a small `Lazy[T]` abstraction that can load values on demand using a
user-provided loader function. It supports optional per-context TTL caching and
preloaded values.

## Installation

Requires Go 1.19+ 

```bash
go get github.com/duhnnie/golazy
```

## Features

- Generic `Lazy[T]` interface with `Value`, `Clear`, and `ClearAll`.
- Create a lazy loader with `WithLoader` or `WithLoaderTTL` (TTL-enabled).
- Preload an initial value for a particular context with `Preloaded`/`PreloadedTTL`.
- `Static` for values that never change (useful in tests or constant wiring).

## Usage

### Basic usage

Assume you have a function that loads some config or remote resource:

```go
package main

import (
    "fmt"
    "time"

    "github.com/duhnnie/golazy"
)

func main() {
    loader := func(ctx any) (string, error) {
        // pretend to load from remote or disk using ctx as key
        key := ctx.(string)
        return "value-for-" + key, nil
    }

    lazy := golazy.WithLoader[string](loader)

    v, err := lazy.Value("alpha")
    if err != nil {
        panic(err)
    }
    fmt.Println(v) // prints: value-for-alpha

    // Clear cached value for a key
    lazy.Clear("alpha")
}
```

### Using contexts

You can use the same `golazy.Lazy` var to return a different value for every different context:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/duhnnie/golazy"
)

func dataLoader(ctx any) ([]int, error) {
	if n, ok := ctx.(int); !ok {
		return nil, errors.New("context should be an int")
	} else {
		return []int{n + 0, n + 1, n + 2}, nil
	}
}

func main() {
	data := golazy.WithLoader(dataLoader)
	v, err := data.Value(10) // context 10 returns: 10, 11, 12
	if err != nil {
		panic(err)
	}

	fmt.Println(v)

	v2, err := data.Value(100) // context 100 returns: 100, 101, 102
	if err != nil {
		panic(err)
	}

	fmt.Println(v2)

	data.Clear(10)  // Clears cache for the '10' context
	data.ClearAll() // Clears cache for all contexts
}
```

### Using TTL caching

TTL caching allows you to set an expiration time for your data, so any access after that time loader function will be invoked again.

```go
loader := func(ctx any) (int, error) {
    return 42, nil
}

lazyTTL := golazy.WithLoaderTTL[int](loader, 5*time.Second)
// value will be cached per-context for 5s
// also try golazy.PreloadedTTL()
```

### Preloaded value

Sometimes you already have the value intended to be loaded, for that case use a preloaded-lazy.

```go
pre := golazy.Preloaded[string](loader, "initial", "ctx-1")
v, _ := pre.Value("ctx-1") // returns "initial"
```

### Static value

When the value of your variable is never gonna change use a static-lazy.

```go
s := golazy.Static(123)
v, _ := s.Value(nil) // always returns 123
```

## Real-life example

Usually structs have properties that are lazy loaded. Next code shows how to use `golazy` in that situation:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/duhnnie/golazy"
)

// Define Student struct
type Student struct {
	ID   int
	Name string
}

// Define Couse struct
type Course struct {
	ID       int
	Topic    string
	students golazy.Lazy[[]*Student]
}

func (c *Course) GetStudents() ([]*Student, error) {
	return c.students.Value(c)
}

// Define our loader function, it accepts any value as context
func studentsLoader(ctx any) ([]*Student, error) {
	if course, ok := ctx.(*Course); !ok {
		return nil, errors.New("invalid context")
	} else {
		fmt.Printf("loading studets enrolled in course with id: %d\n", course.ID)

		// here you can fetch students from a database or any other external source
		// simulate time consuming task

		time.Sleep(1 * time.Second)

		return []*Student{
			{1, "Cooper"},
			{2, "Monica"},
			{3, "Mr. Ed"},
		}, nil
	}

}

// Define our Course factory function
func NewCourse(id int, topic string, studentsLazyLoader golazy.LazyFunc[[]*Student]) *Course {
	return &Course{
		ID:       id,
		Topic:    topic,
		students: golazy.WithLoader(studentsLazyLoader),
	}
}

func main() {
    c := NewCourse(23, "Algorithms", studentsLoader)

	s, err := c.GetStudents()
	if err != nil {
		panic(err)
	}

	for _, student := range s {
		fmt.Printf("Student %d: %s\n", student.ID, student.Name)
	}
}
```

## Clear Cache

Variable value is only load once, unless you get some error from loading function. All next calls to `Value(ctx)` will return the value loaded before.

You might want to clear that cached value, to load fresh data next time `Value(ctx)` is called. For that you have two methods:

```go
lazyVar.Clear(ctx) // Clears cached value on supplied context
lazyVar.ClearAll() // Clears cached value for all contexts
```

## Notes

- The package uses `any` for context/key values; you can use strings, structs, or
  other comparable types as keys. If you need complex keying, use a dedicated
  key type.
- The loader receives the same `ctx` value used as the cache key; use it if
  your loader needs it.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. Check the [guidelines](./CONTRIBUTING.md).

## License

[MIT License](LICENSE)
