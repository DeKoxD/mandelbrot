package compute

import (
	"errors"
	"sync"
)

type mconf struct {
	fractal    *[]bool
	iterations int
	limSq      float64
	resX, resY int
	dist       float64
	upperLeft  complex128
	increment  int
}

// Fractal is a structure that implements FractalGenerator.
// The Goroutines is the number os goroutines that will be used to compute the fractal.
type Fractal struct {
	Goroutines int
}

// ComputeFractal returns a boolean slice B with size resX * resY containing the computed fractal.
// A point P(x, y) is part of the Mandelbrot Set if B[resX * y + x] is true.
func (g Fractal) ComputeFractal(fp FractalParameters) ([]bool, error) {
	var wg sync.WaitGroup
	fractal := make([]bool, fp.ResX*fp.ResY)
	dist := 1 / fp.Zoom
	m := mconf{
		resX:       fp.ResX,
		resY:       fp.ResY,
		limSq:      fp.Lim * fp.Lim,
		increment:  g.Goroutines,
		dist:       dist,
		iterations: fp.Iterations,
		upperLeft:  fp.Center + complex(-dist*float64(fp.ResX)/2+dist/2, dist*float64(fp.ResY)/2-dist/2),
		fractal:    &fractal,
	}
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
		if k >= m.resY {
			break
		}
		for j = carry; j < m.resX; j += m.increment {
			pixel, _ := m.evalCoord(complex(float64(j)*m.dist, float64(-k)*m.dist) + m.upperLeft)
			(*m.fractal)[k*m.resX+j] = pixel
		}
		carry = j - m.resX
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
