package cpu

// This package provides emulation of the 6502 CPU.
// Decimal mode is not implemented as it's not required for the NES.

import (
	"fmt"
)

// The 6502 CPU emulation code only requires an implementation of this interface.
type MemoryInterface interface {
	// Read the byte at 'addr'.
	Read(addr uint16) (val uint8)

	// Write 'val' at 'addr'.  Returns how many extra CPU cycles the write took.
	Write(addr uint16, val uint8) (cycles uint64)
}

type CPU struct {
	// Hardware registers.

	// Accumulator
	ac uint8

	// X Register
	xr uint8

	// Y Register
	yr uint8

	// Status Register
	st uint8

	// Stack pointer Register (stack is actually at 0x100 + sp)
	sp uint8

	// Program counter
	pc uint16

	// What is the literal address provided in the opcode?  Used for debugging and one
	// undocumented and terrible instruction.
	//
	// For example, for the instruction "LDA $2002, X" with XR==5, opRawAddr is $2002
	// and opAddr is $2007.
	opRawAddr uint16

	// What is the fully-resolved memory address to read from?  See example above.
	opAddr uint16

	// Provides memory read and write functionality.
	mem MemoryInterface

	// How many clock cycles were used in execution of the opcode we just interpreted?
	//
	// Used to track how many CPU clock cycles were used in the execution of an instruction.
	// We can't execute part of an instruction, but we track how long each instruction takes
	// in order to synchronize the PPU (which runs at a higher clock) with the CPU.
	clockCycles uint64

	// Set to true to log every instruction to the console.
	Debug bool
}

// Allocate a new CPU and initialize its internal state.
func NewCPU(mem MemoryInterface) (cpu *CPU) {
	cpu = new(CPU)
	cpu.mem = mem

	// This is the initial hardware state, according to a test ROM from nesdev.
	cpu.pc = uint16(cpu.mem.Read(vectorReset))
	cpu.pc |= (uint16(cpu.mem.Read(vectorReset + 1)) << 8)
	cpu.yr = 0
	cpu.xr = 0
	cpu.ac = 0
	cpu.st = 0x34
	cpu.sp = 0xfd

	// Internal initial state.
	cpu.opRawAddr = 0
	cpu.opAddr = 0
	cpu.clockCycles = 0
	return
}

// Reset the CPU.
func (cpu *CPU) Reset() {
	// This is the hardware state upon reset according to a test ROM from nesdev.
	cpu.pc = uint16(cpu.mem.Read(vectorReset))
	cpu.pc |= (uint16(cpu.mem.Read(vectorReset + 1)) << 8)
	cpu.sp |= I
	cpu.sp -= 3

	// Reset internal state.
	cpu.opRawAddr = 0
	cpu.opAddr = 0
	cpu.clockCycles = 0
}

// Hardware vectors for interrupts.  The vectors below are the location of an address
// loaded into the PC.
const (
	// Triggered on power-up.
	vectorReset uint16 = 0xfffc

	// Triggered every VBlank.
	vectorNMI uint16 = 0xfffa

	// Triggered when a 'BRK' instruction is called, or when a hardware IRQ occurs.
	vectorIRQBRK uint16 = 0xfffe
)

func (cpu *CPU) NMI() uint64 {
	if cpu.Debug {
		output := cpu.formatRegisters()
		output += " [NMI]"
		fmt.Println(output)
	}

	cpu.clockCycles = 0
	cpu.pushWord(cpu.pc)
	cpu.push(cpu.st)
	cpu.set(I, true)
	cpu.pc = uint16(cpu.mem.Read(vectorNMI))
	cpu.pc |= (uint16(cpu.mem.Read(vectorNMI + 1)) << 8)
	return 7
}

func (cpu *CPU) Interpret() uint64 {
	cpu.clockCycles = 0

	// Save this for logging.  The PC is incremented as part of execution but we want
	// to log where we're executing from.
	opAddr := cpu.pc

	opcode := cpu.readPC8()
	op := opTable[opcode]

	// Some opcodes are bad.  Perhaps they should trigger a reset?
	if nil == op {
		panic(fmt.Sprintf("no optable entry for opcode 0x%02x, opaddr=0x%4x, regs=%s",
				  opcode, opAddr, cpu.formatRegisters()))
	}

	// There are a handful of addressing modes that each instruction can choose from.
	// The resolution of the final address only depends on the addressing mode, so we
	// tag each opcode with its addressing mode and resolve the address as its own step.
	cpu.readAddressOfOperand(op)

	if cpu.Debug {
		cpu.logOp(opAddr, opcode, op)
	}

	// Execute the op and track how many cycles it took.  We use += instead of = to track
	// the cycles because the address mode resolution may incur extra clock cycles.
	cpu.clockCycles += op.cycles
	op.exec(cpu)

	return cpu.clockCycles
}
