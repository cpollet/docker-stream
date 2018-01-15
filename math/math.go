package math

func Abs(x int) int {
	return Max(x, 0)
}

func Max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}
