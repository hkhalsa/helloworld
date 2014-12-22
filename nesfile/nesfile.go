package nesfile

// This package parses the iNES file format into a NesFile structure.

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// See the PPU package for details on what mirroring means.
type Mirroring int

// Possible mirroring values.
const (
	Horizontal = iota
	Vertical
	FourScreen
)

// The dump of the NES file.
type NesFile struct {
	// Some details of the PPU address space mapping are specified in the header.
	Mirroring int

	// True if there is a battery-backed RAM available.
	SramEnabled bool

	// What address mapping hardware is in the cart?  Each set of address mapping
	// hardware has a number that identifies it.
	Mapper int

	// Each bank of PrgRom is 16K.
	PrgRom [][]byte

	// Each bank of ChrRom is 8K.
	ChrRom [][]byte
}

// Read from 'file' into 'target' and die on error.
func readAndCheck(file *os.File, target []byte) {
	bytesRead, err := file.Read(target)

	if len(target) != bytesRead {
		fmt.Println("Wanted to read ", len(target), " bytes, got ", bytesRead)
	}

	if nil != err {
		log.Fatal(err)
	}
}

// Read the provided file in iNES format and output it.  Dies if the file is malformed.
func ReadNesFile(fileName string) (nesFile *NesFile) {
	nesFile = new(NesFile)

	// Open the provided file.
	file, err := os.Open(fileName);
	if nil != err {
		log.Fatal(err)
	}

	// Read the full 16-byte header.
	fileHeader := make([]byte, 16)
	readAndCheck(file, fileHeader)

	// Look for the magic value in the header: 'NES\x1a'
	canonicalHeader := []byte{'N', 'E', 'S', '\x1a'}
	if !bytes.Equal(canonicalHeader, fileHeader[0:4]) {
		panic("error reading iNES file: first 4 bytes not magic value")
	}

	// Read mirroring information from the header.  Look in the PPU package for details on what
	// this means.
	if 0 == (fileHeader[6] & 1) {
		nesFile.Mirroring = Horizontal
	} else {
		nesFile.Mirroring = Vertical
	}

	if 0 != fileHeader[6] & 8 {
		nesFile.Mirroring = FourScreen
	}

	// Not impl'd for now but might as well set it correctly.
	nesFile.SramEnabled = (2 == (fileHeader[6] & 2))

	// The mapper is a byte whose nibbles are in the high 4 bits of each ROM control byte.
	nesFile.Mapper = int((0xf & (fileHeader[6]>>4)) | (0xf0 & fileHeader[7]))

	// I would be surprised if anyone ever had this, but it can happen.
	hasTrainer := 0 != (fileHeader[6] & (1<<2))
	if hasTrainer {
		// There's a 512-byte thing to skip over.
		ignoredTrainer := make([]byte, 512)
		readAndCheck(file, ignoredTrainer)
	}

	// Read in the ROM banks.
	nesFile.PrgRom = make([][]byte, int(fileHeader[4]))
	for i := range nesFile.PrgRom {
		nesFile.PrgRom[i] = make([]byte, 1 << 14)
		readAndCheck(file, nesFile.PrgRom[i])
	}

	nesFile.ChrRom = make([][]byte, int(fileHeader[5]))
	for i := range nesFile.ChrRom {
		nesFile.ChrRom[i] = make([]byte, 1 << 13)
		readAndCheck(file, nesFile.ChrRom[i])
	}

	return
}
