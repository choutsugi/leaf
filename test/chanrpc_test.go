package test

import (
	"fmt"
	"github.com/choutsugi/forest/chanrpc"
	"sync"
	"testing"
)

func TestChanrpc(t *testing.T) {
	s := chanrpc.NewServer(10)

	var wg sync.WaitGroup

	wg.Add(1)
	// goroutine 1
	go func() {
		// register function
		s.Register("f0", func(args []interface{}) {
			fmt.Printf("CbFunc0 has called, args = %+v, ret = null\n", args)
		})

		s.Register("f1", func(args []interface{}) interface{} {
			fmt.Printf("CbFunc1 has called, args = %+v, ret = 1\n", args)
			return 1
		})

		s.Register("fn", func(args []interface{}) []interface{} {
			fmt.Printf("CbFuncN has called, args = %+v, ret = %+v\n", args, []interface{}{1, 2, 3})
			return []interface{}{1, 2, 3}
		})

		s.Register("add", func(args []interface{}) interface{} {
			n1 := args[0].(int)
			n2 := args[1].(int)
			fmt.Printf("CbFuncAdd has called, args = %+v, ret = %+v\n", args, n1+n2)
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
		} else {
			fmt.Printf("Call0 executed, args = f0, ret = null\n")
		}

		r1, err := c.Call1("f1")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Call1 executed, args = f1, ret = %+v\n", r1)
		}

		rn, err := c.CallN("fn")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("CallN executed, args = fn, ret = %+v\n", rn)
		}

		ra, err := c.Call1("add", 1, 2)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Call1 executed, args = [add,1,2], ret = %+v\n", ra)
		}

		// async call
		c.AsyncCall("f0", func(err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("AsyncCall executed, args = f0, ret = null\n")
			}
		})

		c.AsyncCall("f1", func(ret interface{}, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("AsyncCall executed, args = f1, ret = %+v\n", ret)
			}
		})

		c.AsyncCall("fn", func(ret []interface{}, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("AsyncCall executed, args = fn, ret = %+v\n", ret)
			}
		})

		c.AsyncCall("add", 1, 2, func(ret interface{}, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("AsyncCall executed, args = add, ret = %+v\n", ret)
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
	//	CbFunc0 has called, args = [], ret = null
	//	Call0 executed, args = f0, ret = null
	//	CbFunc1 has called, args = [], ret = 1
	//	Call1 executed, args = f1, ret = 1
	//	CbFuncN has called, args = [], ret = [1 2 3]
	//	CallN executed, args = fn, ret = [1 2 3]
	//	CbFuncAdd has called, args = [1 2], ret = 3
	//	Call1 executed, args = [add,1,2], ret = 3
	//	CbFunc0 has called, args = [], ret = null
	//	CbFunc1 has called, args = [], ret = 1
	//	CbFuncN has called, args = [], ret = [1 2 3]
	//	CbFuncAdd has called, args = [1 2], ret = 3
	//	AsyncCall executed, args = f0, ret = null
	//	AsyncCall executed, args = f1, ret = 1
	//	AsyncCall executed, args = fn, ret = [1 2 3]
	//	AsyncCall executed, args = add, ret = 3
	//	CbFunc0 has called, args = [], ret = null
}
