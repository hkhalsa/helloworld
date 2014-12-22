package ppu

// Is background rendering enabled?
func (ppu *PPU) shouldRenderBackground() bool {
	return 8 == (ppu.ppuMask & 8)
}

// Is sprite rendering enabled?
func (ppu *PPU) shouldRenderSprites() bool {
	return 0x10 == (ppu.ppuMask & 0x10)
}

// Does the palette entry 'palIndex' correspond to the background color?
func isBG(palIndex byte) bool {
	return 0 == (palIndex & 3)
}

// Given the two bytes that make up the same row in a tile, get the 'x'-th pixel where x=[0, 8)
func getPixel(tileBit0, tileBit1 byte, x int) (out byte) {
	out = (tileBit0 >> (7 - byte(x))) & 1
	if 7 == x {
		out |= 2 & (tileBit1 << 1)
	} else {
		out |= 2 & (tileBit1 >> (6 - byte(x)))
	}
	return
}

// Render one scan line.
func (ppu *PPU) RenderScanLine() {
	// There is no rendering enabled, just show the bg color.
	if !ppu.shouldRenderBackground() && !ppu.shouldRenderSprites() {
		bg := &Palette[ppu.pal[0]]

		for i := 0; i < DisplayWidth; i++ {
			ppu.window.SetPixel(i, int(ppu.scanLineCounter), bg.r, bg.g, bg.b)
		}
		ppu.scanLineCounter++
		return
	}

	if 0 == ppu.scanLineCounter {
		// At the start of rendering, loopyV is loaded from loopyT.
		ppu.loopyV = ppu.loopyT
	} else {
		// At each scanline start, the PPU copies all bits relating to horizontal position
		// from t to v.  Recall that loopyV looks like:
		//
		// 0yyy NNYY YYYX XXXX
		// aka
		// yyy NN YYYYY XXXXX
		// ||| || ||||| +++++-- coarse X scroll
		// ||| || +++++-------- coarse Y scroll
		// ||| ++-------------- nametable select
		// +++----------------- fine Y scroll
		ppu.loopyV &= ^uint16(0x041f)
		ppu.loopyV |= (ppu.loopyT & 0x041f)
	}

	// TODO: Should these be stored in 'ppu'?  Does it matter?
	var background [256]byte
	var sprites [256]byte
	var spriteHasPriority [256]bool

	if ppu.shouldRenderBackground() {
		ppu.renderBackground(&background)
	}

	if ppu.shouldRenderSprites() {
		ppu.renderSprites(&sprites, &spriteHasPriority, &background)
	}

	// Merge the sprite pixels and the bg pixels.
	for i := 0; i < DisplayWidth; i++ {
		// What is the resulting palette entry to render for this pixel after
		// background/sprite priority is resolved?
		var palEntry byte

		// See http://wiki.nesdev.com/w/index.php/PPU_rendering for the rules implemented
		// below.
		if isBG(background[i]) {
			if isBG(sprites[i]) {
				palEntry = 0
			} else {
				palEntry = 0x10 + sprites[i]
			}
		} else if isBG(sprites[i]) {
			palEntry = background[i]
		} else {
			if spriteHasPriority[i] {
				palEntry = 0x10 + sprites[i]
			} else {
				palEntry = background[i]
			}
		}

		bg := &Palette[ppu.pal[palEntry]]
		ppu.window.SetPixel(i, int(ppu.scanLineCounter), bg.r, bg.g, bg.b)
	}

	ppu.scanLineCounter++
}

