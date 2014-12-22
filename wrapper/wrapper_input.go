package wrapper

// Wraps system interaction (graphics, input, sound) behind a shim.

import (
	"github.com/veandco/go-sdl2/sdl"
)

// These are the possible user input events that the InputProvider will supply.
const (
	// Player 1 input.
	KEY_UP_1 = iota
	KEY_DOWN_1
	KEY_LEFT_1
	KEY_RIGHT_1
	KEY_SELECT_1
	KEY_START_1
	KEY_B_1
	KEY_A_1

	// "System" inputs.
	KEY_RESET
	KEY_QUIT
)

// Create a new InputProvider.  An InputProvider maps user key presses to buttons/events that occur
// on the NES (controller inputs, reset, power off).
func NewInputProvider() (out *InputProvider) {
	out = new(InputProvider)
	out.pressed = make(map[int]bool)

	if sdl.NumJoysticks() == 0 {
		out.joy = nil
	} else {
		out.joy = sdl.JoystickOpen(0)
		if nil == out.joy {
			panic("there's joy but i can't access it, life is suffering")
		}
	}

	return
}

// An InputProvider describes what keys are pressed.
type InputProvider struct {
	// Used to answer "is this logical NES button pressed" queries.
	// Maybe this is too obvious but pressed[KEY_UP] is true if the user has pressed 'up.
	pressed map[int]bool

	joy *sdl.Joystick
}

// This is where the human-accessible keyboard is mapped to the NES button presses.
var keyMapping = map[sdl.Keycode]int {
	// Game-accessible bindings.

	// Player 1.
	sdl.K_w: KEY_UP_1,
	sdl.K_s: KEY_DOWN_1,
	sdl.K_a: KEY_LEFT_1,
	sdl.K_d: KEY_RIGHT_1,
	sdl.K_y: KEY_SELECT_1,
	sdl.K_u: KEY_START_1,
	sdl.K_h: KEY_B_1,
	sdl.K_j: KEY_A_1,

	// Not-game-accessible bindings.
	sdl.K_r: KEY_RESET,
	sdl.K_q: KEY_QUIT,
}

// These are hardcoded button IDs for the PlayStation controller I use.
var gamepadMapping = map[uint8]int {
	4: KEY_UP_1,
	6: KEY_DOWN_1,
	7: KEY_LEFT_1,
	5: KEY_RIGHT_1,
	0: KEY_SELECT_1,
	3: KEY_START_1,
	14: KEY_B_1,
	13: KEY_A_1,
}

// Returns true if the provided key is pressed, false otherwise.
// It is expected that 'nesKey' is one of the KEY_xxx consts above.
func (ip *InputProvider) IsKeyPressed(nesKey int) bool {
	// Process all outstanding events.  TODO: Perhaps this should be done in a separate thread
	// which loops forever and updates state that's read by others?
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.KeyUpEvent:
			if key, ok := keyMapping[t.Keysym.Sym]; ok {
				ip.pressed[key] = false
			}
		case *sdl.KeyDownEvent:
			if key, ok := keyMapping[t.Keysym.Sym]; ok {
				ip.pressed[key] = true
			}
		case *sdl.JoyButtonEvent:
			if key, ok := gamepadMapping[t.Button]; ok {
				// Both down and up are put into the same event.
				ip.pressed[key] = (sdl.JOYBUTTONDOWN == t.Type)
			}
		}
	}

	// Return the information the user actually wants.
	return ip.pressed[nesKey]
}
