package mapper

import "nesfile"

type Mapper3 struct {
	MapperAddressSpace

	// The CHR-ROM banks in the cart.
	chrRom [][]byte
}

func NewMapper3(nesFile *nesfile.NesFile) (Mapper) {
	out := new(Mapper3)

	// CPU mappings.
	// First PRG-ROM bank is loaded at $8000.
	out.cpuPages[0] = nesFile.PrgRom[0]
	// Last PRG-ROM bank is loaded into $c000.  There are 1 or 2 PRG-ROM banks.
	out.cpuPages[1] = nesFile.PrgRom[len(nesFile.PrgRom) - 1]

	// The first CHR-ROM bank is loaded.
	out.ppuPt0 = nesFile.ChrRom[0][0:0x1000]
	out.ppuPt1 = nesFile.ChrRom[0][0x1000:0x2000]

	// CHR-ROM can be remapped.
	out.chrRom = nesFile.ChrRom

	out.MapperAddressSpace.setupNametables(nesFile)
	return out
}

func (mapper *Mapper3) WriteCPU(addr uint16, val uint8) (cycles uint64) {
	// Any write swaps in an 8K VROM bank at 0x0000.  Only the lower 2 bits are used.
	mapper.ppuPt0 = mapper.chrRom[val & 3][0:0x1000]
	mapper.ppuPt1 = mapper.chrRom[val & 3][0x1000:0x2000]
	return 0
}
