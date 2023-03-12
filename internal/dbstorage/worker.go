// Package dbstorage used for persisting urls in the database
package dbstorage

import (
	"strconv"
	"sync"
)

func newWorker(input chan string, out chan int64) {
	go func() {
		for val := range input {
			id, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				out <- id
			}
		}
		close(out)
	}()
}

func fanOut(inputCh chan string, n int) []chan string {
	chs := make([]chan string, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan string) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			val, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- val
		}
	}()

	return chs
}

func fanIn(inputChs ...chan int64) chan int64 {
	outCh := make(chan int64)
	go func() {
		wg := &sync.WaitGroup{}

		for _, inputCh := range inputChs {
			wg.Add(1)

			go func(inputCh chan int64) {
				defer wg.Done()
				for item := range inputCh {
					outCh <- item
				}
			}(inputCh)
		}
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
