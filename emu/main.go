package main

import (
	// Things from Go.
	"fmt"
	"os"
	"time"

	// Things from Me.
	"cpu"
	"mapper"
	"nesfile"
	"ppu"
	"wrapper"
)

const (
	// There is a master clock that is divided differently for the PPU and CPU.
	// The PPU is clocked faster than the CPU so we measure time in PPU clocks.
	PPUCyclesPerScanLine = 341
	PPUCyclesPerCPUCycle = 3
)

func main() {
	if len(os.Args)  < 2 {
		fmt.Println("Usage: ", os.Args[0], " somefile.nes <debug>")
		return
	}

	wrapper.Init()

	// Read the iNES formatted file.  Dies if errors encountered.
	nesFile := nesfile.ReadNesFile(os.Args[1])

	// The mapper is the on-cart address mapping logic.
	nesMapper := mapper.GetMapper(nesFile)

	// Open the window.  The PPU will draw into this.
	mainWindow := wrapper.NewWindow(ppu.DisplayHeight, ppu.DisplayWidth, "hello world")

	// Processes graphical data and renders it into the window opened above.
	nesPpu := ppu.NewPPU(nesMapper, mainWindow)

	// Polls keyboard events and provides key press data.
	input := wrapper.NewInputProvider()

	// Implements the bus on the CPU.
	nesMemory := NewNESMemory(nesPpu, nesMapper, input)

	// Interprets and executes the opcodes.
	nesCpu := cpu.NewCPU(nesMemory)

	// If there are any trailing arguments turn on debugging.
	if len(os.Args) > 2 {
		nesCpu.Debug = true
		nesPpu.Debug = true
		nesMapper.Debug(true)
	}

	// Main render loop
	for {
		if input.IsKeyPressed(wrapper.KEY_QUIT) {
			break
		} else if input.IsKeyPressed(wrapper.KEY_RESET) {
			nesCpu.Reset()
		}

		var cycles uint64 = 0

		for scanline := 0; scanline < 262; scanline++ {
			// Lines 0-239 are visible.
			if scanline < 240 {
				// We might execute too many cycles here...
				for cycles < PPUCyclesPerScanLine {
					cycles += PPUCyclesPerCPUCycle * nesCpu.Interpret()
				}
				nesPpu.RenderScanLine()
				time.Sleep(10000)
				// But we'll execute them later by doing a subtraction here instead
				// of resetting cycles to 0.
				cycles -= PPUCyclesPerScanLine
			} else if scanline == 240 {
				// Line 240 is the "post-render line"
				for cycles < PPUCyclesPerScanLine {
					cycles += PPUCyclesPerCPUCycle * nesCpu.Interpret()
				}
				cycles -= PPUCyclesPerScanLine
			} else if scanline == 241 {
				// VBlank is entered on line 241 (or at least the flag is set)
				// Blit what we've rendered to the window.
				mainWindow.Blit()

				// Notify the PPU that we're in VBlank, and see if we should tell
				// the CPU to execute an NMI.
				if nesPpu.EnterVBlankShouldNMI() {
					cycles += PPUCyclesPerCPUCycle * nesCpu.NMI()
				}
				for cycles < PPUCyclesPerScanLine {
					cycles += PPUCyclesPerCPUCycle * nesCpu.Interpret()
				}
				cycles -= PPUCyclesPerScanLine
			} else if scanline < 261 {
				// Lines 242 -> 260 don't render or set flags.
				for cycles < PPUCyclesPerScanLine {
					// Turn on debugging during vblank for now.
					cycles += PPUCyclesPerCPUCycle * nesCpu.Interpret()
				}
				cycles -= PPUCyclesPerScanLine
			} else {
				// The "pre-render" line clears the VBlank flag early.
				nesPpu.ExitVBlank()
				for cycles < PPUCyclesPerScanLine {
					// Turn on debugging during vblank for now.
					cycles += PPUCyclesPerCPUCycle * nesCpu.Interpret()
				}
				cycles -= PPUCyclesPerScanLine
			}
		}
	}
}
