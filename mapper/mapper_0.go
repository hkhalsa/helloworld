package mapper

import "nesfile"

type Mapper0 struct {
	MapperAddressSpace
}

func NewMapper0(nesFile *nesfile.NesFile) (Mapper) {
	out := new(Mapper0)

	// CPU mappings.
	out.cpuPages[0] = nesFile.PrgRom[0]
	// If there is one page, it's mapped at 0x8000 and 0xc000
	// If there is more than one page, map the last.  Should only be two.
	out.cpuPages[1] = nesFile.PrgRom[len(nesFile.PrgRom) - 1]

	// PPU mappings.
	if 0 == len(nesFile.ChrRom) {
		// No CHR-ROM so we assume it's CHR-RAM for pattern tables.
		out.ppuPt0 = make([]byte, 0x1000)
		out.ppuPt1 = make([]byte, 0x1000)
		out.ppuPtIsROM = false
	} else {
		out.ppuPtIsROM = true
		// There may be multiple CHR-ROM banks but there are no provisions for switching
		// between them in this mapper, so we just point into the first bank.
		out.ppuPt0 = nesFile.ChrRom[0][0:0x1000]
		out.ppuPt1 = nesFile.ChrRom[0][0x1000:0x2000]
	}

	out.MapperAddressSpace.setupNametables(nesFile)

	return out
}

func (mapper *Mapper0) WriteCPU(addr uint16, val uint8) (cycles uint64) {
	// No remapping with mapper 0.
	return 0
}
