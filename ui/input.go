package ui

import rl "github.com/gen2brain/raylib-go/raylib"

func player1PressingUp() bool {
	return rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) || rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceUp)
}

func player1PressingDown() bool {
	return rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) || rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceDown)
}

func player1PressingLeft() bool {
	return rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) || rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceLeft)
}

func player1PressingRight() bool {
	return rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) || rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceRight)
}

func player1PressingA() bool {
	return rl.IsKeyDown(rl.KeyPeriod) || rl.IsKeyDown(rl.KeyX) || rl.IsGamepadButtonDown(0, rl.GamepadButtonRightFaceDown)
}

func player1PressingB() bool {
	return rl.IsKeyDown(rl.KeyComma) || rl.IsKeyDown(rl.KeyZ) || rl.IsGamepadButtonDown(0, rl.GamepadButtonRightFaceLeft)
}

func player1PressingSelect() bool {
	return rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) || rl.IsGamepadButtonDown(0, rl.GamepadButtonMiddleLeft)
}

func player1PressingStart() bool {
	return rl.IsKeyDown(rl.KeyEnter) || rl.IsGamepadButtonDown(0, rl.GamepadButtonMiddleRight)
}

func player2PressingUp() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonLeftFaceUp)
}

func player2PressingDown() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonLeftFaceDown)
}

func player2PressingLeft() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonLeftFaceLeft)
}

func player2PressingRight() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonLeftFaceRight)
}

func player2PressingA() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonRightFaceDown)
}

func player2PressingB() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonRightFaceLeft)
}

func player2PressingSelect() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonMiddleLeft)
}

func player2PressingStart() bool {
	return rl.IsGamepadButtonDown(1, rl.GamepadButtonMiddleRight)
}
