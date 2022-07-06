package boolbitmap

import (
	"fmt"
	"math"
	"sync"
)

func bool2BitmapRoutine(p []bool, img *[]byte, offset, inc int, wg *sync.WaitGroup) {
	var aux byte
	var rem int
	for i := offset * 8; i < len(p); i += 8 * inc {
		aux = 0b0
		rem = len(p) - i
		if rem > 8 {
			rem = 8
		}
		for j, v := range p[i : i+rem] {
			if v {
				aux |= 1 << j
			}
		}
		(*img)[i/8] = aux
	}
	wg.Done()
}

// MarshalParallel executes in n Goroutines and returns a bit array from the boolean array v.
func MarshalParallel(v interface{}, n int) ([]byte, error) {
	p, ok := v.([]bool)
	if !ok {
		return nil, fmt.Errorf("boolbit: Can't marshal type %T", v)
	}
	var wg sync.WaitGroup
	b := make([]byte, int(math.Ceil(float64(len(p))/8)))

	for i := 0; i < n; i++ {
		wg.Add(1)
		go bool2BitmapRoutine(p, &b, i, n, &wg)
	}
	wg.Wait()
	return b, nil
}

// Marshal returns a bit array from boolean array v.
func Marshal(v interface{}) ([]byte, error) {
	return MarshalParallel(v, 1)
}
