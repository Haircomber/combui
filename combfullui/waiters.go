// this file implements the dual waiters, a two jobs that alternate
package main

import "sync"

// like golang wait group, except allows to wait for first of two threads safely
type WaitGroup chan struct{}

func (wg *WaitGroup) Add(int) {
}

func (wg *WaitGroup) Done() {
	*wg <- struct{}{}
}

func (wg *WaitGroup) Wait() {
	<-*wg
}

func dualWaitersRoutinesCount(j int) int {
	if j == 0 {
		return 2
	}
	return 1
}

func dualWaitersMsWaitTime(i, j, k int) int {
	if j > k {
		j = k
	}
	return 1000 * i << j
}

func dualWaiters(waiter func(int, int) error, k int) (err error) {
	var wg = make(WaitGroup)
	var mut sync.Mutex
	for j := 0; true; j++ {
		wg.Add(1)
		for i := 1; i <= dualWaitersRoutinesCount(j); i++ {
			wt := dualWaitersMsWaitTime(i, j, k)
			go func() {
				var err2 = waiter(wt, j)
				if err2 != nil {
					mut.Lock()
					if err == nil {
						err = err2
					}
					mut.Unlock()
				}
				wg.Done()
			}()
		}
		wg.Wait()
		mut.Lock()
		if err != nil {
			mut.Unlock()
			return err
		}
		mut.Unlock()
	}
	return nil
}

/*
	EXPECTED OUTPUT:
	one 2000 0
	one 1000 0
	two 1000 0
	one 2000 1
	two 2000 0
	one 4000 2
	two 2000 1
	one 8000 3
	two 4000 2
	one 16000 4
	two 8000 3
	one 32000 5
	two 16000 4
	one 64000 6
	two 32000 5
	one 128000 7
	...

	func dualWaitersTest() {
		println(dualWaiters(func(i, j int) error {
			println("one", i, j)
			time.Sleep(time.Duration(i) * time.Millisecond)
			println("two", i, j)
			return nil
		}, 16))
	}
*/