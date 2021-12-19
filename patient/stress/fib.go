package stress

// simple fib calculation that uses up some CPU cycles
// make it recursive so it's extra slow :D
func Fib(n int) int {
	if n <= 1 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}
