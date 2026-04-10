# Go NES Emulator
## Version Make-it-Better (Beta)

<img width="256" height="240" alt="Screenshot 2026-03-06 150332" src="https://github.com/user-attachments/assets/69bae8ca-39a2-4fd3-b406-727e7771eaf3" /><img width="256" height="240" alt="Screenshot 2026-03-06 150513" src="https://github.com/user-attachments/assets/fe82da95-cc23-4cd7-a4da-55257f600690" />



### Features implemented:
- CPU with all official opcodes
- PPU with Graphics using raylib
- Two different keyboard layouts as well as controller support
- Two players

### Cartridges Supported
- Mapper 0 NROM
- Mapper 1 MMC1
- Mapper 2 UxROM
- Mapper 3 CNROM
- Mapper 7 AxROM

## Roadmap for Beta release
- Controller support ✓
- Two players ✓
- NES APU implementation and audio
- MMC3 mapper

## To run the emulator:   
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

## Controls

| NES | Keyboard 1 | Keyboard 2 | Controller (XBox names) |
| --- | --- | --- | --- |
| Directions | Arrow Keys | WASD | D-pad |
| A | X | . | A |
| B | Z | , | X |
| Select | Shift | Shift | Select |
| Start | Enter | Enter | Start |
