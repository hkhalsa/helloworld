package cpu

// Read the argument to the opcode.  Assumes opAddr is already filled according to the
// addressing mode implied by the opcode.
func (cpu *CPU) readOpData() uint8 {
	return cpu.mem.Read(cpu.opAddr)
}

// Push the provided byte on the stack.
func (cpu *CPU) push(value uint8) {
	var addr uint16 = 0x100 + uint16(cpu.sp)
	cpu.clockCycles += cpu.mem.Write(addr, value)
	cpu.sp--
}

// Pop the top element from the stack and return it.
func (cpu *CPU) pop() uint8 {
	cpu.sp++
	var addr uint16 = 0x100 + uint16(cpu.sp)
	return cpu.mem.Read(addr)
}

// Push a two byte word to the stack.
func (cpu *CPU) pushWord(word uint16) {
	cpu.push(uint8(0xff & (word >> 8)))
	cpu.push(uint8(0xff & word))
}

// Pop a two byte word from the stack.
func (cpu *CPU) popWord() uint16 {
	return uint16(cpu.pop()) | (uint16(cpu.pop()) << 8)
}

// Generic branching routine.  The logic for applying the branch offset and incurring
// extra clock cycles is the same for every branching operation.
func (cpu *CPU) genericBranch(should bool) {
	if !should {
		return
	}

	offset := uint16(cpu.readOpData())

	// Sign-extend offset from 8 bits to 16 bits so that a negative branch is applied
	// correctly.
	if 0x80 == (0x80 & offset) {
		offset |= 0xff00
	}

	newPC := cpu.pc + offset

	// If the new PC is on a different page the branch takes longer.
	if (newPC & 0xff00) != (cpu.pc & 0xff00) {
		cpu.clockCycles += 2
	} else {
		cpu.clockCycles += 1
	}

	cpu.pc = newPC
}

// Flags for the status register
const (
	// Carry flag
	C uint8 = 0x01

	// Zero flag
	Z uint8 = 0x02

	// Interrupt enable/disable flag
	I uint8 = 0x04

	// Decimal flag.  Decimal not supported on the NES, thankfully.
	D uint8 = 0x08

	// Set when BRK is executed.
	B uint8 = 0x10

	// 0x20 is unused but supposed to be 1 at all times.
	ALWAYS_ON uint8 = 0x20

	// Overflow.  When a math operation overflows.
	V uint8 = 0x40

	// Sign flag.
	N uint8 = 0x80
)

// Is the flag 'flag' set?  'flag' should be an OR-ing of the flag consts above.
func (cpu *CPU) isSet(flag uint8) bool {
	return flag == (cpu.st & flag)
}

// Set the flag 'flag' if 'on' is true, clear it otherwise.
func (cpu *CPU) set(flag uint8, on bool) {
	if on {
		cpu.st |= flag
	} else {
		cpu.st &= ^flag
	}
}

// Set the Z and N flags from the result.
func (cpu *CPU) setZN(result uint8) {
	// Z is set when a result is zero.
	if 0 == result {
		cpu.st |= Z
	} else {
		cpu.st &= ^Z
	}

	// N is set when the (2's complement) number is negative.
	if 0x80 == (result & 0x80) {
		cpu.st |= N
	} else {
		cpu.st &= ^N
	}
}

//
// Opcodes below.  See 6502 documentation for details -- too tedius to comment every single
// one.
//

