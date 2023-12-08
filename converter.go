package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"reflect"
)

func goToLuaType(ls *lua.LState, v any) (lua.LValue, error) {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Invalid:
		return lua.LNil, nil
	case reflect.Ptr:
		if rv.IsNil() {
			return lua.LNil, nil
		}

		return goToLuaType(ls, rv.Elem().Interface())
	case reflect.String:
		return lua.LString(rv.String()), nil
	case reflect.Bool:
		return lua.LBool(rv.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(rv.Int()), nil
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(rv.Float()), nil
	case reflect.Map:
		table := ls.NewTable()
		it := rv.MapRange()

		for it.Next() {
			k := it.Key()

			// Ensure k is a string
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("map k is not a string, got: %s", k.String())
			}

			luaV, err := goToLuaType(ls, it.Value())
			if err != nil {
				return nil, err
			}

			table.RawSetString(k.String(), luaV)
		}

		return table, nil
	case reflect.Struct:
		table := ls.NewTable()
		for i := 0; i < rv.NumField(); i++ {
			fn := rv.Type().Field(i).Name
			fv, err := goToLuaType(ls, rv.Field(i).Interface())
			if err != nil {
				return nil, err
			}

			table.RawSetString(fn, fv)
		}

		return table, nil
	case reflect.Slice, reflect.Array:
		table := ls.NewTable()
		for i := 0; i < rv.Len(); i++ {
			ev, err := goToLuaType(ls, rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}

			table.Append(ev)
		}

		return table, nil
	}

	return nil, fmt.Errorf("unsupported type for go to lua conversion: %s", rv.Type())
}

func luaToGoType(lv lua.LValue) (any, error) {
	if lv == lua.LNil {
		return nil, nil
	}

	switch converted := lv.(type) {
	case lua.LString:
		return string(converted), nil
	case lua.LNumber:
		return int(converted), nil
	case lua.LBool:
		return bool(converted), nil
	case *lua.LTable:
		tMap := make(map[lua.LValue]lua.LValue)
		var out []any
		for k, v := converted.Next(lua.LNil); k != lua.LNil; k, v = converted.Next(k) {
			strV, ok := v.(lua.LString)
			if !ok {
				return nil, &LuaError{err: fmt.Errorf("value at key %v is not a string - type: %T", k, v)}
			}

			out = append(out, strV)
		}

		return tMap, nil
	default:
		return nil, fmt.Errorf("unsupported lua value type: %T", lv)
	}
}

func typeWillBeLuaTable(v any) bool {
	rv := reflect.ValueOf(v)

	// repeat dereference until ptr value is reached
	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}
