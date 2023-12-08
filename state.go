package main

import (
	"errors"
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"sync"
)

type (
	LuaResource interface {
		IsLuaResource()
	}

	LuaRunner struct {
		ls           *lua.LState
		cleanLs      *lua.LState
		lsLock       sync.Mutex
		skipOpenLibs bool
		isLsClean    bool
	}

	LuaLib            string
	LuaFunc           string
	LuaFunctionName   string
	LuaFunctionCode   string
	LuaFunctionRetNum uint
)

func (lr *LuaRunner) TEST() {}

func (l LuaLib) IsLuaResource()  {}
func (f LuaFunc) IsLuaResource() {}

func NewLuaRunner(skipOpenLibs bool, resources ...LuaResource) (*LuaRunner, error) {
	lr := &LuaRunner{
		ls:           lua.NewState(lua.Options{SkipOpenLibs: skipOpenLibs}),
		lsLock:       sync.Mutex{},
		skipOpenLibs: false,
		isLsClean:    true, // Start with a clean state
	}

	for _, r := range resources {
		switch key := r.(type) {
		case LuaLib:
			lib, found := baseLuaLibs[key]
			if !found {
				return nil, fmt.Errorf("unsupported lua lib: %s", r)
			}

			if err := lr.loadLuaLib(string(key), lib); err != nil {
				return nil, fmt.Errorf("cannot load lua.%s base lib", r)
			}

		case LuaFunc:
			fn, found := baseLuaFns[key]
			if !found {
				return nil, fmt.Errorf("unsupported lua function: %s", r)
			}

			lr.ls.SetGlobal(string(key), lr.ls.NewFunction(fn))
		}
	}

	lr.cleanLs = lr.ls
	return lr, nil
}

func (lr *LuaRunner) Run(luaFnName LuaFunctionName, luaFn LuaFunctionCode, retNum LuaFunctionRetNum, args ...any) ([]any, error) {
	lr.lsLock.Lock()
	defer lr.lsLock.Unlock()

	if !lr.isLsClean {
		lr.refreshLState()
	}

	if err := lr.ls.DoString(string(luaFn)); err != nil {
		var luaErr *lua.ApiError
		if errors.As(err, &luaErr) && luaErr.Cause != nil {
			return nil, fmt.Errorf("lua script error: %s", luaErr.Cause.Error())
		}

		return nil, fmt.Errorf("cannot load string into lua state: %w", err)
	}

	luaArgs, err := lr.convertArgsToLua(args)
	if err != nil {
		return nil, fmt.Errorf("runLua: %w", err)
	}

	if err = lr.ls.CallByParam(lua.P{Fn: lr.ls.GetGlobal(string(luaFnName)), NRet: int(retNum), Protect: true}, luaArgs...); err != nil {
		return nil, &LuaError{err: fmt.Errorf("cannot run lua fn %s: %w", luaFnName, err)}
	}

	results := make([]any, retNum)
	for i := 0; i < int(retNum); i++ {
		// lua stack is 1-indexed and negative indices count from top of stack
		lv := lr.ls.Get(-1 - i)
		goVal, cErr := luaToGoType(lv)
		if cErr != nil {
			return nil, fmt.Errorf("lua - go conversion error: %w", err)
		}

		results[i] = goVal
	}

	return results, nil
}

func (lr *LuaRunner) convertArgsToLua(args []any) ([]lua.LValue, error) {
	var luaArgs []lua.LValue
	for _, arg := range args {
		lVal, err := goToLuaType(lr.ls, arg)
		if err != nil {
			return nil, fmt.Errorf("cannot convert arg %T to lua type: %w", arg, err)
		}

		luaArgs = append(luaArgs, lVal)
	}

	return luaArgs, nil
}

func (lr *LuaRunner) refreshLState() {
	lr.ls.Close()
	lr.ls = lr.cleanLs
	lr.isLsClean = true
}

func (lr *LuaRunner) loadLuaLib(name string, fn lua.LGFunction) error {
	return lr.ls.CallByParam(lua.P{
		Fn:      lr.ls.NewFunction(fn),
		NRet:    0,
		Protect: true,
	}, lua.LString(name))
}