// Render the sprites for the current scan line.  Output the palette selections into 'out'.  If the
// sprite that rendered the pixel at out[i] is of higher priority than the background, set
// priorityOut[i] to true.
//
// background is provided so that sprite 0 hit detection can be done here.
func (ppu *PPU) renderSprites(out *[256]byte, priorityOut *[256]bool, background *[256]byte) {
	var spriteSize uint16 = 8

	// This bit controls whether or not we using 8x16 sprites instead of the default 8x8.
	if 0x20 == (ppu.ppuCtrl & 0x20) {
		spriteSize = 16
	}

	// 256 bytes of sprite memory / 4 bytes per sprite == 64 sprites.
	for i := 0; i < 64; i++ {
		spriteOffset := i * 4

		// The first byte of the sprite data is the Y coordinate minus one.
		yCoord := uint16(ppu.oamData[spriteOffset]) + 1

		// If we're rendering before the sprite, bail out.
		if ppu.scanLineCounter < yCoord {
			continue
		}
		// If we're rendering after the sprite, bail out.
		if ppu.scanLineCounter >= yCoord + spriteSize {
			continue
		}

		// The second byte tells us what tile to use
		tileNoToRender := uint16(ppu.oamData[spriteOffset + 1])

		// The third byte is the sprite's attribute byte.  Has the upper two bits of the
		// 4-bit palette entry, sprite size, and priority.
		attr := ppu.oamData[spriteOffset + 2]

		// X position on the screen.
		xpos := byte(ppu.oamData[spriteOffset + 3])

		// Sprite pattern table that we use for each sprite.
		var ptBaseAddr uint16

		// For 8x8 sprites, ptBaseAddr is set via ppuCtrl.
		// For 8x16 sprites, ptBaseAddr is set per-sprite.
		if 8 == spriteSize {
			if 8 == (ppu.ppuCtrl & 8) {
				ptBaseAddr = 0x1000
			} else {
				ptBaseAddr = 0
			}
		} else {
			if 0 == tileNoToRender & 1 {
				ptBaseAddr = 0
			} else {
				// For 8x16 sprites, odd tile numbers are fetched from the pattern
				// table at 0x1000, and the lowest bit is dropped.
				ptBaseAddr = 0x1000
				tileNoToRender &= 0xfffe
			}
		}

		// Which line (0 to 7 inclusive) are we rendering in the sprite's tile?
		// Takes into account flipping and 8x16 sprites, possibly adjusting which
		// tile we're rendering.
		//
		// The sprite starts at 'yCoord' and goes to 'yCoord + spriteSize' and the line
		// we're rendering is somewhere in the [middle).
		var tileYOffset uint16

		// If this bit is set the tile is flipped vertically.
		if 0x80 == (attr & 0x80) {
			if 8 == spriteSize {
				// For one-tile sprites, this is easy.
				tileYOffset = 7 - (ppu.scanLineCounter - yCoord)
			} else {
				// TODO document better.
				yOffset := (ppu.scanLineCounter - yCoord)
				if yOffset < 8 {
					tileNoToRender++
					tileYOffset = 7 - yOffset
				} else {
					tileYOffset = 15 - yOffset
				}
			}
		} else {
			// No flipping.
			tileYOffset = ppu.scanLineCounter - yCoord
			if 16 == spriteSize && tileYOffset > 7 {
				// If we're rendering the [8, 16) line of a large sprite, the data
				// is actually in the next tile.
				tileYOffset -= 8
				tileNoToRender++
			}
		}

		// Each tile is 16 bytes, 8 bytes of the 0-th bit, and 8 bytes of the 1st bit.
		bit0addr := ptBaseAddr + tileNoToRender * 16 + tileYOffset
		tileBit0 := ppu.cartMapper.ReadPPU(bit0addr)
		tileBit1 := ppu.cartMapper.ReadPPU(bit0addr + 8)

		// All sprites are 8 pixels wide.
		for x := 0; x < 8; x++ {
			// What position are we writing to in this scan line?  We render the tile in
			// its "normal" order but write it backwards if horizontal flipping is on.
			//
			// This is not a uint8 because we want to detect when writes would wrap
			// around the 256-pixel screen.
			var xout uint16

			if 0x40 == (attr & 0x40) {
				// The sprite is flipped horizontally, so we write it backwards.
				xout = uint16(xpos) + uint16(7 - x)
			} else {
				// We write the tile's pixels in the "normal" order.
				xout = uint16(xpos) + uint16(x)
			}

			// Sprites do not wrap around the screen.
			if xout >= 256 {
				continue
			}

			// This flag controls whether or not sprites are shown in the left 8 pixels.
			if (0 == (ppu.ppuMask & 4)) && (xout < 8) {
				continue
			}

			// A higher priority sprite has already written to this pixel, so we cannot
			// overwrite the pixel data.
			// http://wiki.nesdev.com/w/index.php/PPU_sprite_priority
			if !isBG(out[xout]) {
				continue
			}

			// Decode the x-th pixel from the tile bytes and get the upper 2 palette
			// bits from the attribute.
			val := getPixel(tileBit0, tileBit1, x) | ((attr & 3) << 2)

			// This bit in 'attr' controls the priority of the sprite.  Sprites are
			// either high or low priority with respect to the background.  Among other
			// sprites, the one with the lowest index has the highest priority.  By
			// processing the sprites starting with 0, we ensure that the highest
			// priority sprite always has its data written.
			priorityOut[xout] = (0 == (attr & 0x20))

			// Store the pixel data.  Note that 'val' is a lookup into the sprite
			// palette which is 0x10 bytes after the bg palette, but this is accounted
			// for in RenderScanline(...) above
			out[xout] = val

			// Sprite 0 detection.  A CPU-checkable flag is set when the first
			// non-bgcolor background pixel overlaps a non-bg sprite pixel from sprite
			// number 0.  This isn't set for the 255-th pixel.
			if (0 == i) && (xout < 255) && !isBG(val) && !isBG(background[xout]) {
				// Set the "sprite 0 hit" flag.
				ppu.ppuStatus |= 0x40
			}
		}
	}
}

