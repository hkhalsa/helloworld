package main

import (
	"mapper"
	"ppu"
	"wrapper"
)

// The NES-specific implementation of the CPU memory interface.  Some addresses are handled by the
// PPU, some by the cart mapper, some are on-board RAM, and some read from external devices.
type NESMemory struct {
	// [0x0000 -> 0x07FF] is actual RAM.
	// [0x0800 -> 0x1FFF] mirrors it.
	ram [0x800]byte

	// [0x2000 -> 0x2007] are memory mapped I/O registers in the PPU, mirrored throughout
	// [0x2008 -> 0x3FFF].
	// We just defer all this IO to the PPU struct.
	ppu *ppu.PPU

	// [0x4000 -> 0x4017] are audio or controller mmio registers.

	// For 0x4016:
	// Which of keyReadOrder are we currently returning?
	currentKeyRead int
	// Provides "is this key pressed" functionality.
	input *wrapper.InputProvider

	// [0x4018 -> 0xFFFF] is mapped by the cart.
	cartMapper mapper.Mapper
}

// Key presses are returned via the controller mmio regs in the following order.
var keyReadOrder = [8]int {
	wrapper.KEY_A_1,
	wrapper.KEY_B_1,
	wrapper.KEY_SELECT_1,
	wrapper.KEY_START_1,
	wrapper.KEY_UP_1,
	wrapper.KEY_DOWN_1,
	wrapper.KEY_LEFT_1,
	wrapper.KEY_RIGHT_1,
}

// The CPU is reading from 'addr'.  Dispatch to the correct handler.
func (mem *NESMemory) Read(addr uint16) uint8 {
	if addr < 0x2000 {
		// [0x0000 -> 0x1FFF], but mirrored.
		addr &= 0x7ff
		return mem.ram[addr]
	} else if addr < 0x4000 {
		// [0x2000 -> 0x3FFF] is memory-mapped PPU space.
		return mem.ppu.ReadRegister(addr)
	} else if addr < 0x4018 {
		// [0x4000 -> 0x4017] is audio/input device registers.  And sprite DMA but that is
		// write-only.

		// 0x4016 is the 1st controller mmio register and the only one currently
		// implemented.
		if 0x4016 != addr {
			return 0
		}

		if mem.currentKeyRead >= len(keyReadOrder) {
			return 1
		}
		// Keys are returned in the 1st bit of the read.
		// The order is: A B Select Start Up Down Left Right
		pressed := mem.input.IsKeyPressed(keyReadOrder[mem.currentKeyRead])
		mem.currentKeyRead = mem.currentKeyRead + 1
		if pressed {
			return 1
		} else {
			return 0
		}
	} else {
		return mem.cartMapper.ReadCPU(addr)
	}

	panic("This shouldn't be reached")
	return 0
}

// The CPU is writing 'val' to 'addr'.  Dispatch to the correct handler.  Returns how many extra
// cycles the write takes.
func (mem *NESMemory) Write(addr uint16, val uint8) (cycles uint64) {
	if addr < 0x2000 {
		// [0x0000 -> 0x1FFF]
		addr &= 0x7ff
		mem.ram[addr] = val
	} else if addr < 0x4000 {
		// [0x2000 -> 0x3FFF]
		mem.ppu.WriteRegister(addr, val)
	} else if addr < 0x4018 {
		if addr >= 0x4000 && addr <= 0x4013 {
			// These are audio registers.  Ignored until we have sound.
		} else if addr == 0x4014 {
			// Sprite DMA.  Transfer 256 bytes of memory to SPR-RAM from
			// 0x100 * val.
			start := 0x100 * int(val)
			end := start + 256
			mem.ppu.SpriteDMA(mem.ram[start:end])
			return 513
		} else if addr == 0x4015 {
			// This is also an ignored audio register.
		} else if addr == 0x4016 {
			// 0x4016 is a write register that is strobed to reset the game pad(s).
			// When 1 is written it continually reads the state of the game pad(s).
			// When 0 is written it stops doing so and makes the state available.
			if 0 == (val & 1) {
				mem.currentKeyRead = 0
			}
		}
		// 0x4017 is a read-only register.
	} else {
		// [0x4018 -> 0xFFFF]
		return mem.cartMapper.WriteCPU(addr, val)
	}

	// Most memory accesses don't take extra CPU cycles, but sprite DMA does.
	return 0
}

func NewNESMemory(ppu *ppu.PPU, cartMapper mapper.Mapper, input *wrapper.InputProvider) (nesMem *NESMemory) {
	nesMem = new(NESMemory)
	nesMem.ppu = ppu
	nesMem.cartMapper = cartMapper
	nesMem.currentKeyRead = 0
	nesMem.input = input
	return
}
