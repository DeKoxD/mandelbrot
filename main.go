package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
)

type response struct {
	ResX, ResY int
	Image      []byte
}

func Bool2BitmapRoutine(p []bool, img *[]byte, offset, inc int, wg *sync.WaitGroup) {
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

func (r *response) Bool2Bitmap(p []bool, threads int) {
	var wg sync.WaitGroup
	if cap(r.Image) < len(p) {
		r.Image = make([]byte, len(p)/8+1)
	}
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go Bool2BitmapRoutine(p, &r.Image, i, threads, &wg)
	}
	wg.Wait()
}

type ByteArrayMngr interface {
	IndexWrite(idx int, v interface{}) bool
	Resize(size int)
}

type boolColor struct {
	Vmap []bool
}

func (b *boolColor) IndexWrite(idx int, v interface{}) (ok bool) {
	b.Vmap[idx], ok = v.(bool)
	return
}

func (b *boolColor) Resize(size int) {
	if cap(b.Vmap) < size {
		b.Vmap = make([]bool, size)
	} else {
		b.Vmap = b.Vmap[:size]
	}
}

type mandelbrot interface {
	EvalCoord(complex128) (int, bool)
	BoolMap(bmap *[]bool, center complex128, x, y, threads int) error
}

type mandelbrotSet struct {
	iterations int
	limSq      float64
	iw         ByteArrayMngr
	x, y       int
	dist       float64
	ul         complex128
	increment  int
}

func (m *mandelbrotSet) EvalCoord(c complex128) (int, bool) {
	var i int
	z := complex128(0)
	for i = 0; i < m.iterations; i++ {
		z = z*z + c
		if real(z)*real(z)+imag(z)*imag(z) > m.limSq {
			return i, false
		}
	}
	return i, true
}

func BoolMapRoutine(m *mandelbrotSet, offset int, wg *sync.WaitGroup) {
	var j, k int
	carry := offset
	for {
		if k >= m.y {
			break
		}
		for j = carry; j < m.x; j += m.increment {
			_, it := m.EvalCoord(complex(float64(j)*m.dist, float64(-k)*m.dist) + m.ul)
			m.iw.IndexWrite(k*m.x+j, it)
		}
		carry = j - m.x
		k++
	}
	wg.Done()
}

func (m *mandelbrotSet) BoolMap(center complex128, zoom float64, x, y, threads, it int, lim float64) error {
	var wg sync.WaitGroup
	m.x, m.y, m.increment, m.limSq = x, y, threads, lim*lim
	m.iterations = 1 + int(float64(it)*math.Log1p(zoom))
	m.dist = 1 / zoom
	m.ul = center + complex(-m.dist*float64(x)/2+m.dist/2, m.dist*float64(y)/2-m.dist/2)
	m.iw.Resize(x * y)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go BoolMapRoutine(m, i, &wg)
	}
	wg.Wait()
	return nil
}

func fractalHandler(it int, lim float64, threads int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		centerx, centery, zoom, resx, resy := r.FormValue("centerx"), r.FormValue("centery"), r.FormValue("zoom"), r.FormValue("resx"), r.FormValue("resy")
		if len(centerx) == 0 || len(centery) == 0 || len(zoom) == 0 || len(resx) == 0 || len(resy) == 0 {
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
		x, err := strconv.Atoi(resx)
		if err != nil {
			return
		}
		y, err := strconv.Atoi(resy)
		if err != nil {
			return
		}
		b := &boolColor{}
		mset := mandelbrotSet{iterations: it, limSq: lim * lim, iw: b}
		mset.BoolMap(complex(ctx, cty), zm, x, y, threads, it, lim)
		payload := response{ResX: x, ResY: y}
		payload.Bool2Bitmap(b.Vmap, threads)
		js, err := json.Marshal(payload)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(js))
	}
}

func main() {
	threads := flag.Int("t", 1, "Number of threads")
	limit := flag.Float64("l", 2, "Escaping limit")
	iterations := flag.Int("i", 100, "Max number of iterations")
	port := flag.String("p", "8080", "Port to serve on")
	address := flag.String("a", "", "Address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static/")))
	mux.HandleFunc("/fractal", fractalHandler(*iterations, *limit, *threads))

	log.Printf("Serving on HTTP port: %s address: %s\nComputing on %v threads\n", *port, *address, *threads)
	log.Fatal(http.ListenAndServe(*address+":"+*port, mux))
}