func (ppu *PPU) renderBackground(out *[256]byte) {
	// The pattern table used for rendering the background is set via ppuCtrl.
	var ptBaseAddr uint16
	if 0x10 == (ppu.ppuCtrl & 0x10) {
		ptBaseAddr = 0x1000
	} else {
		ptBaseAddr = 0x0000
	}

	// For the tile math below, recall that LoopyV looks like this:
	//
	// 0yyy NNYY YYYX XXXX
	// aka
	// yyy NN YYYYY XXXXX
	// ||| || ||||| +++++-- coarse X scroll
	// ||| || +++++-------- coarse Y scroll
	// ||| ++-------------- nametable select
	// +++----------------- fine Y scroll
	//
	// LoopyX is 3 bits of fine X scroll.

	for i := 0; i < DisplayWidth; i++ {
		// We're rendering this tile.
		tileNoToRender := uint16(ppu.cartMapper.ReadPPU(0x2000 | (ppu.loopyV & 0x0fff)))

		// But we're rendering this line of it.
		tileYOffset := 7 & (ppu.loopyV >> 12)

		// So we read these bytes.
		bit0Addr := ptBaseAddr + (tileNoToRender * 16) + tileYOffset
		tileBit0 := ppu.cartMapper.ReadPPU(bit0Addr)
		tileBit1 := ppu.cartMapper.ReadPPU(bit0Addr + 8)

		// And calculate the value of the tile.
		val := getPixel(tileBit0, tileBit1, int(ppu.loopyX))

		// Next, we look up the attribute byte information for the upper two bits of the
		// palette index.
		//
		// How the attrAddr is built:  0x23c0 (base attribute address) plus:
		//
		// NN 1111 YYY XXX
		// || |||| ||| +++-- high 3 bits of coarse X (x/4)
		// || |||| +++------ high 3 bits of coarse Y (y/4)
		// || ++++---------- attribute offset (960 bytes), included in 0x23c0
		// ++--------------- nametable select
		attrAddr := uint16(0x23c0)
		attrAddr |= (ppu.loopyV & 0x0c00)
		attrAddr |= ((ppu.loopyV >> 4) & 0x38)
		attrAddr |= ((ppu.loopyV >> 2) & 0x07)
		attrByte := ppu.cartMapper.ReadPPU(attrAddr)

		// The attribute tile represents a 32x32 pixel area.  So we need the lower 5 bits of
		// the X and Y scroll info (as 2^5 == 32) to figure out which bits in the attribute
		// tile we care about.

		// The lower 5 bits of the X scroll.
		attrXOffset := ppu.loopyX
		attrXOffset |= byte(ppu.loopyV & 3) << 3

		// The lower 5 bits of the Y scroll.
		attrYOffset := byte(tileYOffset)
		attrYOffset |= byte(ppu.loopyV >> 2) & 0x18

		if attrYOffset < 16 {
			// Bits [0...3] describe y = [0..15]
			attrByte &= 0xf
		} else {
			// Bits [4...7] describe y = [16..31]
			attrByte >>= 4
		}

		// At this point, attrByte is a 4-bit number.  Bits 0..1 describe the x = [0..15].
		// Since the attribute byte is supposed to provide the upper 2 bits of the palette
		// lookup value we output, we shift it left 2 bits so the attribute bits are in the
		// right position.
		//
		// Bits 2..3 describe x = [16..31] and are already in the right position.
		if attrXOffset < 16 {
			attrByte <<= 2
		}

		// Add the upper 2 palette lookup bits to val.
		val |= (attrByte & 0xC)

		// And write to the output buffer.
		if i >= 8 {
			out[i] = val
		} else if 2 == (ppu.ppuMask & 2) {
			// Background rendering in the left 8 pixels can be disabled by a flag in
			// ppuMask.
			out[i] = val
		}

		// Move to the next pixel.  loopyX is an offset into an 8-pixel tile.  If it hits 8,
		// we must move to the next tile.  Which tile we're rendering is stored in loopyV.
		if ppu.loopyX < 7 {
			ppu.loopyX++
		} else {
			ppu.loopyX = 0
			if 0x1f == (ppu.loopyV & 0x1f) {
				// We've hit the last tile in this name table, move to the next.
				ppu.loopyV &= ^uint16(0x1f)
				ppu.loopyV ^= 0x0400
			} else {
				// The lower 5 bits are the tile index, so it's OK to just increment
				// loopyV.
				ppu.loopyV++
			}
		}
	}

	// Move to the next line by incrementing the fine Y scroll value.
	// Recall that loopyV looks like:
	//
	// 0yyy NNYY YYYX XXXX
	// aka
	// yyy NN YYYYY XXXXX
	// ||| || ||||| +++++-- coarse X scroll
	// ||| || +++++-------- coarse Y scroll
	// ||| ++-------------- nametable select
	// +++----------------- fine Y scroll
	//
	if 0x7000 != (ppu.loopyV & 0x7000) {
		// The fine Y hasn't hit the max value (of 7)
		ppu.loopyV += 0x1000
	} else {
		// Fine Y will be reset to 0, and coarse Y must be incremented.
		ppu.loopyV &= ^uint16(0x7000)
		y := (ppu.loopyV & 0x03e0) >> 5

		// Coarse Y "normally" ranges from 0 to 29.
		if 29 == y {
			// If we hit the end, switch to the next nametable.
			y = 0
			ppu.loopyV ^= 0x800
		} else if 31 == y {
			y = 0
		} else {
			y++
		}

		ppu.loopyV = (ppu.loopyV & ^uint16(0x03e0)) | (y << 5)
	}
}
