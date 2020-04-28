package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/DeKoxD/mandelbrot/mandelbrotocl"

	"github.com/DeKoxD/boolbitmap"
	"github.com/DeKoxD/mandelbrot"
)

type response struct {
	ResX, ResY int
	Image      []byte
}

func calcIterations(iterations int, zoom float64) int {
	return int(float64(iterations) * math.Log1p(zoom))
}

func calcZoom(zoom int) float64 {
	return math.Pow(1.25, float64(zoom))
}

func fractalHandler(it int, lim float64, goroutines int, fg mandelbrot.FractalGenerator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		centerx, centery, zoom, x, y := r.FormValue("centerx"), r.FormValue("centery"), r.FormValue("zoom"), r.FormValue("resx"), r.FormValue("resy")
		if len(centerx) == 0 || len(centery) == 0 || len(zoom) == 0 || len(x) == 0 || len(y) == 0 {
			return
		}
		ctx, err := strconv.ParseFloat(centerx, 64)
		if err != nil {
			return
		}
		cty, err := strconv.ParseFloat(centery, 64)
		if err != nil {
			return
		}
		zm, err := strconv.ParseFloat(zoom, 64)
		if err != nil {
			return
		}
		resx, err := strconv.Atoi(x)
		if err != nil {
			return
		}
		resy, err := strconv.Atoi(y)
		if err != nil {
			return
		}

		iterations := calcIterations(it, zm)

		set, err := fg.ComputeFractal(complex(ctx, cty), zm, resx, resy, iterations, lim)
		if err != nil {
			log.Println(err)
		}

		resp := response{
			ResX: resx,
			ResY: resy,
		}
		resp.Image, err = boolbitmap.MarshalParallel(set, goroutines)
		if err != nil {
			log.Println(err)
		}
		payload, err := json.Marshal(resp)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(payload))
	}
}

func main() {
	threads := flag.Int("t", 1, "Number of threads")
	limit := flag.Float64("l", 2, "Escaping limit")
	iterations := flag.Int("i", 100, "Max number of iterations")
	port := flag.String("p", "8080", "Port to serve on")
	address := flag.String("a", "", "Address to serve")
	opencl := flag.Bool("opencl", false, "Use GPU with OpenCL")
	flag.Parse()

	var fg mandelbrot.FractalGenerator
	if *opencl {
		fg = mandelbrotocl.FractalOpenCL{
			LocalSize: 32,
		}
	} else {
		fg = mandelbrot.Fractal{
			Goroutines: *threads,
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static/")))
	mux.HandleFunc("/fractal", fractalHandler(*iterations, *limit, *threads, fg))

	log.Printf("Serving on HTTP port: %s address: %s\nComputing on %v threads\n", *port, *address, *threads)
	log.Fatal(http.ListenAndServe(*address+":"+*port, mux))
}
