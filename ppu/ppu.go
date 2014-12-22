package ppu

import (
	"mapper"
	"wrapper"
)

const (
	// The display is 240 pixels high but the top and bottom 8 are usually cut off by the
	// monitor.  They're currently displayed but often contain garbage since the NES programmer
	// generally does not expect them to be displayed on an NTSC TV.
	DisplayHeight = 240

	DisplayWidth = 256
)

type PPU struct {
	// Set to true to enable some debugging logging.
	Debug bool

	// PPU Address space.  0x0000 -> 0x3FFF is mirrored repeatedly at 0x4000 and above.
	// [0x0000 -> 0x3EFF] is dealt with by the cart.
	cartMapper mapper.Mapper

	// 0x3F00 -> 0x3F1F, mirrored through 0x3FFF.
	// 0x3F20 -> 0x3FFF mirrors this.
	pal [0x20]byte

	// Sprite data.
	oamData [256]byte

	// PPU Registers mapped to the CPU's address space.  We store the writes verbatim rather
	// than parse them into a series of bools.

	// 0x2000 -- PPUCTRL -- write only
	//
	// 7654 3210
	// |||| ||||
	// |||| ||++- Base nametable address
	// |||| ||    (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
	// |||| |+--- VRAM address increment per CPU read/write of PPUDATA
	// |||| |     (0: add 1, going across; 1: add 32, going down)
	// |||| +---- Sprite pattern table address for 8x8 sprites
	// ||||       (0: $0000; 1: $1000; ignored in 8x16 mode)
	// |||+------ Background pattern table address (0: $0000; 1: $1000)
	// ||+------- Sprite size (0: 8x8; 1: 8x16)
	// |+-------- PPU master/slave select
	// |          (0: read backdrop from EXT pins; 1: output color on EXT pins)
	// +--------- Generate an NMI at the start of the
	//            vertical blanking interval (0: off; 1: on)
	ppuCtrl byte

	// 0x2001 -- PPUMASK -- write only
	//
	// 76543210
	// ||||||||
	// |||||||+- Grayscale (0: normal color; 1: produce a monochrome display)
	// ||||||+-- 1: Show background in leftmost 8 pixels of screen; 0: Hide
	// |||||+--- 1: Show sprites in leftmost 8 pixels of screen; 0: Hide
	// ||||+---- 1: Show background
	// |||+----- 1: Show sprites
	// ||+------ Intensify reds (and darken other colors)
	// |+------- Intensify greens (and darken other colors)
	// +-------- Intensify blues (and darken other colors)
	ppuMask byte

	// 0x2002 -- PPUSTATUS -- read only
	//
	// 7654 3210
	// |||| ||||
	// |||+-++++- Least significant bits previously written into a PPU register
	// |||        (due to register not being updated for this address)
	// ||+------- Sprite overflow. The intent was for this flag to be set
	// ||         whenever more than eight sprites appear on a scanline, but a
	// ||         hardware bug causes the actual behavior to be more complicated
	// ||         and generate false positives as well as false negatives; see
	// ||         PPU sprite evaluation. This flag is set during sprite
	// ||         evaluation and cleared at dot 1 (the second dot) of the
	// ||         pre-render line.
	// |+-------- Sprite 0 Hit.  Set when a nonzero pixel of sprite 0 overlaps
	// |          a nonzero background pixel; cleared at dot 1 of the pre-render
	// |          line.  Used for raster timing.
	// +--------- Vertical blank has started (0: not in VBLANK; 1: in VBLANK).
	// 	      Set at dot 1 of line 241 (the line *after* the post-render
	// 	      line); cleared after reading $2002 and at dot 1 of the
	// 	      pre-render line.
	ppuStatus byte

	// 0x2003 -- write only.  The address of the sprite RAM to r/w when 2004 is accessed.
	// 0x2004 -- read/write.  Where the sprite IO occurs.  oamAddr is incremented after
	// this is rw'd.
	oamAddr byte

	// 0x2005 -- PPU scrolling.  How many pixels from the "top left" do we begin to display?
	// These are complicated, see:
	// http://wiki.nesdev.com/w/index.php/The_skinny_on_NES_scrolling
	loopyT uint16
	loopyV uint16
	loopyX byte
	addressLatch byte

	// When data is read through 0x2007, if we're not reading palette data,
	// THEN it's buffered and returned on a subsequent read.
	bufferedReadData uint8

	// The graphics window we render into.
	window *wrapper.GraphicsWindow

	// What scan line are we rendering?
	scanLineCounter uint16
}

func NewPPU(cartMapper mapper.Mapper, window *wrapper.GraphicsWindow) (ppu *PPU) {
	ppu = new(PPU)
	ppu.window = window
	ppu.cartMapper = cartMapper
	ppu.addressLatch = 0
	ppu.scanLineCounter = 0
	return
}

// Enter the VBlank period.  Returns true if the CPU should handle a NMI, false otherwise.
func (ppu *PPU) EnterVBlankShouldNMI() bool {
	// Turn on the "we're in VBlank" flag.
	ppu.ppuStatus |= 0x80

	// If this bit is set, the CPU wants an NMI to occur when VBlank happens.
	return 0x80 == (ppu.ppuCtrl & 0x80)
}

// Exit the VBlank period.
func (ppu *PPU) ExitVBlank() {
	// Turn off the sprite zero and VBlank flags.
	ppu.ppuStatus &= 0x3f

	// Start rendering line 0.
	ppu.scanLineCounter = 0
}
