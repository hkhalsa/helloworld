package cpu

import (
	"testing"
)

// An implementation of the memory interface used in the tests below.
type MemoryForTesting struct {
	memory map[uint16]uint8
}

func NewMemoryForTesting() (*MemoryForTesting) {
	var ret = new(MemoryForTesting)
	ret.memory = make(map[uint16]uint8)
	return ret
}

func (mem *MemoryForTesting) Read(addr uint16) (val uint8) {
	return mem.memory[addr]
}

func (mem *MemoryForTesting) Write(addr uint16, val uint8) (cycles uint64) {
	mem.memory[addr] = val
	return 0
}

// Store a byte at an address and try to read it.
func TestReadPC8(t *testing.T) {
	var mymem = NewMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var addr uint16 = 0x1010
	var val uint8 = 0x80

	mymem.Write(addr, val)
	mycpu.pc = addr

	if mycpu.readPC8() != val {
		t.Fail()
	}
}

// Ensure that we're reading LSB-first
func TestReadPC16(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var addr uint16 = 0x1010
	var val_lsb uint8 = 0x11
	var val_msb uint8 = 0x22
	var val uint16 = 0x2211

	mymem.Write(addr, val_lsb)
	mymem.Write(addr + 1, val_msb)
	mycpu.pc = addr

	if mycpu.readPC16() != val {
		t.Fail()
	}
}

// Test immediate addressing aka reading from the PC
func TestAddressImmediate(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var addr uint16 = 0x1010
	var val uint8 = 0x80

	// PC points to 'addr'
	mycpu.pc = addr
	// memory at 'addr' has 'val'
	mymem.Write(addr, val)

	// Immediate addressing means we read the byte we want from the PC.
	mycpu.addrIMM()
	if mycpu.opAddr != addr {
		t.Fatal("opAddr isn't the same as the PC")
	}
}

// Test zero page absolute addressing.
func TestAddressAbsoluteZeroPage(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var addr uint16 = 0x1010
	var val uint8 = 0x80

	mycpu.pc = addr
	mymem.Write(addr, val)

	mycpu.addrZP()

	if mycpu.opAddr != uint16(val) {
		t.Fatal("opAddr isn't expected zero-page value")
	}
}

func TestZeroPageIndexedX(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var addr uint16 = 0x1010
	var val uint8 = 0xff

	mycpu.pc = addr
	mycpu.xr = 1
	mymem.Write(addr, val)

	mycpu.addrZPX()

	if mycpu.opAddr != 0 {
		t.Fatal("opAddr isn't expected zero-page + xr value, is ", mycpu.opAddr)
	}
}

func TestZeroPageIndexedY(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var addr uint16 = 0x1010
	var val uint8 = 0xc0

	mycpu.pc = addr
	mycpu.yr = 0x60
	mymem.Write(addr, val)

	mycpu.addrZPY()

	if mycpu.opAddr != 0x20 {
		t.Fatal("opAddr isn't expected zero-page + yr value, is", mycpu.opAddr)
	}
}

func TestAddressAbsolute(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var addr uint16 = 0x1010
	var val_lsb uint8 = 0x11
	var val_msb uint8 = 0x22
	var val uint16 = 0x2211

	mymem.Write(addr, val_lsb)
	mymem.Write(addr + 1, val_msb)
	mycpu.pc = addr

	mycpu.addrABS()

	if mycpu.opAddr != val {
		t.Fail()
	}
}

func TestAddressIndirect(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	// The address that the operand refers to is 0xc1ff.
	mymem.Write(0xc100, 0x11)
	mymem.Write(0xc101, 0x22)

	// The PC points here.
	mycpu.pc = 0xd001
	mymem.Write(0xd001, 0x00)
	mymem.Write(0xd002, 0xc1)

	mycpu.addrIND()

	if mycpu.opAddr != 0x2211 {
		t.Fail()
	}
}

func TestAddressIndirectNoCarryFromLowByte(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	// The address that the operand refers to is 0xc1ff.
	mymem.Write(0xc100, 0x4f)
	mymem.Write(0xc1ff, 0x00)
	mymem.Write(0xc200, 0x23)

	// The PC points here.
	mycpu.pc = 0xd001
	mymem.Write(0xd001, 0xff)
	mymem.Write(0xd002, 0xc1)

	mycpu.addrIND()

	if mycpu.opAddr != 0x4f00 {
		t.Fail()
	}
}

