package compute

// FractalParameters is the structure that holds FractalGenerator parameters
type FractalParameters struct {
	Center     complex128
	Zoom       float64
	ResX       int
	ResY       int
	Iterations int
	Lim        float64
}

// FractalGenerator is a interface that wraps a fractal generator.
type FractalGenerator interface {
	ComputeFractal(fp FractalParameters) ([]bool, error)
}
