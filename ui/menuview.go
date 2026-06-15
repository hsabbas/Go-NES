package ui

import (
	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type menuview struct {
	fm              *fileManager
	errMsgs         []string
	startGame       func([]byte) error
	selectedSection int

	listViewStr       string
	listViewScroll    int32
	listViewActive    int32
	pathBoxText       string
	pathEditMode      bool
	searchPressed     bool
	openPressed       bool
	startPressed      bool
	goToParentPressed bool
}

func createMenuView(romPath string, startGame func([]byte) error) *menuview {
	errMsgs := make([]string, 0)
	fm := &fileManager{}
	err := fm.initConf()
	if err != nil {
		errMsgs = append(errMsgs, "failed to load config file")
	}

	ok := false
	if len(romPath) != 0 {
		err := fm.tryUserPath(romPath)
		if err != nil {
			errMsgs = append(errMsgs, "invalid path argument provided")
		} else {
			ok = true
		}
	}

	if !ok {
		err = fm.tryDefaultDir()
		if err != nil {
			errMsgs = append(errMsgs, "failed to load default folder")
		}
	}

	mv := &menuview{
		startGame:   startGame,
		fm:          fm,
		errMsgs:     errMsgs,
		pathBoxText: fm.currentDir,
	}
	mv.updateUI()
	return mv
}

func (mv *menuview) updateUI() {
	mv.listViewStr = ""
	for i, dir := range mv.fm.childDirs {
		if i != 0 {
			mv.listViewStr += ";"
		}
		mv.listViewStr += "#217#" + dir
	}
	for _, rom := range mv.fm.nesFiles {
		if len(mv.listViewStr) != 0 {
			mv.listViewStr += ";"
		}
		mv.listViewStr += "#002#" + rom
	}

	mv.pathBoxText = mv.fm.currentDir
}

func (mv *menuview) processInput() {
	if mv.searchPressed {
		mv.errMsgs = []string{}
		err := mv.fm.setDir(mv.pathBoxText)
		if err != nil {
			mv.errMsgs = append(mv.errMsgs, "failed to open folder")
			return
		}
		mv.updateUI()
	}

	if mv.openPressed {
		mv.errMsgs = []string{}
		if mv.listViewActive == -1 || mv.listViewActive >= int32(len(mv.fm.childDirs)) {
			return
		}
		err := mv.fm.goToChildDir(mv.listViewActive)
		if err != nil {
			mv.errMsgs = append(mv.errMsgs, "failed to open folder")
			return
		}
		mv.updateUI()
	}

	if mv.startPressed {
		mv.errMsgs = []string{}
		dirCount := int32(len(mv.fm.childDirs))
		if mv.listViewActive == -1 || mv.listViewActive < dirCount {
			return
		}

		rom, err := mv.fm.readRomFile(mv.listViewActive - dirCount)
		if err != nil {
			mv.errMsgs = append(mv.errMsgs, "failed to read ROM file")
			return
		}

		err = mv.startGame(rom)
		if err != nil {
			mv.errMsgs = append(mv.errMsgs, err.Error())
		} else {
			mv.fm.storeDir()
		}
	}

	if mv.goToParentPressed {
		err := mv.fm.goToParent()
		if err != nil {
			mv.errMsgs = append(mv.errMsgs, "error occurred going to parent folder")
		}
		mv.updateUI()
	}
}

func (mv *menuview) update() {}

func (mv *menuview) render() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)
	if gui.TextBox(rl.Rectangle{X: 48, Y: 48, Width: 512, Height: 24}, &mv.pathBoxText, 128, mv.pathEditMode) {
		mv.pathEditMode = !mv.pathEditMode
	}
	mv.searchPressed = gui.Button(
		rl.Rectangle{X: 592, Y: 48, Width: 64, Height: 24},
		"Search",
	)
	mv.openPressed = gui.Button(
		rl.Rectangle{X: 592, Y: 96, Width: 128, Height: 24},
		"Open Folder",
	)
	mv.startPressed = gui.Button(
		rl.Rectangle{X: 592, Y: 146, Width: 128, Height: 24},
		"Start Game",
	)
	mv.goToParentPressed = gui.Button(
		rl.Rectangle{X: 48, Y: 96, Width: 512, Height: 16},
		"#121#",
	)
	gui.ListView(
		rl.Rectangle{X: 48, Y: 112, Width: 512, Height: 512},
		mv.listViewStr,
		&mv.listViewScroll,
		&mv.listViewActive,
	)

	y := int32(194)
	for _, msg := range mv.errMsgs {
		rl.DrawText(msg, 592, y, 12, rl.Red)
		y += 48
	}

	rl.EndDrawing()
}

func (mv *menuview) close() {
}
