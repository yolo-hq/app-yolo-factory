//go:build ignore

package testdata

import "fmt"

func DoSomething() error {
	val, err := compute()
	if err != nil {
		return err
	}
	fmt.Println(val)
	return nil
}

func compute() (int, error) {
	return 42, nil
}
