package mapper

import "nesfile"

type Mapper2 struct {
	MapperAddressSpace

	// The PRG-ROM banks in the cart.
	prgRom [][]byte
}

func NewMapper2(nesFile *nesfile.NesFile) (Mapper) {
	out := new(Mapper2)

	out.prgRom = nesFile.PrgRom

	// CPU mappings.
	// First PRG-ROM bank is loaded at $8000.
	out.cpuPages[0] = nesFile.PrgRom[0]
	// Last PRG-ROM bank is loaded into $c000.
	out.cpuPages[1] = nesFile.PrgRom[len(nesFile.PrgRom) - 1]

	// THere shouldn't be any CHR-ROM, so pattern tables will be RAM.
	out.ppuPt0 = make([]byte, 0x1000)
	out.ppuPt1 = make([]byte, 0x1000)
	out.ppuPtIsROM = false

	out.MapperAddressSpace.setupNametables(nesFile)
	return out
}

func (mapper *Mapper2) WriteCPU(addr uint16, val uint8) (cycles uint64) {
	// Any write swaps in a 16k ROM bank at 0x8000
	mapper.cpuPages[0] = mapper.prgRom[val]
	return 0
}
