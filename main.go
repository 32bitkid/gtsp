package main

import (
	"flag"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"os"
	"runtime"
	"unsafe"
)

import (
	"github.com/32bitkid/mpeg/pes"
	"github.com/32bitkid/mpeg/ts"
	"github.com/32bitkid/mpeg/video"
)

const (
	winTitle            string = "Go-SDL2 MPEG-2 Player"
	winWidth, winHeight int    = 1920 >> 1, 1080 >> 1

	maxFrameSize = 16384 * 16384 * 3
)

var pid = flag.Int("pid", 0x21, "the PID to play")

func init() {
	runtime.LockOSThread()
}

func play(file *os.File, pid uint32) {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var texture *sdl.Texture
	var err error

	window, err = sdl.CreateWindow(winTitle,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight,
		sdl.WINDOW_SHOWN)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		os.Exit(1)
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(2)
	}
	defer renderer.Destroy()

	tsReader := ts.NewPayloadUnitReader(file, ts.IsPID(pid))
	pesReader := pes.NewPayloadReader(tsReader)
	seq := video.NewVideoSequence(pesReader)
	seq.AlignTo(video.SequenceHeaderStartCode)

	var pointer unsafe.Pointer
	var pitch int
	var ySize int
	var cSize int

	for {
		img, err := seq.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Decoding error: %s\n", err)
			os.Exit(1)
		}

		if texture == nil {
			w, h := seq.Size()
			fmt.Println(w, h)
			ySize = w * h
			cSize = (w * h) >> 2
			texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_IYUV, sdl.TEXTUREACCESS_STREAMING, w, h)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
				os.Exit(4)
			}
			defer texture.Destroy()
		}

		texture.Lock(nil, &pointer, &pitch)
		pixels := (*[maxFrameSize]uint8)(pointer)
		y := pixels[0:ySize]
		cb := pixels[ySize : ySize+cSize]
		cr := pixels[ySize+cSize : ySize+cSize+cSize]
		copy(y, img.Y)
		copy(cb, img.Cb)
		copy(cr, img.Cr)
		texture.Unlock()

		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}
}

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}

	play(file, uint32(*pid))
}
