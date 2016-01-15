package main

import (
	"flag"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
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
	winWidth, winHeight int    = 1280, 720

	ySize = winWidth * winHeight
	cSize = ySize >> 2
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

	texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_IYUV, sdl.TEXTUREACCESS_STREAMING, winWidth, winHeight)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
		os.Exit(4)
	}
	defer texture.Destroy()

	tsReader := ts.NewPayloadUnitReader(file, ts.IsPID(pid))
	pesReader := pes.NewPayloadReader(tsReader)
	seq := video.NewVideoSequence(pesReader)
	seq.AlignTo(video.SequenceHeaderStartCode)

	var pointer unsafe.Pointer
	var pitch int

	for {
		img, imgErr := seq.Next()
		if imgErr != nil {
			break
		}

		texture.Lock(nil, &pointer, &pitch)
		pixels := (*[ySize + 2*cSize]uint8)(pointer)
		y := pixels[0:ySize]
		cb := pixels[ySize : ySize+cSize]
		cr := pixels[ySize+cSize:]
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
