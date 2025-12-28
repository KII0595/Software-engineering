package core

import (
	"context"
	"sync"
)

func NumberGenerator(ctx context.Context, start, count int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; i < count; i++ {
			select {
			case <-ctx.Done():
				return
			case out <- start + i:
			}
		}
	}()
	return out
}

func Transform(ctx context.Context, in <-chan int, fn func(int) int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case <-ctx.Done():
				return
			case out <- fn(v):
			}
		}
	}()
	return out
}

func Collect(ch <-chan int) []int {
	var result []int
	for v := range ch {
		result = append(result, v)
	}
	return result
}

func MergeChannels(ctx context.Context, ins ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(ins))

	for _, c := range ins {
		go func(ch <-chan int) {
			defer wg.Done()
			for v := range ch {
				select {
				case <-ctx.Done():
					return
				case out <- v:
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
