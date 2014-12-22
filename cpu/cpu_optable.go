package cpu

// Each opcode has an OpcodeEntry.
type OpcodeEntry struct {
	// User-readable name of the opcode.
	name string

	// The function that actually executes the opcode.
	exec func(cpu *CPU)()

	// How many cycles does the opcode take?
	cycles uint64

	// Are there "extra cycles" that must be accounted for if reads cross page boundaries?
	// Only certain opcodes incur extra cycles during execution.  See cpu_exec.go for
	// which.
	extraCycles uint64

	// What kind of addressing mode does the opcode use?
	addressing int
}

var opTable = map[uint8]*OpcodeEntry {
	// AAC is undocumented.
	0x0B: { "AAC", (*CPU).opAac, 2, 0, IMM },
	0x2B: { "AAC", (*CPU).opAac, 2, 0, IMM },

	// AAX is undocumented.
	0x87: { "AAX", (*CPU).opAax, 3, 0, ZP },
	0x97: { "AAX", (*CPU).opAax, 4, 0, ZPY },
	0x83: { "AAX", (*CPU).opAax, 6, 0, INDX },
	0x8F: { "AAX", (*CPU).opAax, 4, 0, ABS},

	0x69: { "ADC", (*CPU).opAdc, 2, 0, IMM },
	0x65: { "ADC", (*CPU).opAdc, 3, 0, ZP },
	0x75: { "ADC", (*CPU).opAdc, 4, 0, ZPX },
	0x6D: { "ADC", (*CPU).opAdc, 4, 0, ABS },
	0x7D: { "ADC", (*CPU).opAdc, 4, 1, ABSX },
	0x79: { "ADC", (*CPU).opAdc, 4, 1, ABSY },
	0x61: { "ADC", (*CPU).opAdc, 6, 0, INDX },
	0x71: { "ADC", (*CPU).opAdc, 5, 1, INDY },

	0x29: { "AND", (*CPU).opAnd, 2, 0, IMM },
	0x25: { "AND", (*CPU).opAnd, 3, 0, ZP },
	0x35: { "AND", (*CPU).opAnd, 4, 0, ZPX },
	0x2D: { "AND", (*CPU).opAnd, 4, 0, ABS },
	0x3D: { "AND", (*CPU).opAnd, 4, 1, ABSX },
	0x39: { "AND", (*CPU).opAnd, 4, 1, ABSY },
	0x21: { "AND", (*CPU).opAnd, 6, 0, INDX },
	0x31: { "AND", (*CPU).opAnd, 5, 1, INDY },

	0x0A: { "ASL", (*CPU).opAslAcc, 2, 0, IMP },
	0x06: { "ASL", (*CPU).opAsl, 5, 0, ZP },
	0x16: { "ASL", (*CPU).opAsl, 6, 0, ZPX },
	0x0E: { "ASL", (*CPU).opAsl, 6, 0, ABS },
	0x1E: { "ASL", (*CPU).opAsl, 7, 0, ABSX },

	0x4B: { "ASR", (*CPU).opAsr, 2, 0, IMM },  // undocumented

	0x6B: { "ARR", (*CPU).opArr, 2, 0, IMM },  // undocumented

	0xAB: { "ATX", (*CPU).opAtx, 2, 0, IMM },  // undocumented

	0xCB: { "AXS", (*CPU).opAxs, 2, 0, IMM },  // undocumented

	0x90: { "BCC", (*CPU).opBcc, 2, 0, IMM },
	0xB0: { "BCS", (*CPU).opBcs, 2, 0, IMM },
	0xF0: { "BEQ", (*CPU).opBeq, 2, 0, IMM },

	0x24: { "BIT", (*CPU).opBit, 3, 0, ZP },
	0x2C: { "BIT", (*CPU).opBit, 4, 0, ABS },

	0x30: { "BMI", (*CPU).opBmi, 2, 0, IMM },
	0xD0: { "BNE", (*CPU).opBne, 2, 0, IMM },
	0x10: { "BPL", (*CPU).opBpl, 2, 0, IMM },

	0x00: { "BRK", (*CPU).opBrk, 7, 0, IMP },

	0x50: { "BVC", (*CPU).opBvc, 2, 0, IMM },
	0x70: { "BVS", (*CPU).opBvs, 2, 0, IMM },

	0x18: { "CLC", (*CPU).opClc, 2, 0, IMP },
	0xD8: { "CLD", (*CPU).opCld, 2, 0, IMP },
	0x58: { "CLI", (*CPU).opCli, 2, 0, IMP },
	0xB8: { "CLV", (*CPU).opClv, 2, 0, IMP },

	0xC9: { "CMP", (*CPU).opCmp, 2, 0, IMM },
	0xC5: { "CMP", (*CPU).opCmp, 3, 0, ZP },
	0xD5: { "CMP", (*CPU).opCmp, 4, 0, ZPX },
	0xCD: { "CMP", (*CPU).opCmp, 4, 0, ABS },
	0xDD: { "CMP", (*CPU).opCmp, 4, 1, ABSX },
	0xD9: { "CMP", (*CPU).opCmp, 4, 1, ABSY },
	0xC1: { "CMP", (*CPU).opCmp, 6, 0, INDX },
	0xD1: { "CMP", (*CPU).opCmp, 5, 1, INDY },

	0xE0: { "CPX", (*CPU).opCpx, 2, 0, IMM },
	0xE4: { "CPX", (*CPU).opCpx, 3, 0, ZP },
	0xEC: { "CPX", (*CPU).opCpx, 4, 0, ABS },

	0xC0: { "CPY", (*CPU).opCpy, 2, 0, IMM },
	0xC4: { "CPY", (*CPU).opCpy, 3, 0, ZP },
	0xCC: { "CPY", (*CPU).opCpy, 4, 0, ABS },

	// All DCP opcodes are undocumented.
	0xC7: { "DCP", (*CPU).opDcp, 5, 0, ZP },
	0xD7: { "DCP", (*CPU).opDcp, 6, 0, ZPX },
	0xCF: { "DCP", (*CPU).opDcp, 6, 0, ABS },
	0xDF: { "DCP", (*CPU).opDcp, 7, 0, ABSX },
	0xDB: { "DCP", (*CPU).opDcp, 7, 0, ABSY },
	0xC3: { "DCP", (*CPU).opDcp, 8, 0, INDX },
	0xD3: { "DCP", (*CPU).opDcp, 8, 0, INDY },

	0xC6: { "DEC", (*CPU).opDec, 5, 0, ZP },
	0xD6: { "DEC", (*CPU).opDec, 6, 0, ZPX },
	0xCE: { "DEC", (*CPU).opDec, 6, 0, ABS },
	0xDE: { "DEC", (*CPU).opDec, 7, 0, ABSX },

	0xCA: { "DEX", (*CPU).opDex, 2, 0, IMP },

	0x88: { "DEY", (*CPU).opDey, 2, 0, IMP },

	// All DOP ops are undocumented.
	0x04: { "DOP", (*CPU).opNop, 3, 0, ZP },
	0x14: { "DOP", (*CPU).opNop, 4, 0, ZPX },
	0x34: { "DOP", (*CPU).opNop, 4, 0, ZPX },
	0x44: { "DOP", (*CPU).opNop, 3, 0, ZP },
	0x54: { "DOP", (*CPU).opNop, 4, 0, ZPX },
	0x64: { "DOP", (*CPU).opNop, 3, 0, ZP },
	0x74: { "DOP", (*CPU).opNop, 4, 0, ZPX },
	0x80: { "DOP", (*CPU).opNop, 2, 0, IMM },
	0x82: { "DOP", (*CPU).opNop, 2, 0, IMM },
	0x89: { "DOP", (*CPU).opNop, 2, 0, IMM },
	0xC2: { "DOP", (*CPU).opNop, 2, 0, IMM },
	0xD4: { "DOP", (*CPU).opNop, 4, 0, ZPX },
	0xE2: { "DOP", (*CPU).opNop, 2, 0, IMM },
	0xF4: { "DOP", (*CPU).opNop, 4, 0, ZPX },

	0x49: { "EOR", (*CPU).opEor, 2, 0, IMM },
	0x45: { "EOR", (*CPU).opEor, 3, 0, ZP },
	0x55: { "EOR", (*CPU).opEor, 4, 0, ZPX },
	0x4D: { "EOR", (*CPU).opEor, 4, 0, ABS },
	0x5D: { "EOR", (*CPU).opEor, 4, 1, ABSX },
	0x59: { "EOR", (*CPU).opEor, 4, 1, ABSY },
	0x41: { "EOR", (*CPU).opEor, 6, 0, INDX },
	0x51: { "EOR", (*CPU).opEor, 5, 1, INDY },

	0xE6: { "INC", (*CPU).opInc, 5, 0, ZP },
	0xF6: { "INC", (*CPU).opInc, 6, 0, ZPX },
	0xEE: { "INC", (*CPU).opInc, 6, 0, ABS },
	0xFE: { "INC", (*CPU).opInc, 7, 0, ABSX },

	0xE8: { "INX", (*CPU).opInx, 2, 0, IMP },
	0xC8: { "INY", (*CPU).opIny, 2, 0, IMP },

	// All ISC opcodes are undocumented.
	0xE7: { "ISC", (*CPU).opIsc, 5, 0, ZP },
	0xF7: { "ISC", (*CPU).opIsc, 6, 0, ZPX },
	0xEF: { "ISC", (*CPU).opIsc, 6, 0, ABS },
	0xFF: { "ISC", (*CPU).opIsc, 7, 0, ABSX },
	0xFB: { "ISC", (*CPU).opIsc, 7, 0, ABSY },
	0xE3: { "ISC", (*CPU).opIsc, 8, 0, INDX },
	0xF3: { "ISC", (*CPU).opIsc, 8, 0, INDY },

	0x4C: { "JMP", (*CPU).opJmp, 3, 0, ABS },
	0x6C: { "JMP", (*CPU).opJmp, 5, 0, IND },
	0x20: { "JSR", (*CPU).opJsr, 6, 0, ABS },

	// All LAX opcodes are undefined.
	0xA7: { "LAX", (*CPU).opLax, 3, 0, ZP },
	0xB7: { "LAX", (*CPU).opLax, 4, 0, ZPY },
	0xAF: { "LAX", (*CPU).opLax, 4, 0, ABS },
	0xBF: { "LAX", (*CPU).opLax, 4, 1, ABSY },
	0xA3: { "LAX", (*CPU).opLax, 6, 0, INDX },
	0xB3: { "LAX", (*CPU).opLax, 5, 1, INDY},

	0xA9: { "LDA", (*CPU).opLda, 2, 0, IMM },
	0xA5: { "LDA", (*CPU).opLda, 3, 0, ZP },
	0xB5: { "LDA", (*CPU).opLda, 4, 0, ZPX },
	0xAD: { "LDA", (*CPU).opLda, 4, 0, ABS },
	0xBD: { "LDA", (*CPU).opLda, 4, 1, ABSX },
	0xB9: { "LDA", (*CPU).opLda, 4, 1, ABSY },
	0xA1: { "LDA", (*CPU).opLda, 6, 0, INDX },
	0xB1: { "LDA", (*CPU).opLda, 5, 1, INDY },

	0xA2: { "LDX", (*CPU).opLdx, 2, 0, IMM },
	0xA6: { "LDX", (*CPU).opLdx, 3, 0, ZP },
	0xB6: { "LDX", (*CPU).opLdx, 4, 0, ZPY },
	0xAE: { "LDX", (*CPU).opLdx, 4, 0, ABS },
	0xBE: { "LDX", (*CPU).opLdx, 4, 1, ABSY },

	0xA0: { "LDY", (*CPU).opLdy, 2, 0, IMM },
	0xA4: { "LDY", (*CPU).opLdy, 3, 0, ZP },
	0xB4: { "LDY", (*CPU).opLdy, 4, 0, ZPX },
	0xAC: { "LDY", (*CPU).opLdy, 4, 0, ABS },
	0xBC: { "LDY", (*CPU).opLdy, 4, 1, ABSX },

	0x4A: { "LSR", (*CPU).opLsrAcc, 2, 0, IMP },
	0x46: { "LSR", (*CPU).opLsr, 5, 0, ZP },
	0x56: { "LSR", (*CPU).opLsr, 6, 0, ZPX },
	0x4E: { "LSR", (*CPU).opLsr, 6, 0, ABS },
	0x5E: { "LSR", (*CPU).opLsr, 7, 0, ABSX },

	0xEA: { "NOP", (*CPU).opNop, 2, 0, IMP },
	0x1A: { "NOP", (*CPU).opNop, 2, 0, IMP },  // undocumented
	0x3A: { "NOP", (*CPU).opNop, 2, 0, IMP },  // undocumented
	0x5A: { "NOP", (*CPU).opNop, 2, 0, IMP },  // undocumented
	0x7A: { "NOP", (*CPU).opNop, 2, 0, IMP },  // undocumented
	0xDA: { "NOP", (*CPU).opNop, 2, 0, IMP },  // undocumented
	0xFA: { "NOP", (*CPU).opNop, 2, 0, IMP },  // undocumented

	0x09: { "ORA", (*CPU).opOra, 2, 0, IMM },
	0x05: { "ORA", (*CPU).opOra, 3, 0, ZP },
	0x15: { "ORA", (*CPU).opOra, 4, 0, ZPX },
	0x0D: { "ORA", (*CPU).opOra, 4, 0, ABS },
	0x1D: { "ORA", (*CPU).opOra, 4, 1, ABSX },
	0x19: { "ORA", (*CPU).opOra, 4, 1, ABSY },
	0x01: { "ORA", (*CPU).opOra, 6, 0, INDX },
	0x11: { "ORA", (*CPU).opOra, 5, 1, INDY },

	0x48: { "PHA", (*CPU).opPha, 3, 0, IMP },
	0x08: { "PHP", (*CPU).opPhp, 3, 0, IMP },
	0x68: { "PLA", (*CPU).opPla, 4, 0, IMP },
	0x28: { "PLP", (*CPU).opPlp, 4, 0, IMP },

	// All RLA are undocumented.
	0x27: { "RLA", (*CPU).opRla, 5, 0, ZP },
	0x37: { "RLA", (*CPU).opRla, 6, 0, ZPX },
	0x2F: { "RLA", (*CPU).opRla, 6, 0, ABS },
	0x3F: { "RLA", (*CPU).opRla, 7, 0, ABSX },
	0x3B: { "RLA", (*CPU).opRla, 7, 0, ABSY },
	0x23: { "RLA", (*CPU).opRla, 8, 0, INDX },
	0x33: { "RLA", (*CPU).opRla, 8, 0, INDY },

	0x2A: { "ROL", (*CPU).opRolAcc, 2, 0, IMP },
	0x26: { "ROL", (*CPU).opRol, 5, 0, ZP },
	0x36: { "ROL", (*CPU).opRol, 6, 0, ZPX },
	0x2E: { "ROL", (*CPU).opRol, 6, 0, ABS },
	0x3E: { "ROL", (*CPU).opRol, 7, 0, ABSX },

	0x6A: { "ROR", (*CPU).opRorAcc, 2, 0, IMP },
	0x66: { "ROR", (*CPU).opRor, 5, 0, ZP },
	0x76: { "ROR", (*CPU).opRor, 6, 0, ZPX },
	0x6E: { "ROR", (*CPU).opRor, 6, 0, ABS },
	0x7E: { "ROR", (*CPU).opRor, 7, 0, ABSX },

	// All RRA opcodes are undocumented.
	0x67: { "RRA", (*CPU).opRra, 5, 0, ZP },
	0x77: { "RRA", (*CPU).opRra, 6, 0, ZPX },
	0x6F: { "RRA", (*CPU).opRra, 6, 0, ABS },
	0x7F: { "RRA", (*CPU).opRra, 7, 0, ABSX },
	0x7B: { "RRA", (*CPU).opRra, 7, 0, ABSY },
	0x63: { "RRA", (*CPU).opRra, 8, 0, INDX },
	0x73: { "RRA", (*CPU).opRra, 8, 0, INDY },

	0x40: { "RTI", (*CPU).opRti, 6, 0, IMP },
	0x60: { "RTS", (*CPU).opRts, 6, 0, IMP },

	0xE9: { "SBC", (*CPU).opSbc, 2, 0, IMM },
	0xEB: { "SBC", (*CPU).opSbc, 2, 0, IMM },   // undocumented, actually same as 0xE9.
	0xE5: { "SBC", (*CPU).opSbc, 3, 0, ZP },
	0xF5: { "SBC", (*CPU).opSbc, 4, 0, ZPX },
	0xED: { "SBC", (*CPU).opSbc, 4, 0, ABS },
	0xFD: { "SBC", (*CPU).opSbc, 4, 1, ABSX },
	0xF9: { "SBC", (*CPU).opSbc, 4, 1, ABSY },
	0xE1: { "SBC", (*CPU).opSbc, 6, 0, INDX },
	0xF1: { "SBC", (*CPU).opSbc, 5, 1, INDY },

	0x38: { "SEC", (*CPU).opSec, 2, 0, IMP },
	0xF8: { "SED", (*CPU).opSed, 2, 0, IMP },
	0x78: { "SEI", (*CPU).opSei, 2, 0, IMP },

	// All SLO ops are undocumented.
	0x07: { "SLO", (*CPU).opSlo, 5, 0, ZP },
	0x17: { "SLO", (*CPU).opSlo, 6, 0, ZPX },
	0x0F: { "SLO", (*CPU).opSlo, 6, 0, ABS },
	0x1F: { "SLO", (*CPU).opSlo, 7, 0, ABSX },
	0x1B: { "SLO", (*CPU).opSlo, 7, 0, ABSY },
	0x03: { "SLO", (*CPU).opSlo, 8, 0, INDX },
	0x13: { "SLO", (*CPU).opSlo, 8, 0, INDY },

	// All SRE opcodes are undocumented.
	0x47: { "SRE", (*CPU).opSre, 5, 0, ZP },
	0x57: { "SRE", (*CPU).opSre, 6, 0, ZPX },
	0x4F: { "SRE", (*CPU).opSre, 6, 0, ABS },
	0x5F: { "SRE", (*CPU).opSre, 7, 0, ABSX },
	0x5B: { "SRE", (*CPU).opSre, 7, 0, ABSY },
	0x43: { "SRE", (*CPU).opSre, 8, 0, INDX },
	0x53: { "SRE", (*CPU).opSre, 8, 0, INDY },

	0x85: { "STA", (*CPU).opSta, 3, 0, ZP },
	0x95: { "STA", (*CPU).opSta, 4, 0, ZPX },
	0x8D: { "STA", (*CPU).opSta, 4, 0, ABS },
	0x9D: { "STA", (*CPU).opSta, 5, 0, ABSX },
	0x99: { "STA", (*CPU).opSta, 5, 0, ABSY },
	0x81: { "STA", (*CPU).opSta, 6, 0, INDX },
	0x91: { "STA", (*CPU).opSta, 6, 0, INDY },

	0x86: { "STX", (*CPU).opStx, 3, 0, ZP },
	0x96: { "STX", (*CPU).opStx, 4, 0, ZPY },
	0x8E: { "STX", (*CPU).opStx, 4, 0, ABS },

	0x84: { "STY", (*CPU).opSty, 3, 0, ZP },
	0x94: { "STY", (*CPU).opSty, 4, 0, ZPX },
	0x8C: { "STY", (*CPU).opSty, 4, 0, ABS },

	0x9E: { "SXA", (*CPU).opSxa, 5, 0, ABSY },  // undocumented
	0x9C: { "SYA", (*CPU).opSya, 5, 0, ABSX },  // undocumented

	0xAA: { "TAX", (*CPU).opTax, 2, 0, IMP },
	0xA8: { "TAY", (*CPU).opTay, 2, 0, IMP },

	// All TOP opcodes are undocumented.
	0x0C: { "TOP", (*CPU).opNop, 4, 0, ABS },
	0x1C: { "TOP", (*CPU).opNop, 4, 1, ABSX },
	0x3C: { "TOP", (*CPU).opNop, 4, 1, ABSX },
	0x5C: { "TOP", (*CPU).opNop, 4, 1, ABSX },
	0x7C: { "TOP", (*CPU).opNop, 4, 1, ABSX },
	0xDC: { "TOP", (*CPU).opNop, 4, 1, ABSX },
	0xFC: { "TOP", (*CPU).opNop, 4, 1, ABSX },

	0xBA: { "TSX", (*CPU).opTsx, 2, 0, IMP },
	0x8A: { "TXA", (*CPU).opTxa, 2, 0, IMP },
	0x9A: { "TXS", (*CPU).opTxs, 2, 0, IMP },
	0x98: { "TYA", (*CPU).opTya, 2, 0, IMP },
}
