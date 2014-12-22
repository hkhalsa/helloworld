package wrapper

import (
	"github.com/veandco/go-sdl2/sdl"
	"runtime"
)

func Init() {
	// SDL must do some kind of thread-local business because it'll crash if I don't do this.
	// TODO: Debug this further.
	runtime.LockOSThread()
	sdl.Init(sdl.INIT_EVERYTHING)
}
