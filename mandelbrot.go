package mandelbrot

import (
	"errors"
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

// FractalGenerator is a interface that wraps a fractal generator.
type FractalGenerator interface {
	ComputeFractal(
		center complex128,
		zoom float64,
		resx,
		resy,
		iterations int,
		lim float64,
	) ([]bool, error)
}

// Fractal is a structure that implements FractalGenerator.
// The Goroutines is the number os goroutines that will be used to compute the fractal.
type Fractal struct {
	Goroutines int
}

// ComputeFractal returns a boolean slice B with size resx * resy containing the computed fractal.
// A point P(x, y) is part of the Mandelbrot Set if B[resx * y + x] is true.
func (g Fractal) ComputeFractal(center complex128, zoom float64, resx, resy, iterations int, lim float64) ([]bool, error) {
	var wg sync.WaitGroup
	m := mconf{
		resx:       resx,
		resy:       resy,
		limSq:      lim * lim,
		increment:  g.Goroutines,
		dist:       1 / zoom,
		iterations: iterations,
	}
	m.upperLeft = center + complex(-m.dist*float64(resx)/2+m.dist/2, m.dist*float64(resy)/2-m.dist/2)
	fractal := make([]bool, resx*resy)
	m.fractal = &fractal
	if g.Goroutines < 1 {
		return nil, errors.New("mandelbrot: Goroutines cannot be less than 1")
	}

	for i := 0; i < g.Goroutines; i++ {
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
			pixel, _ := m.evalCoord(complex(float64(j)*m.dist, float64(-k)*m.dist) + m.upperLeft)
			(*m.fractal)[k*m.resx+j] = pixel
		}
		carry = j - m.resx
		k++
	}
	wg.Done()
}

func (m *mconf) evalCoord(c complex128) (bool, int) {
	var i int
	z := complex128(0)
	for i = 0; i < m.iterations; i++ {
		z = z*z + c
		if real(z)*real(z)+imag(z)*imag(z) > m.limSq {
			return false, i
		}
	}
	return true, -1
}
