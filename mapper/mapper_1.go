package mapper

import "nesfile"

// Mapper1 works by writing a 5-bit value one bit at a time to any address in the PRG-ROM space.
// On the 5th write, the value is interpreted depending on the address written to.
//
// For details see http://wiki.nesdev.com/w/index.php/MMC1
type Mapper1 struct {
	MapperAddressSpace

	// The PRG-ROM banks in the cart.
	prgRom [][]byte

	// The CHR-ROM banks in the cart.
	chrRom [][]byte

	// The value being written one bit at a time into a shift register.
	shiftReg byte

	// What bit is being written next?
	whichBit byte

	// One of the mmc1 registers controls the behavior of the others, stored verbatim.
	//
	// From http://wiki.nesdev.com/w/index.php/MMC1
	//
	// 4bit0
	// -----
	// CPPMM
	// |||||
	// |||++- Mirroring (0: one-screen, lower bank; 1: one-screen, upper bank;
	// |||               2: vertical; 3: horizontal)
	// |++--- PRG ROM bank mode (0, 1: switch 32 KB at $8000, ignoring low bit of bank number;
	// |                         2: fix first bank at $8000 and switch 16 KB bank at $C000;
	// |                         3: fix last bank at $C000 and switch 16 KB bank at $8000)
	// +----- CHR ROM bank mode (0: switch 8 KB at a time; 1: switch two separate 4 KB banks)
	controlReg byte
}

func NewMapper1(nesFile *nesfile.NesFile) (Mapper) {
	out := new(Mapper1)

	// Keep a handle to the data we remap later.
	out.prgRom = nesFile.PrgRom
	out.chrRom = nesFile.ChrRom

	// Initial mappings:
	// First PRG-ROM bank is loaded at 0x8000, last into 0xc000
	out.cpuPages[0] = nesFile.PrgRom[0]
	out.cpuPages[1] = nesFile.PrgRom[len(nesFile.PrgRom) - 1]

	if 0 == len(nesFile.ChrRom) {
		// There is no ROM so we'll make some RAM.
		out.ppuPt0 = make([]byte, 0x1000)
		out.ppuPt1 = make([]byte, 0x1000)
		out.ppuPtIsROM = false
	} else {
		out.ppuPtIsROM = true
		out.ppuPt0 = nesFile.ChrRom[0][0:0x1000]
		out.ppuPt1 = nesFile.ChrRom[0][0x1000:0x2000]
	}

	out.MapperAddressSpace.setupNametables(nesFile)
	return out
}

func (mapper *Mapper1) WriteCPU(addr uint16, val uint8) (cycles uint64) {
	// TODO: Actually write SRAM to disk occasionally.
	if addr < 0x8000 {
		mapper.cpuSram[addr & 0x1fff] = val
		return 0
	}

	// Writing any value with the high bit set resets the shift register's value.
	if 0x80 == (val & 0x80) {
		mapper.controlReg |= 0x0c
		mapper.controlRegChanged()
		mapper.shiftReg = 0
		mapper.whichBit = 0
		return 0
	}

	// If the high bit isn't set we're adding to the shift register.
	mapper.shiftReg |= (val & 1) << mapper.whichBit
	mapper.whichBit++

	// The shift register is 5 bits wide and must be filled before being interpreted.
	if mapper.whichBit < 5 {
		return 0
	}

	// After 5 writes without a reset of the shift register, the shift register's value
	// is interpreted depending on the 5th write's address.

	// Reset the "which bit are we on" counter and apply the write to the register below.
	mapper.whichBit = 0

	if addr < 0xa000 {
		// 0x8000 -> 0x9fff, control register
		mapper.controlReg = mapper.shiftReg
		mapper.controlRegChanged()
	} else if addr < 0xc000 {
		// 0xa000 -> 0xbfff, controls CHR mapping at 0x0000
		mapper.remapChr0()
	} else if addr < 0xe000 {
		// 0xc000 -> 0xdfff
		mapper.remapChr1()
	} else {
		// 0xe000 -> 0xffff
		mapper.remapPrg()
	}

	mapper.shiftReg = 0
	return 0
}

