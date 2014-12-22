package cpu

// Opcodes select one of several addressing modes to determine the address of the data operand.
// The calculation of the final address depends only on the addressing mode, not on the opcode.
// So, as a preamble to the actual execution of an opcode, we populate 'cpu.opAddr' with the
// fully resolved address of the operand to the opcode.

// All addressing modes.  See the addrXXXX functions below for full details.
const (
	BAD = iota
	// Implied means there is no memory access (eg shift a register)
	IMP
	// Immediate means the byte we care about is right after the opcode.
	IMM
	ZP
	ZPX
	ZPY
	ABS
	ABSX
	ABSY
	IND
	INDY
	INDX
)

// Calculate the full address that 'op' refers to and store it in 'cpu.opAddr'.
// Also places the literal address in the opcode into 'cpu.opRawAddr'.
func (cpu *CPU) readAddressOfOperand(op *OpcodeEntry) {
	switch(op.addressing) {
	case BAD:
		// noop
	case IMP:
		// noop
	case IMM:
		cpu.addrIMM()
	case ZP:
		cpu.addrZP()
	case ZPX:
		cpu.addrZPX()
	case ZPY:
		cpu.addrZPY()
	case ABS:
		cpu.addrABS()
	case ABSX:
		cpu.addrABSX(op.extraCycles)
	case ABSY:
		cpu.addrABSY(op.extraCycles)
	case IND:
		cpu.addrIND()
	case INDY:
		cpu.addrINDY(op.extraCycles)
	case INDX:
		cpu.addrINDX()
	}
}

// Read one byte from *pc and increment pc.
func (cpu *CPU) readPC8() (out uint8) {
	out = cpu.mem.Read(cpu.pc)
	cpu.pc++
	return out
}

// Read a 16-bit word from *pc.  The 6502 is LSB first.
func (cpu *CPU) readPC16() (out uint16) {
	out = uint16(cpu.readPC8()) | (uint16(cpu.readPC8()) << 8)
	return out
}

// Immediate addressing.  The 8-bit operand is next to the opcode, so we just read from the PC.
func (cpu *CPU) addrIMM() {
	cpu.opAddr = cpu.pc
	cpu.pc++
}

// Absolute zero-page addressing.  A one-byte address is specified as part of the opcode.
// Terminology note: zero page means the address is on the 0-th 256-byte page of memory.
// Since the full address is 0x00?? only one byte needs to be used to specify the address.
func (cpu *CPU) addrZP() {
	cpu.opRawAddr = uint16(cpu.readPC8())
	cpu.opAddr = cpu.opRawAddr
}

// Zero-page indexed X.  The zero-page address given is added to the X register to give the actual
// address, which is itself zero-page.  As an example, xr=1 and *pc=0xff implies that opAddr=0.
func (cpu *CPU) addrZPX() {
	cpu.opRawAddr = uint16(cpu.readPC8())
	cpu.opAddr = cpu.opRawAddr + uint16(cpu.xr)
	cpu.opAddr &= 0xff
}

// Zero-page indexed Y.  The zero-page address given is added to the Y register to give the actual
// address, which is itself zero-page.
func (cpu *CPU) addrZPY() {
	cpu.opRawAddr = uint16(cpu.readPC8())
	cpu.opAddr = cpu.opRawAddr + uint16(cpu.yr)
	cpu.opAddr &= 0xff
}

// Absolute.  The 16-bit address is given in full following the opcode.
func (cpu *CPU) addrABS() {
	cpu.opRawAddr = cpu.readPC16()
	cpu.opAddr = cpu.opRawAddr
}

// Indirect addressing.  This is only used with JMP.  There is a bug in the hardware that we
// recreate (and unit test).  To quote nestech.txt:
//
// The 6502 has a bug in opcode $6C (jump absolute indirect). The CPU does
// not correctly calculate the effective address if the low-byte is $FF.
//
// Example:
//
// C100: 4F
// C1FF: 00
// C200: 23
// ..
// D000: 6C FF C1 - JMP ($C1FF)
//
// Logically, this will jump to address $2300. However, due to the fact
// that the high-byte of the calculate address is *NOT* increased on a
// page-wrap, this will actually jump to $4F00.
func (cpu *CPU) addrIND() {
	tmpAddr := cpu.readPC16()
	cpu.opRawAddr = tmpAddr
	lowByte := cpu.mem.Read(tmpAddr)

	// We add 1 to tmpAddr to get the address of the higher-order byte for the address.
	// However, there isn't a carry from the lower byte to the higher byte, so if we'd carry out
	// of the lower-order byte, instead wrap to zero.
	if 0xff == (tmpAddr & 0xff) {
		tmpAddr &= 0xff00
	} else {
		tmpAddr++
	}

	highByte := cpu.mem.Read(tmpAddr)
	cpu.opAddr = uint16(lowByte) | (uint16(highByte) << 8);
}

// Absolute + X.  A 16-bit address follows the opcode.  We add it to the XR.
func (cpu *CPU) addrABSX(extraCycles uint64) {
	cpu.opRawAddr = cpu.readPC16()
	cpu.opAddr = cpu.opRawAddr + uint16(cpu.xr)
	if (cpu.opRawAddr & 0xff00) != (cpu.opAddr & 0xff00) {
		cpu.clockCycles += extraCycles
	}
}

// Absolute + Y.
func (cpu *CPU) addrABSY(extraCycles uint64) {
	cpu.opRawAddr = cpu.readPC16()
	cpu.opAddr = cpu.opRawAddr + uint16(cpu.yr)
	if (cpu.opRawAddr & 0xff00) != (cpu.opAddr & 0xff00) {
		cpu.clockCycles += extraCycles
	}
}

// Pre-indexed indirect.  A zero-page address is added to xr to give the zero-page address of the
// bytes holding the address of the operand.
func (cpu *CPU) addrINDX() {
	// The addition of the zp addr and the xr are done as uint8 as the resulting address
	// is on the zero page.  All reads are done from the zero page.
	cpu.opRawAddr = uint16(cpu.readPC8())
	zpaddr := 0xff & (cpu.opRawAddr + uint16(cpu.xr))
	cpu.opAddr = uint16(cpu.mem.Read(zpaddr))
	cpu.opAddr |= (uint16(cpu.mem.Read(0xff & (zpaddr + 1))) << 8)
}

// Post-indexed indirect.  A zero-page address is provided in the opcode.  A 16-bit address is
// read from there and the yr is added to it to obtain the target address.
func (cpu *CPU) addrINDY(extraCycles uint64) {
	cpu.opRawAddr = uint16(cpu.readPC8())
	cpu.opAddr = uint16(cpu.mem.Read(cpu.opRawAddr))
	// The high byte is located on the zero page as well, so we wrap opRawAddr+1.
	cpu.opAddr |= (uint16(cpu.mem.Read(0xff & (cpu.opRawAddr + 1))) << 8)
	addrPreYRegister := cpu.opAddr
	cpu.opAddr += uint16(cpu.yr)
	if (cpu.opAddr & 0xff00) != (addrPreYRegister & 0xff00) {
		cpu.clockCycles += extraCycles
	}
}
