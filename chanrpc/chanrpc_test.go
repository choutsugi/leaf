package chanrpc

import (
	"fmt"
	"sync"
)

func Example() {
	s := NewServer(10)

	var wg sync.WaitGroup

	wg.Add(1)
	// goroutine 1
	go func() {
		// register function
		s.Register("f0", func(args []interface{}) {

		})

		s.Register("f1", func(args []interface{}) interface{} {
			return 1
		})

		s.Register("fn", func(args []interface{}) []interface{} {
			return []interface{}{1, 2, 3}
		})

		s.Register("add", func(args []interface{}) interface{} {
			n1 := args[0].(int)
			n2 := args[1].(int)
			return n1 + n2
		})

		// register has completed
		wg.Done()

		// loop for listen and execute function
		for {
			err := s.Exec(<-s.ChanCall)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	// wait for register be completed
	wg.Wait()

	wg.Add(1)
	// goroutine 2
	go func() {
		c := s.Open(10)

		// sync call
		err := c.Call0("f0")
		if err != nil {
			fmt.Println(err)
		}

		r1, err := c.Call1("f1")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(r1)
		}

		rn, err := c.CallN("fn")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(rn[0], rn[1], rn[2])
		}

		ra, err := c.Call1("add", 1, 2)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(ra)
		}

		// async call
		c.AsyncCall("f0", func(err error) {
			if err != nil {
				fmt.Println(err)
			}
		})

		c.AsyncCall("f1", func(ret interface{}, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(ret)
			}
		})

		c.AsyncCall("fn", func(ret []interface{}, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(ret[0], ret[1], ret[2])
			}
		})

		c.AsyncCall("add", 1, 2, func(ret interface{}, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(ret)
			}
		})

		c.Cb(<-c.ChanAsyncRet)
		c.Cb(<-c.ChanAsyncRet)
		c.Cb(<-c.ChanAsyncRet)
		c.Cb(<-c.ChanAsyncRet)

		// go
		s.Go("f0")

		wg.Done()
	}()

	wg.Wait()

	// Output:
	// 1
	// 1 2 3
	// 3
	// 1
	// 1 2 3
	// 3
}
