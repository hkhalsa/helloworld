helloworld
==========

A Nintendo Entertainment System emulator written in Go as a "getting-to-know-Go" project.

# Issues that will be resolved at some point soon

* Sound is currently lacking, as is MMC3 support.  But there is joypad support.

* Uses the Go + SDL bindings available at https://github.com/veandco/go-sdl2 ...but! there is a
  small change required to the current implementation of sdl.Texture.Lock that I haven't submitted a
  pull request for yet.

* The package organization is probably not compatible with "go get" but I'll experiment with fixing
  that later when I can actually "go get" the code.

* The frame-rate limiting is done through sleep calls.  So, I hope you're running the same CPU I am.
