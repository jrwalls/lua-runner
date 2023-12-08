package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"strconv"
	"strings"
)

const (
	BaseToString LuaFunc = "tostring"
	BaseToNumber LuaFunc = "tonumber"
	BaseError    LuaFunc = "error"
	BaseType     LuaFunc = "type"
	BasePrint    LuaFunc = "print"
)

var (
	baseLuaFns = map[LuaFunc]lua.LGFunction{
		BaseToString: baseToString,
		BaseToNumber: baseToNumber,
		BaseError:    baseError,
		BaseType:     baseType,
		BasePrint:    basePrint,
	}

	baseLuaLibs = map[LuaLib]lua.LGFunction{
		lua.LoadLibName:      lua.OpenPackage,
		lua.BaseLibName:      lua.OpenBase,
		lua.TabLibName:       lua.OpenTable,
		lua.IoLibName:        lua.OpenIo,
		lua.OsLibName:        lua.OpenOs,
		lua.StringLibName:    lua.OpenString,
		lua.MathLibName:      lua.OpenMath,
		lua.DebugLibName:     lua.OpenDebug,
		lua.ChannelLibName:   lua.OpenChannel,
		lua.CoroutineLibName: lua.OpenCoroutine,
	}
)

func baseToString(L *lua.LState) int {
	v1 := L.CheckAny(1)
	L.Push(L.ToStringMeta(v1))
	return 1
}

func baseToNumber(L *lua.LState) int {
	base := L.OptInt(2, 10)
	noBase := L.Get(2) == lua.LNil

	switch lv := L.CheckAny(1).(type) {
	case lua.LNumber:
		L.Push(lv)
	case lua.LString:
		str := strings.Trim(string(lv), " \n\t")
		if strings.Index(str, ".") > -1 {
			if v, err := strconv.ParseFloat(str, lua.LNumberBit); err != nil {
				L.Push(lua.LNil)
			} else {
				L.Push(lua.LNumber(v))
			}
		} else {
			if noBase && strings.HasPrefix(strings.ToLower(str), "0x") {
				base, str = 16, str[2:] // Hex number
			}
			if v, err := strconv.ParseInt(str, base, lua.LNumberBit); err != nil {
				L.Push(lua.LNil)
			} else {
				L.Push(lua.LNumber(v))
			}
		}
	default:
		L.Push(lua.LNil)
	}
	return 1
}

func baseError(L *lua.LState) int {
	obj := L.CheckAny(1)
	level := L.OptInt(2, 1)
	L.Error(obj, level)
	return 0
}

func basePrint(L *lua.LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		fmt.Print(L.ToStringMeta(L.Get(i)).String())
		if i != top {
			fmt.Print("\t")
		}
	}
	fmt.Println("")
	return 0
}

func baseType(L *lua.LState) int {
	L.Push(lua.LString(L.CheckAny(1).Type().String()))
	return 1
}
