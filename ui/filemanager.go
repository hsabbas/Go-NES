package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type fileManager struct {
	confFile    string
	currentDir  string
	selectedRom int
	nesFiles    []string
	childDirs   []string
}

func (fm *fileManager) initConf() error {
	userConf, err := os.UserConfigDir()
	if err != nil {
		log.Println("Couldn't get user config directory")
		return err
	}

	goNesConf := filepath.Join(userConf, "go-nes")
	info, err := os.Stat(goNesConf)
	if err != nil || !info.IsDir() {
		err = os.Mkdir(goNesConf, os.ModePerm)
		if err != nil {
			return err
		}
	}

	confFile := filepath.Join(goNesConf, "go-nes.txt")
	_, err = os.OpenFile(confFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	fm.confFile = confFile
	return nil
}

func (fm *fileManager) tryDefaultDir() error {
	err := fm.readStoredDir()
	if err != nil {
		log.Println(err)
		path, err := os.Getwd()
		if err == nil {
			return fm.setDir(path)
		}
		return err
	}
	return err
}

func (fm *fileManager) tryUserPath(path string) error {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid path provided")
	}

	return fm.setDir(path)
}

func (fm *fileManager) readStoredDir() error {
	if len(fm.confFile) == 0 {
		return fmt.Errorf("no config file")
	}

	data, err := os.ReadFile(fm.confFile)
	path := string(data)
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return err
	}

	return fm.setDir(path)
}

func (fm *fileManager) setDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	fm.currentDir = path
	fm.childDirs = []string{}
	fm.nesFiles = []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			fm.childDirs = append(fm.childDirs, entry.Name())
		}
		if isNesRom(entry.Name()) {
			fm.nesFiles = append(fm.nesFiles, entry.Name())
		}
	}

	return nil
}

func (fm *fileManager) storeDir() {
	if len(fm.confFile) != 0 {
		os.WriteFile(fm.confFile, []byte(fm.currentDir), 0666)
	}
}

func (fm *fileManager) readRomFile(ind int32) ([]byte, error) {
	path := filepath.Join(fm.currentDir, fm.nesFiles[ind])
	return os.ReadFile(path)
}

func (fm *fileManager) goToChildDir(ind int32) error {
	fullPath := filepath.Join(fm.currentDir, fm.childDirs[ind])
	return fm.setDir(fullPath)
}

func (fm *fileManager) goToParent() error {
	newDir := filepath.Dir(fm.currentDir)
	newDir, err := filepath.Abs(newDir)
	if err != nil {
		return err
	}
	return fm.setDir(newDir)
}

func isNesRom(name string) bool {
	return filepath.Ext(name) == ".nes"
}
