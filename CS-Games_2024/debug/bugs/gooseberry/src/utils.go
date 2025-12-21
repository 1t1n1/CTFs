package main


func Reverse[T any](input []T) []T {
	l := len(input)
	output := make([]T, l)

	for i, ele := range input {
		output[l-i-1] = ele
	}
	return output
}
