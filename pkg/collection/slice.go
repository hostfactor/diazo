package collection

import "fmt"

type FinderFunc[T any] func(t T) bool

type IndexFunc[K comparable, V any] func(v V) K

type MapFunc[F, T any] func(f F) T

func Find[T any](sl []T, f FinderFunc[T]) (out T) {
	for _, v := range sl {
		if f(v) {
			return v
		}
	}

	return
}

func Map[F, T any](sl []F, f MapFunc[F, T]) []T {
	out := make([]T, 0, len(sl))
	for _, v := range sl {
		out = append(out, f(v))
	}

	return out
}

func MapString[F fmt.Stringer](sl []F) []string {
	return Map(sl, func(f F) string {
		return f.String()
	})
}

func Filter[T any](sl []T, f FinderFunc[T]) []T {
	out := make([]T, 0, len(sl))
	for i, v := range sl {
		if f(v) {
			out = append(out, sl[i])
		}
	}

	return out
}

func IndexStringer[V fmt.Stringer](sl []V) map[string]V {
	return Index[string, V](sl, func(v V) string {
		return v.String()
	})
}

func Index[K comparable, V any](sl []V, f IndexFunc[K, V]) map[K]V {
	m := map[K]V{}
	for i, v := range sl {
		m[f(v)] = sl[i]
	}

	return m
}

func Any[T any](sl []T, f FinderFunc[T]) bool {
	for _, v := range sl {
		if f(v) {
			return true
		}
	}
	return false
}

func All[T any](sl []T, f FinderFunc[T]) bool {
	out := true
	for _, v := range sl {
		out = out && f(v)
	}
	return out
}
