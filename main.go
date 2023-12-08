package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"log"
	"sync"
	"time"
)

type (
	Runner struct {
		ID string
		Stats
	}

	Stats struct {
		StartTime int64
	}
)

func main() {
	const numRunners = 100

	var wg sync.WaitGroup
	wg.Add(numRunners)

	for i := 0; i < numRunners; i++ {
		go func(id int) {
			defer wg.Done()
			lr, err := NewLuaRunner(false, LuaLib(lua.StringLibName), BasePrint)
			if err != nil {
				log.Fatal(err)
			}

			lfc := LuaFunctionCode(`
				function Run(input)
					print("Hello from: " .. input.ID .. ", Start time: " .. tostring(input.Stats.StartTime))
				end
			`)

			runner := &Runner{
				ID: fmt.Sprintf("Runner ID %d", id),
				Stats: Stats{
					StartTime: time.Now().UnixMilli(),
				},
			}

			_, err = lr.Run("Run", lfc, 1, runner)
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}

	wg.Wait()
}
