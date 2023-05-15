package test

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func MustNotErr(err error) {
	if err != nil {
		panic(err)
	}
}