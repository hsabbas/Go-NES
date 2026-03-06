# Go NES Emulator
## Version Make-it-Work-More (Alpha)

![Playing Mario](https://github.com/user-attachments/assets/70cfe606-6151-422f-ac77-c2cfcbcea3a1)

### And it does work! Now with more games supported!
Features implemented:
- CPU with all official opcodes
- PPU with Graphics using raylib
- Controls with two different keyboard layouts for input

### Cartridges Supported
- Mapper 0 NROM
- Mapper 1 MMC1
- Mapper 2 UxROM
- Mapper 3 CNROM
- Mapper 7 AxROM

Currently doesn't support two players or audio.

#### To run the emulator:   
Clone the repository  
Enable CGo to build with raylib:  
```
go env -w "CGO_ENABLED=1"
```
  
Run with path to the game ROM:  
```
go run main.go ./path/to/rom.nes  
```

If no path is provied, it runs with the first .nes file it finds in the current directory.

### Controls
Includes layout for arrow keys as well as WASD controls.

| NES | layout 1 | layout 2 |
| --- | --- | --- |
| Directions | Arrow Keys | WASD |
| A | X | . |
| B | Z | , |
| Select | Shift | Shift |
| Start | Enter | Enter |
