package cpu

import "fmt"

// Format the provided op for logging.  Assumes that the address has already been resolved.
// We don't implement this on the OpcodeEntry because we the address is stored in the CPU state.
func (cpu *CPU) formatOp(op *OpcodeEntry) (ret string){
	ret = op.name + " "

	switch(op.addressing) {
	case BAD:
		ret += "BAD?!"
	case IMP:
		// Nothing to do here, there's no address.
	case IMM:
		ret += fmt.Sprintf("#$%02X", cpu.mem.Read(cpu.opAddr))
	case ZP:
		ret += fmt.Sprintf("$%02X", cpu.opRawAddr)
	case ZPX:
		ret += fmt.Sprintf("$%02X,X", cpu.opRawAddr)
	case ZPY:
		ret += fmt.Sprintf("$%02X,Y", cpu.opRawAddr)
	case ABS:
		ret += fmt.Sprintf("$%04X", cpu.opRawAddr)
	case ABSX:
		ret += fmt.Sprintf("$%04X,X", cpu.opRawAddr)
	case ABSY:
		ret += fmt.Sprintf("$%04X,Y", cpu.opRawAddr)
	case IND:
		ret += fmt.Sprintf("$(%04X)", cpu.opRawAddr)
	case INDY:
		ret += fmt.Sprintf("$(%02X),Y", cpu.opRawAddr)
	case INDX:
		ret += fmt.Sprintf("$(%02X,X)", cpu.opRawAddr)
	}

	return
}

// Format the hardware register state of the CPU for logging, including a verbose expansion
// of the flags.
func (cpu *CPU) formatRegisters() (out string) {
	out = fmt.Sprintf("[A:%02X X:%02X Y:%02X T:%02X SP:%02X [",
			  cpu.ac, cpu.xr, cpu.yr, cpu.st, cpu.sp)

	if cpu.isSet(N) {
		out += "N"
	} else {
		out += "-"
	}

	if cpu.isSet(V) {
		out += "V"
	} else {
		out += "-"
	}

	// This bit in the status flag is always on.
	out += "1"

	if cpu.isSet(B) {
		out += "B"
	} else {
		out += "-"
	}

	if cpu.isSet(D) {
		out += "D"
	} else {
		out += "-"
	}

	if cpu.isSet(I) {
		out += "I"
	} else {
		out += "-"
	}

	if cpu.isSet(Z) {
		out += "Z"
	} else {
		out += "-"
	}

	if cpu.isSet(C) {
		out += "C"
	} else {
		out += "-"
	}

	out += "]"
	return
}

// Log the op 'opcode' (fully described in 'op') executing at address 'opAddr'
func (cpu *CPU) logOp(opAddr uint16, opcode byte, opEntry *OpcodeEntry) {
	// This is probably not the most efficient way to build a string, but if you're logging
	// the execution like this you don't really care about efficiency...
	output := cpu.formatRegisters()
	output += fmt.Sprintf(" [0x%04x] 0x%02x %s", opAddr, opcode, cpu.formatOp(opEntry))
	fmt.Println(output)
}