func TestAddressAbsoluteX(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var pcAddr uint16 = 0xc0c0
	// The address after the opcode will be 0x2211
	mymem.Write(pcAddr, 0x11)
	mymem.Write(pcAddr + 1, 0x22)
	mycpu.pc = pcAddr

	// XR will have 0x01
	mycpu.xr = 0x01

	mycpu.addrABSX(0)
	// 0x2211 + 0x01 == 0x2212
	if mycpu.opAddr != 0x2212 {
		t.Fail()
	}
}

func TestAddressAbsoluteY(t *testing.T) {
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)

	var pcAddr uint16 = 0xc0c0
	// The address after the opcode will be 0x2211
	mymem.Write(pcAddr, 0x11)
	mymem.Write(pcAddr + 1, 0x22)
	mycpu.pc = pcAddr

	// YR will have 0x01
	mycpu.yr = 0x02

	mycpu.addrABSY(0)
	// 0x2211 + 0x02 == 0x2213
	if mycpu.opAddr != 0x2213 {
		t.Fail()
	}
}

func TestAddressPreIndexedIndirect(t *testing.T) {
	// Set up base state
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)
	var pcAddr uint16 = 0xc0c0
	mycpu.pc = pcAddr

	mymem.Write(pcAddr, 0x3e)
	mycpu.xr = 0x05

	// The 16-bit address is read from 0x3e + 0x5 == 0x43.
	// The address is 0x2415
	mymem.Write(0x43, 0x15)
	mymem.Write(0x44, 0x24)

	mycpu.addrINDX()

	if mycpu.opAddr != 0x2415 {
		t.Fail()
	}
}

// Test that XR + zp addr are summed with 8bit math
func TestAddressPreIndexedIndirectZeroPage(t *testing.T) {
	// Set up base state
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)
	var pcAddr uint16 = 0xc0c0
	mycpu.pc = pcAddr

	mymem.Write(pcAddr, 0x60)
	mycpu.xr = 0xc0

	// The 16-bit address is read from 0xc0 + 0x60 == 0x120 == 0x20.
	// The address is 0x2415
	mymem.Write(0x20, 0x15)
	mymem.Write(0x21, 0x24)

	mycpu.addrINDX()

	if mycpu.opAddr != 0x2415 {
		t.Fail()
	}
}

// Test that reading the address across a page boundary is ok (that is, from 0xff and 0x100)
func TestAddressPreIndexedIndirectZeroPageToFirstPage(t *testing.T) {
	// Set up base state
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)
	var pcAddr uint16 = 0xc0c0
	mycpu.pc = pcAddr

	mymem.Write(pcAddr, 0x0f)
	mycpu.xr = 0xf0

	// The 16-bit address is read from 0xf0 + 0x0f == 0xff
	// The address is 0x2415
	mymem.Write(0xff, 0x15)
	mymem.Write(0x00, 0x24)

	mycpu.addrINDX()

	if mycpu.opAddr != 0x2415 {
		t.Fatal("opAddr is ", mycpu.opAddr)
	}
}

// Test indirect indexed aka ind_y().
func TestAddressPreIndexedIndirectIndexedY(t *testing.T) {
	// Set up base state
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)
	var pcAddr uint16 = 0xc0c0
	mycpu.pc = pcAddr

	mymem.Write(pcAddr, 0x0f)
	mycpu.yr = 0x01

	// The 16-bit address is read from 0x0f, and then xr is added.
	// The address is 0x2415 + yr(1) == 0x2416
	mymem.Write(0x0f, 0x15)
	mymem.Write(0x10, 0x24)

	mycpu.addrINDY(0)

	if mycpu.opAddr != 0x2416 {
		t.Fail()
	}
}

// Test that indirect indexed aka ind_y() wraps if the zp addr is 0xff
func TestAddressPreIndexedIndirectIndexedYWraps(t *testing.T) {
	// Set up base state
	var mymem = mem.MakeMemoryForTesting()
	var mycpu = NewCPU(mymem)
	var pcAddr uint16 = 0xc0c0
	mycpu.pc = pcAddr

	mymem.Write(pcAddr, 0xff)
	mycpu.yr = 0x01

	// The 16-bit address is read from 0xff (lower byte) and 0x00 (higher byte).
	// The address is 0x2415 + yr(1) == 0x2416
	mymem.Write(0x00, 0x24)
	mymem.Write(0xff, 0x15)

	mycpu.addrINDY(0)

	if mycpu.opAddr != 0x2416 {
		t.Fail()
	}
}
