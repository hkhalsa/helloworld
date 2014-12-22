package mapper

import (
	"nesfile"
)

// Mappers handle reads and writes to the remappable addressable space of both the CPU and PPU.
type Mapper interface {
	// These are handled by MapperAddressSpace below, which every mapper impl. should embed.
	ReadCPU(addr uint16) (val uint8)
	ReadPPU(addr uint16) (val uint8)
	WritePPU(addr uint16, val uint8) (cycles uint64)

	// This is probably the only method that mappers need to implement.  Remapping is done by
	// writing to the upper half of the CPU address space.
	WriteCPU(addr uint16, val uint8) (cycles uint64)

	// Enable/disable debugging output.
	Debug(on bool)
}

// Every mapper should embed this.
type MapperAddressSpace struct {
	//
	// The CPU cart address space is 0x4018 -> 0xFFFF.
	//

	// [0x4018 -> 0x5FFF] can be used by carts for various stuff,
	// called "expansion ROM" in some places, and ignored by me for now.

	// [0x6000 -> 0x7FFF] is SRAM, allocated but ignored for now, could be written out
	// and read in at some point.
	cpuSram [0x2000]byte

	// 0x8000 -> 0xFFFF is ROM.
	// Most mappers map this address space in 16K blocks to on-cart ROM, so that's what
	// we'll do here.
	cpuPages [2][]byte

	//
	// Aside from palette data, the PPU address space is entirely on the cart.
	//

	// 0x0000 -> 0x0FFF
	// Could be RAM or ROM depending on mapper.
	ppuPt0[]byte

	// 0x1000 -> 0x1FFF
	// Could be RAM or ROM depending on mapper.
	ppuPt1 []byte

	// The pattern tables can be ROM or RAM.  If they're ROM we'll complain if they're written.
	ppuPtIsROM bool

	// 0x2000 -> 0x2FFF
	// 0x3000 -> 0x3EFF mirrors this
	// Note the weird mirroring!  Palette data starts at PPU 0x3f00 and is in the PPU proper.
	ppuNts [4][]byte

	// Consider these the "physical" nametable pages.  The nametable address lines are mapped
	// to these depending on mirroring settings.  These are really in the PPU, but the mapping
	// is managed here, so we allocate them here is well.
	ppuNtBank0 [0x400]byte
	ppuNtBank1 [0x400]byte

	// Debug flag
	debug bool
}

func (mapper *MapperAddressSpace) Debug(val bool) {
	mapper.debug = val
}

func (mapper *MapperAddressSpace) ReadCPU(addr uint16) (val uint8) {
	if addr < 0x4018 {
		panic("too-low address passed to ReadCPU")
	} else if addr < 0x6000 {
		// 0x4018 -> 0x5FFF ignored
		return 0
	} else if addr < 0x8000 {
		// 0x6000 -> 0x7fff is SRAM
		return mapper.cpuSram[addr & 0x1fff]
	} else if addr < 0xc000 {
		// 0x8000 -> 0xbfff is the first 16k page of rom
		return mapper.cpuPages[0][addr & 0x3fff]
	} else {
		// 0xc000 -> 0xffff is the second 16k page of rom
		return mapper.cpuPages[1][addr & 0x3fff]
	}
}

func (mapper *MapperAddressSpace) ReadPPU(addr uint16) (val uint8) {
	if addr < 0x1000 {
		return mapper.ppuPt0[addr]
	} else if addr < 0x2000 {
		return mapper.ppuPt1[addr & 0xfff]
	} else if addr < 0x3f00 {
		addr &= 0x0fff
		return mapper.ppuNts[addr / 0x400][addr & 0x3ff]
	} else {
		panic("Trying to read palette data from on-cart PPU address mapper?")
		return 0
	}
}

func (mapper *MapperAddressSpace) WritePPU(addr uint16, val uint8) (cycles uint64) {
	if addr < 0x1000 {
		mapper.ppuPt0[addr] = val
	} else if addr < 0x2000 {
		mapper.ppuPt1[addr & 0xfff] = val
	} else if addr < 0x3f00 {
		addr &= 0x0fff
		mapper.ppuNts[addr / 0x400][addr & 0x3ff] = val
	} else {
		panic("Trying to write palette data from on-cart PPU address mapper?")
	}
	return 0
}

// The nametable mirroring is specified in the iNES file header.  Mapper implementations
// should use this function as part of their initialization.
func (mas *MapperAddressSpace) setupNametables(nesFile *nesfile.NesFile) {
	if nesfile.Horizontal == nesFile.Mirroring {
		mas.ppuNts[0] = mas.ppuNtBank0[0:0x400]
		mas.ppuNts[1] = mas.ppuNts[0]
		mas.ppuNts[2] = mas.ppuNtBank1[0:0x400]
		mas.ppuNts[3] = mas.ppuNts[2]
	} else if nesfile.Vertical == nesFile.Mirroring {
		mas.ppuNts[0] = mas.ppuNtBank0[0:0x400]
		mas.ppuNts[2] = mas.ppuNts[0]
		mas.ppuNts[1] = mas.ppuNtBank1[0:0x400]
		mas.ppuNts[3] = mas.ppuNts[1]
	} else {
		panic("I can't deal with the non-horizontal non-vertical mirroring yet")
	}
}

// An entry in the table of mappers.
type MapperEntry struct {
	ctor func(nesfile *nesfile.NesFile) (Mapper)
}

// All implemented mappers go here.
var mapperTable = map[int]*MapperEntry {
	0: { NewMapper0 },
	1: { NewMapper1 },
	2: { NewMapper2 },
	3: { NewMapper3 },
}

// Allocate the correct Mapper and return it.  Die if we can't provide the mapper.
func GetMapper(nesFile *nesfile.NesFile) (out Mapper) {
	if mapper, ok := mapperTable[nesFile.Mapper]; ok {
		out = mapper.ctor(nesFile)
	} else {
		panic("TODO: Implement more mappers")
	}
	return
}
