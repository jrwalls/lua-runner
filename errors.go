package main

import "fmt"

type (
	LuaError struct {
		err error
	}
)

func (e *LuaError) Error() string {
	return fmt.Sprintf("error running lua function: %s", e.err)
}

func (e *LuaError) Unwrap() error {
	return e.err
}
