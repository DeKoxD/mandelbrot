package compute

// #cgo CFLAGS: -std=c11
// #cgo LDFLAGS: -lOpenCL
// #include "mandelbrotocl.h"
import "C"

// FractalOpenCL is a structure that implements FractalGenerator.
type FractalOpenCL struct {
	LocalSize int
}

// ComputeFractal returns a boolean slice B with size ResX * ResY containing the computed fractal.
// A point P(x, y) is part of the Mandelbrot Set if B[ResX * y + x] is true.
func (g FractalOpenCL) ComputeFractal(fp FractalParameters) ([]bool, error) {

	cOut := make([]C.bool, fp.ResX*fp.ResY)
	C.compute_fractal(
		&cOut[0],
		C.double(real(fp.Center)),
		C.double(imag(fp.Center)),
		C.double(1/fp.Zoom),
		C.int(fp.ResX),
		C.int(fp.ResY),
		C.double(fp.Lim*fp.Lim),
		C.int(fp.Iterations),
		C.size_t(g.LocalSize),
	)
	out := make([]bool, fp.ResX*fp.ResY)
	for i, v := range cOut {
		out[i] = bool(v)
	}
	return out, nil
}
