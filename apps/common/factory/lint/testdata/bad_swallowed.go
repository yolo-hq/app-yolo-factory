//go:build ignore

package testdata

func BadSwallowed() {
	_, _ = compute()
}

func compute() (int, error) {
	return 0, nil
}
