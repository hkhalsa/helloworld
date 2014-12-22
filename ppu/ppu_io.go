package ppu

// Internally used constants.  Each of these addresses is in CPU space and corresponds to a
// different PPU register.
const (
	PPUCTRL   = 0x2000
	PPUMASK   = 0x2001
	PPUSTATUS = 0x2002
	OAMADDR   = 0x2003
	OAMDATA   = 0x2004
	PPUSCROLL = 0x2005
	PPUADDR   = 0x2006
	PPUDATA   = 0x2007
)

// If the CPU writes to 0x4014, it triggers a DMA copy from CPU address space to the sprite memory.
func (ppu *PPU) SpriteDMA(src []byte) {
	if 256 != copy(ppu.oamData[:], src) {
		panic("Could not copy 256 bytes via SpriteDMA")
	}
}

// PPU registers are mapped to 0x2000 -> 0x3FFF in the CPU address space.  Reads to those
// addresses (in CPU address space) wind up here.
func (ppu *PPU) ReadRegister(mmioreg uint16)(val uint8) {
	switch(mmioreg & 0x2007) {
	case PPUCTRL:
		return ppu.ppuCtrl
	case PPUMASK:
		return ppu.ppuMask
	case PPUSTATUS:
		// Save the pre-side-effect value for output.
		val = ppu.ppuStatus
		// Side-effects upon reading: the VBlank bit is cleared,
		ppu.ppuStatus &= 0x7f
		// and the address latch is reset.
		ppu.addressLatch = 0
		return
	case OAMADDR:
		return ppu.oamAddr
	case OAMDATA:
		val = ppu.oamData[ppu.oamAddr]
		return
	case PPUSCROLL:
		// Write-only.
		return 0
	case PPUADDR:
		// Write-only.
		return 0
	case PPUDATA:
		// Return what's in the PPU read buffer.
		val = ppu.bufferedReadData

		// This address space is handled by the cart mapper.
		if ppu.loopyV < 0x3f00 {
			ppu.bufferedReadData = ppu.cartMapper.ReadPPU(ppu.loopyV)
		} else {
			// Palette memory is in the PPU.
                        addr := ppu.loopyV & 0x1f
                        val = ppu.pal[addr]
			// The internal VRAM buffer is still filled in (does this matter?)
			// bufAddr := ppu.loopyV & 0x0fff
                        // ppu.bufferedReadData = ppu.nts[bufAddr / 0x400][bufAddr & 0x3ff]
                }

		ppu.incVRAMAddress()
		return
	default:
		panic("trying to read unknown ppu reg")
	}
}

// The internal address counter is incremented during PPUDATA accesses.
func (ppu *PPU) incVRAMAddress() {
	// The 2nd bit in ppuCtrl dictates how we increment the VRAM address after it's used to
	// index into memory.
	if 0 == (ppu.ppuCtrl & 4) {
		ppu.loopyV++
	} else {
		ppu.loopyV += 32
	}
}

// PPU registers are mapped to 0x2000 -> 0x3FFF in the CPU address space.  Writes to those
// addresses (in CPU address space) wind up here.
func (ppu *PPU) WriteRegister(mmioreg uint16, val uint8) {
	switch(mmioreg & 0x2007) {
	case PPUCTRL:
		ppu.ppuCtrl = val
		ppu.loopyT &= ^uint16(0x0c00)
		ppu.loopyT |= uint16(val & 3) << 10
	case PPUMASK:
		ppu.ppuMask = val
	case PPUSTATUS:
		// Ignored, read-only register, not sure this should happen, maybe panic?
	case OAMADDR:
		ppu.oamAddr = val
	case OAMDATA:
		ppu.oamData[ppu.oamAddr] = val
		// TODO: reads during VBlank don't increment, do we care?
		ppu.oamAddr++
	case PPUSCROLL:
		if 0 == ppu.addressLatch {
			ppu.loopyX = val & 7
			ppu.loopyT &= ^uint16(0x1F)
			ppu.loopyT |= (0x1F & uint16(val >> 3))
			ppu.addressLatch = 1
		} else {
			ppu.loopyT &= ^uint16(0x73E0)
			ppu.loopyT |= uint16(val & 7) << 12
			ppu.loopyT |= uint16(val & 0xF8) << 2
			ppu.addressLatch = 0
		}
	case PPUADDR:
		if 0 == ppu.addressLatch {
			ppu.loopyT &= 0x00ff
			ppu.loopyT |= uint16(val & 0x3f) << 8
			ppu.addressLatch = 1
		} else {
			ppu.loopyT &= 0xff00
			ppu.loopyT |= uint16(val)
			ppu.loopyV = ppu.loopyT
			ppu.addressLatch = 0
		}
	case PPUDATA: // 0x2007
		if ppu.loopyV < 0x3f00 {
			ppu.cartMapper.WritePPU(ppu.loopyV, val)
                } else {
			addr := ppu.loopyV & 0x1f
			// The upper 2 bits of the palette are ignored so we mask them out here.
			val &= 0x3f
			ppu.pal[addr] = val
			if 0 == (addr & 3) {
				// 0x3f00 and 0x3f10 are mirrored to each other.
				// 0x3f04 and 0x3f14 are mirrored to each other.
				// 0x3f08 and 0x3f18 are mirrored to each other.
				// 0x3f0c and 0x3f1c are mirrored to each other.
				ppu.pal[addr ^ 0x10] = val
			}
                }
		ppu.incVRAMAddress()
	default:
		panic("trying to read unknown ppu reg")
	}
}
