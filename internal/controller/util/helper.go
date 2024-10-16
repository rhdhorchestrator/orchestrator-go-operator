// util/helper.go
package util

func MakePointer[T any](t T) *T {
	return &t
}
