package main

import (
	"flag"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"os"
	"reflect"
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
	var yLen int
	var cLen int

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
			yLen = w * h
			cLen = (w * h) >> 2
			texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_IYUV, sdl.TEXTUREACCESS_STREAMING, w, h)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
				os.Exit(4)
			}
			defer texture.Destroy()
		}

		{
			texture.Lock(nil, &pointer, &pitch)

			// Convert pointer to []uint8
			pixels := *(*[]uint8)(unsafe.Pointer(&reflect.SliceHeader{
				Data: uintptr(pointer),
				Len:  yLen + 2*cLen,
				Cap:  yLen + 2*cLen,
			}))

			// Select color planes
			y := pixels[0:yLen]
			cb := pixels[yLen : yLen+cLen]
			cr := pixels[yLen+cLen : yLen+cLen+cLen]

			// Copy image data into texture
			copy(y, img.Y)
			copy(cb, img.Cb)
			copy(cr, img.Cr)

			texture.Unlock()
		}

		err = renderer.Copy(texture, nil, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Copy failed: %s\n", err)
			os.Exit(5)
		}
		renderer.Present()
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}

	play(file, uint32(*pid))
}
