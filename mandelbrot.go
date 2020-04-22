package mandelbrot

import (
	"math"
	"sync"
)

type mconf struct {
	fractal    *[]bool
	iterations int
	limSq      float64
	resx, resy int
	dist       float64
	upperLeft  complex128
	increment  int
}

// ComputeFractal returns a boolean slice with size resx * resy representing, for each point, if it's part of the Mandelbrot Set.
func ComputeFractal(center complex128, zoom float64, resx, resy, goroutines, iterations int, lim float64) ([]bool, error) {
	var wg sync.WaitGroup
	m := mconf{
		resx:       resx,
		resy:       resy,
		limSq:      lim * lim,
		increment:  goroutines,
		dist:       1 / zoom,
		iterations: 1 + int(float64(iterations)*math.Log1p(zoom)),
	}
	m.upperLeft = center + complex(-m.dist*float64(resx)/2+m.dist/2, m.dist*float64(resy)/2-m.dist/2)
	fractal := make([]bool, resx*resy)
	m.fractal = &fractal

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go m.boolArrayRoutine(i, &wg)
	}
	wg.Wait()
	return *m.fractal, nil
}

func (m mconf) boolArrayRoutine(offset int, wg *sync.WaitGroup) {
	var j, k int
	carry := offset
	for {
		if k >= m.resy {
			break
		}
		for j = carry; j < m.resx; j += m.increment {
			pixel := m.evalCoord(complex(float64(j)*m.dist, float64(-k)*m.dist) + m.upperLeft)
			(*m.fractal)[k*m.resx+j] = pixel
		}
		carry = j - m.resx
		k++
	}
	wg.Done()
}

func (m *mconf) evalCoord(c complex128) bool {
	var i int
	z := complex128(0)
	for i = 0; i < m.iterations; i++ {
		z = z*z + c
		if real(z)*real(z)+imag(z)*imag(z) > m.limSq {
			return false
		}
	}
	return true
}
