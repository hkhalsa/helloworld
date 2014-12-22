package wrapper

// Wraps system interaction (graphics, input, sound) behind a shim.

import (
	"github.com/veandco/go-sdl2/sdl"
	"reflect"
	"unsafe"
)

// A window with an RGB buffer.
type GraphicsWindow struct {
	window *sdl.Window
	renderer *sdl.Renderer
	texture *sdl.Texture

	// Units for Height and Width are pixels.
	Height int
	Width int

	// The raw pixels we write to via SetPixel.
	lockedPixels []byte
}

// Set the (x,y)-th pixel to (r,g,b).
func (gw *GraphicsWindow) SetPixel(x, y int, r, g, b byte) {
	base := 4 * (y * gw.Width + x)

	// The format is asserted to be a 32-bit ARGB pixel.
	gw.lockedPixels[base] = b
	gw.lockedPixels[base + 1] = g
	gw.lockedPixels[base + 2] = r
	gw.lockedPixels[base + 3] = 0
}

// Used internally.  Pixels are written directly to video memory.  The memory must be 'locked'
// in order to write to it.
func (gw *GraphicsWindow) lockTexture() {
	var pixels unsafe.Pointer
	var pitch int

	err := gw.texture.Lock(nil, &pixels, &pitch)
	if nil != err {
		panic(err)
	}

	// Turn 'pixels' into a slice.
	length := 4 * gw.Height * gw.Width
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&gw.lockedPixels))
	sliceHeader.Cap = int(length)
	sliceHeader.Len = int(length)
	sliceHeader.Data = uintptr(pixels)
}

// Make the graphics buffer visible to the window.  Pixels written via SetPixel are not visible
// until Blit() is called.
func (gw *GraphicsWindow) Blit() {
	gw.texture.Unlock()
	gw.renderer.Copy(gw.texture, nil, nil)
	gw.renderer.Present()
	gw.lockTexture()
}

// Create a new window with the provided size and title.  Dies if any error is encountered
// or if the window isn't the format we expect.
func NewWindow(height int, width int, title string)(gw *GraphicsWindow) {
	gw = new(GraphicsWindow)
	gw.Width = width
	gw.Height = height

	var err error;
	gw.window, err = sdl.CreateWindow(title,
					  sdl.WINDOWPOS_UNDEFINED,
					  sdl.WINDOWPOS_UNDEFINED,
					  2 * width,
					  2 * height,
					  sdl.WINDOW_SHOWN)
	if nil != err {
		panic(err)
	}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "linear")
	// TODO: Passing in sdl.RENDERER_ACCELERATED as the flags creates a slow renderer.  Why?
	gw.renderer, err = sdl.CreateRenderer(gw.window, -1, 0)
	if nil != err {
		panic(err)
	}

	gw.texture, err = gw.renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888,
						    sdl.TEXTUREACCESS_STREAMING,
						    width,
						    height)
	if nil != err {
		panic(err)
	}

	// All writes to the texture must be done with the texture "locked"
	gw.lockTexture()
	return
}
