# Go Transport Stream Player (gtsp)

An experimental simple MPEG-2 Transport Stream demuxer/player written in Go. Presentation is handled by [go-sdl2](https://github.com/veandco/go-sdl2), but the *actual* video decoding is done using an experimental pure Go video decoder, [mpeg](https://github.com/32bitkid/mpeg), hence it's slow is as molasses.

The primary goal of [mpeg](https://github.com/32bitkid/mpeg) is to be a human readable alternative to higher performance video decoding libraries, however work is still being done to improve both readability *and* performance where possible.

## Dependencies

`gtsp` requires SDL:

On **Ubuntu 14.04 and above**, use:
apt-get install libsdl2{,-mixer,-image,-ttf}-dev

On **Fedora 20 and above**, use:
yum install SDL2{,_mixer,_image,_ttf}-devel

On **Arch Linux**, use:
pacman -S sdl2{,_mixer,_image,_ttf}

On **Mac OS X**, install SDL2 via Homebrew like so: brew install sdl2{,_image,_ttf,_mixer}

On **Windows**, install SDL2 via Msys2 like so: pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-SDL2{,_mixer,_image,_ttf}

See [go-sdl2](https://github.com/veandco/go-sdl2) for more details.

## Installation

```bash
$ go get github.com/32bitkid/gtsp
```

## Notes

The MPEG-2 decoder is presently experimental and does not support the entire MPEG-2 specification; presently, it only supports a subset of Main Profile encoded videos (I,P,B pictures, 4:2:0 chroma, frame based pictures with frame based motion compensation). Underlying library support for more of the specification is on-going.

At the time of writing this, you should be able to play [PID 0x31 from this clip](http://files.32bitkid.com/video/elephants_dream_clip.ts) from the open source movie [Elephants Dream](https://orange.blender.org/):

```bash
$ gtsp -pid=0x31 elephants_dream_clip.ts
```
