package compute

import (
	"errors"

	"github.com/DeKoxD/mandelbrot/pkg/queue"
)

type queueData struct {
	queue queue.Queue
	fg    FractalGenerator
}

// ComputeFractal function for Queue
func (q *queueData) ComputeFractal(fp FractalParameters) ([]bool, error) {
	err := q.queue.Begin()
	defer q.queue.End()
	if err != nil {
		return nil, err
	}

	return q.fg.ComputeFractal(fp)
}

// Queue returns a limited size FractalGenerator queue.
//
// ComputeFractal calls will run one by one.
//
// When queue is full, new ComputeFractal calls will return error.
func Queue(fg FractalGenerator, size uint) (FractalGenerator, error) {
	if fg == nil {
		return nil, errors.New("Invalid Fractal Generator")
	}
	q := &queueData{
		queue: queue.NewQueue(size),
		fg:    fg,
	}

	return q, nil
}

type raceData struct {
	fgs []FractalGenerator
}

// ComputeFractal function for Race
func (r *raceData) ComputeFractal(fp FractalParameters) ([]bool, error) {
	ch := make(chan []bool)
	for _, fg := range r.fgs {
		go func(fg *FractalGenerator, ch *chan []bool) {
			f, err := (*fg).ComputeFractal(fp)
			if f != nil && err == nil {
				select {
				case *ch <- f:
					return
				default:
				}
			}
		}(&fg, &ch)
	}

	select {
	case res := <-ch:
		return res, nil

	}
}

// Race returns a limited size FractalGenerator race.
//
// ComputeFractal will run in parallel and the first result is returned.
func Race(fgs ...FractalGenerator) (FractalGenerator, error) {
	r := &raceData{
		fgs: fgs,
	}
	return r, nil
}
