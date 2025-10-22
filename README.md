# NES Emulator
## Version Make-it-Work (Pre-Alpha)

![Playing Mario](https://github.com/user-attachments/assets/70cfe606-6151-422f-ac77-c2cfcbcea3a1)

### And it does work! At least a little bit.
Features implemented:
- CPU with all official opcodes
- PPU with Graphics using OpenGL
- Controls with two different keyboard layouts for input
- NROM mapper

Currently doesn't support two players or audio.

#### To run the emulator:   
Clone the repository  
Enable CGo to run OpenGL:  
```
go env -w "CGO_ENABLED=1"
```
  
Now it should run:  
```
go run main.go  
```

Currently runs with the first .nes file it finds in the current directory.

### Controls
Includes layout for arrow keys as well as WASD controls.

| NES | layout 1 | layout 2 |
| --- | --- | --- |
| Directions | Arrow Keys | WASD |
| A | X | . |
| B | Z | , |
| Select | Shift | Shift |
| Start | Enter | Enter |