func (cpu *CPU) opAac() {
	cpu.ac &= cpu.readOpData()
	cpu.set(C, 0x80 == (cpu.ac & 0x80))
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opAax() {
	data := cpu.ac & cpu.xr
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
}

func (cpu *CPU) opAdc() {
	opData := uint16(cpu.readOpData())
	result := uint16(cpu.ac) + uint16(opData)
	if cpu.isSet(C) {
		result++
	}

	// V is complicated, see http://forums.nesdev.com/viewtopic.php?t=6331.
	cpu.set(V, 0 != ((uint16(cpu.ac) ^ result) & (opData ^ result) & 0x80))
	cpu.set(C, result > 0xff)
	cpu.ac = uint8(result & 0xff)
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opAnd() {
	cpu.ac &= cpu.readOpData()
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opAslAcc() {
	cpu.set(C, 0x80 == (cpu.ac & 0x80))
	cpu.ac <<= 1
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opAsl() {
	data := cpu.readOpData()
	cpu.set(C, 0x80 == (data & 0x80))
	data <<= 1
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
	cpu.setZN(data)
}

func (cpu *CPU) opAsr() {
	cpu.ac &= cpu.readOpData()
	cpu.set(C, 1 == (cpu.ac & 1))
	cpu.ac >>= 1
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opArr() {
	cpu.ac = (cpu.ac & cpu.readOpData()) >> 1

	if cpu.isSet(C) {
		cpu.ac |= 0x80
	}

	cpu.setZN(cpu.ac)
	cpu.set(C, 0x40 == cpu.ac & 0x40)

	bit6IsSet := 0x40 == (cpu.ac & 0x40)
	bit5IsSet := 0x20 == (cpu.ac & 0x20)
	cpu.set(V, bit6IsSet != bit5IsSet)
}

func (cpu *CPU) opAtx() {
	cpu.ac = cpu.readOpData()
	cpu.xr = cpu.ac
	cpu.setZN(cpu.xr)
}

func (cpu *CPU) opAxs() {
	result := uint16(cpu.xr & cpu.ac) - uint16(cpu.readOpData())
	cpu.set(C, result < 0x100)
	cpu.xr = uint8(result & 0xff)
	cpu.setZN(cpu.xr)
}

func (cpu *CPU) opBcc() {
	cpu.genericBranch(!cpu.isSet(C))
}

func (cpu *CPU) opBcs() {
	cpu.genericBranch(cpu.isSet(C))
}

func (cpu *CPU) opBeq() {
	cpu.genericBranch(cpu.isSet(Z))
}

func (cpu *CPU) opBit() {
	data := cpu.readOpData()
	cpu.set(Z, 0 == (data & cpu.ac))
	cpu.set(V, 0x40 == (data & 0x40))
	cpu.set(N, 0x80 == (data & 0x80));
}

func (cpu *CPU) opBmi() {
	cpu.genericBranch(cpu.isSet(N))
}

func (cpu *CPU) opBne() {
	cpu.genericBranch(!cpu.isSet(Z))
}

func (cpu *CPU) opBpl() {
	cpu.genericBranch(!cpu.isSet(N))
}

func (cpu *CPU) opBrk() {
	cpu.pushWord(cpu.pc + 1)
	cpu.set(B, true)
	cpu.push(ALWAYS_ON | cpu.st)
	cpu.set(I, true)
	cpu.pc = uint16(cpu.mem.Read(vectorIRQBRK)) | (uint16(cpu.mem.Read(vectorIRQBRK + 1)) << 8)
}

func (cpu *CPU) opBvc() {
	cpu.genericBranch(!cpu.isSet(V))
}

func (cpu *CPU) opBvs() {
	cpu.genericBranch(cpu.isSet(V))
}

func (cpu *CPU) opClc() {
	cpu.set(C, false)
}

func (cpu *CPU) opCld() {
	cpu.set(D, false)
}

func (cpu *CPU) opCli() {
	cpu.set(I, false)
}

func (cpu *CPU) opClv() {
	cpu.set(V, false)
}

func (cpu *CPU) opCmp() {
	data := cpu.readOpData()
	cpu.set(C, cpu.ac >= data)
	cpu.setZN(cpu.ac - data)
}

func (cpu *CPU) opCpx() {
	data := cpu.readOpData()
	cpu.set(C, cpu.xr >= data)
	cpu.setZN(cpu.xr - data)
}

func (cpu *CPU) opCpy() {
	data := cpu.readOpData()
	cpu.set(C, cpu.yr >= data)
	cpu.setZN(cpu.yr - data)
}

func (cpu *CPU) opDcp() {
	data := cpu.readOpData() - 1
	cpu.set(C, cpu.ac >= data)
	cpu.setZN(cpu.ac - data)
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
}

func (cpu *CPU) opDec() {
	data := cpu.readOpData() - 1
	cpu.setZN(data)
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
}

func (cpu *CPU) opDex() {
	cpu.xr--
	cpu.setZN(cpu.xr)
}

func (cpu *CPU) opDey() {
	cpu.yr--
	cpu.setZN(cpu.yr)
}

func (cpu *CPU) opEor() {
	cpu.ac ^= cpu.readOpData()
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opInc() {
	data := cpu.readOpData() + 1
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
	cpu.setZN(data)
}

func (cpu *CPU) opInx() {
	cpu.xr++
	cpu.setZN(cpu.xr)
}

func (cpu *CPU) opIny() {
	cpu.yr++
	cpu.setZN(cpu.yr)
}

func (cpu *CPU) opIsc() {
	opData := uint16(cpu.readOpData())
	opData = 0xff & (opData + 1)
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, uint8(opData))

	// TODO: This is cribbed from opSbc, just factor out?
	result := uint16(cpu.ac) - opData
	if !cpu.isSet(C) {
		result--
	}

	cpu.set(V, 0 != ((uint16(cpu.ac) ^ result) & (uint16(cpu.ac) ^ opData) & 0x80))
	cpu.set(C, result < 0x100)
	cpu.ac = uint8(result & 0xff)
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opJmp() {
	cpu.pc = cpu.opAddr
}

func (cpu *CPU) opJsr() {
	cpu.pushWord(cpu.pc - 1)
	cpu.pc = cpu.opAddr
}

func (cpu *CPU) opLax() {
	cpu.ac = cpu.readOpData()
	cpu.xr = cpu.ac
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opLda() {
	cpu.ac = cpu.readOpData()
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opLdx() {
	cpu.xr = cpu.readOpData()
	cpu.setZN(cpu.xr)
}

func (cpu *CPU) opLdy() {
	cpu.yr = cpu.readOpData()
	cpu.setZN(cpu.yr)
}

func (cpu *CPU) opLsrAcc() {
	cpu.set(C, 1 == (cpu.ac & 1))
	cpu.ac >>= 1
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opLsr() {
	data := cpu.readOpData()
	cpu.set(C, 1 == (data & 1))
	data >>= 1
	cpu.setZN(data)
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
}

func (cpu *CPU) opNop() { }

func (cpu *CPU) opOra() {
	cpu.ac |= cpu.readOpData()
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opPha() {
	cpu.push(cpu.ac)
}

func (cpu *CPU) opPhp() {
	cpu.push(ALWAYS_ON | cpu.st | B)
}

func (cpu *CPU) opPla() {
	cpu.ac = cpu.pop()
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opPlp() {
	cpu.st = ALWAYS_ON | (cpu.pop() & ^B)
}

func (cpu *CPU) opRla() {
	data := cpu.readOpData()
	carry := cpu.isSet(C)
	cpu.set(C, 0x80 == (data & 0x80))
	data <<= 1
	if carry {
		data |= 1
	}
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
	cpu.ac &= data
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opRolAcc() {
	carry := cpu.isSet(C)
	cpu.set(C, 0x80 == (cpu.ac & 0x80))
	cpu.ac <<= 1
	if carry {
		cpu.ac++
	}
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opRol() {
	data := cpu.readOpData()
	carry := cpu.isSet(C)
	cpu.set(C, 0x80 == (data & 0x80))
	data <<= 1
	if carry {
		data++
	}
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
	cpu.setZN(data)
}

func (cpu *CPU) opRorAcc() {
	carry := cpu.isSet(C)
	cpu.set(C, 1 == (cpu.ac & 1))
	cpu.ac >>= 1
	if carry {
		cpu.ac += 0x80
	}
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opRor() {
	data := cpu.readOpData()
	carry := cpu.isSet(C)
	cpu.set(C, 1 == (data & 1))
	data >>= 1
	if carry {
		data += 0x80
	}
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
	cpu.setZN(data)
}

func (cpu *CPU) opRra() {
	opData := uint16(cpu.readOpData())
	carry := cpu.isSet(C)
	cpu.set(C, 1 == (opData & 1))
	opData >>= 1
	if carry {
		opData |= 0x80
	}
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, uint8(opData))

	// TODO: This is duplicated logic from opAdc, somehow factor out?
	result := uint16(cpu.ac) + opData
	if cpu.isSet(C) {
		result++
	}

	// V is complicated, see http://forums.nesdev.com/viewtopic.php?t=6331.
	cpu.set(V, 0 != ((uint16(cpu.ac) ^ result) & (opData ^ result) & 0x80))
	cpu.set(C, result > 0xff)
	cpu.ac = uint8(result & 0xff)
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opRti() {
	// There is no actual BRK flag, only exists in the flag when pushed to stack.
	cpu.st = ALWAYS_ON | (cpu.pop() & ^B)
	cpu.pc = cpu.popWord()
}

func (cpu *CPU) opRts() {
	cpu.pc = cpu.popWord()
	cpu.pc++
}

func (cpu *CPU) opSbc() {
	opData := uint16(cpu.readOpData())

	result := uint16(cpu.ac) - opData
	if !cpu.isSet(C) {
		result--
	}

	cpu.set(V, 0 != ((uint16(cpu.ac) ^ result) & (uint16(cpu.ac) ^ opData) & 0x80))
	cpu.set(C, result < 0x100)
	cpu.ac = uint8(result & 0xff)
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opSec() {
	cpu.set(C, true)
}

func (cpu *CPU) opSed() {
	cpu.set(D, true)
}

func (cpu *CPU) opSei() {
	cpu.set(I, true)
}

func (cpu *CPU) opSlo() {
	data := cpu.readOpData()
	cpu.set(C, 0x80 == (data & 0x80))
	data <<= 1
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
	cpu.ac |= data
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opSre() {
	data := cpu.readOpData()
	cpu.set(C, 1 == (data & 1))
	data >>= 1
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, data)
	cpu.ac ^= data
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opSta() {
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, cpu.ac)
}

func (cpu *CPU) opStx() {
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, cpu.xr)
}

func (cpu *CPU) opSty() {
	cpu.clockCycles += cpu.mem.Write(cpu.opAddr, cpu.yr)
}

// The address calculations in SXA and SYA are kind of crazy and based on a discussion from:
// http://forums.nesdev.com/viewtopic.php?f=3&t=10698&sid=87e2c7959251b873fbb89b443f4d50db&start=15
func (cpu *CPU) opSxa() {
	low := cpu.opRawAddr & 0xff
	high := cpu.opRawAddr >> 8

	address := low
	if low + uint16(cpu.yr) > 0xff {
		address = ((high + 1) & uint16(cpu.xr)) << 8 + ((uint16(cpu.yr) + low) & 0xff)
	} else {
		address = (high << 8) + uint16(cpu.yr) + low
	}

	cpu.clockCycles += cpu.mem.Write(address, cpu.xr & uint8(high + 1))
}

func (cpu *CPU) opSya() {
	low := cpu.opRawAddr & 0xff
	high := cpu.opRawAddr >> 8

	address := low
	if low + uint16(cpu.xr) > 0xff {
		address = ((high + 1) & uint16(cpu.yr)) << 8 + ((uint16(cpu.xr) + low) & 0xff)
	} else {
		address = (high << 8) + uint16(cpu.xr) + low
	}

	cpu.clockCycles += cpu.mem.Write(address, cpu.yr & uint8(high + 1))
}

func (cpu *CPU) opTax() {
	cpu.xr = cpu.ac
	cpu.setZN(cpu.xr)
}

func (cpu *CPU) opTay() {
	cpu.yr = cpu.ac
	cpu.setZN(cpu.yr)
}

func (cpu *CPU) opTsx() {
	cpu.xr = cpu.sp
	cpu.setZN(cpu.xr)
}

func (cpu *CPU) opTxa() {
	cpu.ac = cpu.xr
	cpu.setZN(cpu.ac)
}

func (cpu *CPU) opTxs() {
	cpu.sp = cpu.xr
}

func (cpu *CPU) opTya() {
	cpu.ac = cpu.yr
	cpu.setZN(cpu.ac)
}