func (mapper *Mapper1) controlRegChanged() {
	switch (mapper.controlReg & 3) {
	case 0:
		// One screen mirroring but the lower area of nametable memory.
		mapper.ppuNts[0] = mapper.ppuNtBank0[0:0x400]
		mapper.ppuNts[1] = mapper.ppuNts[0]
		mapper.ppuNts[2] = mapper.ppuNts[0]
		mapper.ppuNts[3] = mapper.ppuNts[0]
	case 1:
		// One screen mirroring but the higher area of nametable memory.
		mapper.ppuNts[0] = mapper.ppuNtBank1[0:0x400]
		mapper.ppuNts[1] = mapper.ppuNts[0]
		mapper.ppuNts[2] = mapper.ppuNts[0]
		mapper.ppuNts[3] = mapper.ppuNts[0]
	case 2:
		// Normal vertical mirroring.
		mapper.ppuNts[0] = mapper.ppuNtBank0[0:0x400]
		mapper.ppuNts[2] = mapper.ppuNts[0]
		mapper.ppuNts[1] = mapper.ppuNtBank1[0:0x400]
		mapper.ppuNts[3] = mapper.ppuNts[1]
	case 3:
		// Normal horizontal mirroring.
		mapper.ppuNts[0] = mapper.ppuNtBank0[0:0x400]
		mapper.ppuNts[1] = mapper.ppuNts[0]
		mapper.ppuNts[2] = mapper.ppuNtBank1[0:0x400]
		mapper.ppuNts[3] = mapper.ppuNts[2]
	}
}

func (mapper *Mapper1) remapChr0() {
	// TODO: If there's no ROM, is there ever remapping of RAM?
	if len(mapper.chrRom) == 0 {
		return
	}

	if 0 == (0x10 & mapper.controlReg) {
		// Remapping pages 8K at a time.  The low bit is dropped in this case.
		romIndex := int(mapper.shiftReg >> 1) % len(mapper.chrRom)
		// We're (re)mapping an 8k stretch to 0x0000 so both pt0 and pt1 are remapped.
		mapper.ppuPt0 = mapper.chrRom[romIndex][0:0x1000]
		mapper.ppuPt1 = mapper.chrRom[romIndex][0x1000:0x2000]
	} else {
		// Map 4k of data into 0x0000.  Only pt0 is remapped.
		// ChrRom is 8k pages and the shift register refers to a 4k page selection,
		// so we take the first or second half of the bank depending on shiftreg & 1
		romIndex := int(mapper.shiftReg >> 1) % len(mapper.chrRom)
		if 0 == (mapper.shiftReg & 1) {
			mapper.ppuPt0 = mapper.chrRom[romIndex][0:0x1000]
		} else {
			mapper.ppuPt0 = mapper.chrRom[romIndex][0x1000:0x2000]
		}
	}
}

func (mapper *Mapper1) remapChr1() {
	// This register is only used if we switch 4K CHR-ROM banks.
	if 0x10 == (0x10 & mapper.controlReg) {
		romIndex := int(mapper.shiftReg >> 1) % len(mapper.chrRom)
		if 0 == (mapper.shiftReg & 1) {
			mapper.ppuPt1 = mapper.chrRom[romIndex][0:0x1000]
		} else {
			mapper.ppuPt1 = mapper.chrRom[romIndex][0x1000:0x2000]
		}
	}
}

func (mapper *Mapper1) remapPrg() {
	// Bit no. 4 is the PRG RAM chip enable bit.  We ignore it here.
	mapper.shiftReg &= 0xf

	// The control register dictates how the remapping is done.
	prgRomBankMode := (mapper.controlReg >> 2) & 3

	if 0 == prgRomBankMode || 1 == prgRomBankMode {
		// Ignore the lowest bit of shiftReg, map 32kb to 0xc000
		bankIndex := mapper.shiftReg >> 1
		// mapper.prgRom pages are 16k so we map two adjacent pages.
		mapper.cpuPages[0] = mapper.prgRom[bankIndex * 2]
		mapper.cpuPages[1] = mapper.prgRom[bankIndex * 2 + 1]
	} else if 2 == prgRomBankMode {
		// Leave bank at 0x8000 alone and switch 0xc000
		mapper.cpuPages[1] = mapper.prgRom[int(mapper.shiftReg) % len(mapper.prgRom)]
	} else {
		// Leave bank at 0xc000 alone and switch 0x8000
		mapper.cpuPages[0] = mapper.prgRom[int(mapper.shiftReg) % len(mapper.prgRom)]
	}
}
