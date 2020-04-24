package mandelbrotocl

// #cgo CFLAGS: -std=c11
// #cgo LDFLAGS: -lOpenCL
// #include "mandelbrotocl.h"
import "C"

// FractalOpenCL is a structure that implements FractalGenerator.
type FractalOpenCL struct {
	LocalSize int
}

// ComputeFractal returns a boolean slice B with size resx * resy containing the computed fractal.
// A point P(x, y) is part of the Mandelbrot Set if B[resx * y + x] is true.
func (g FractalOpenCL) ComputeFractal(center complex128, zoom float64, resx, resy, iterations int, lim float64) ([]bool, error) {

	cOut := make([]C.bool, resx*resy)
	C.compute_fractal(
		&cOut[0],
		C.double(real(center)),
		C.double(imag(center)),
		C.double(1/zoom),
		C.int(resx),
		C.int(resy),
		C.double(lim*lim),
		C.int(iterations),
		C.size_t(g.LocalSize),
	)
	out := make([]bool, resx*resy)
	for i, v := range cOut {
		out[i] = bool(v)
	}
	return out, nil
}
