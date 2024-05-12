package windowController

import (
	"sync"
	"syscall"
)

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	user32         = syscall.NewLazyDLL("user32.dll")
	procShowWindow = user32.NewProc("ShowWindow")
	once           sync.Once
	instance       *WindowController
)

type WindowController struct {
	console uintptr
	Status  int
}

func GetWindowController() *WindowController {
	once.Do(func() {
		instance = newController()
	})
	return instance
}

func newController() *WindowController {
	procGetConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	console, _, _ := procGetConsoleWindow.Call()

	return &WindowController{console: console, Status: 1}
}

func (c *WindowController) HideWindow() {
	c.Status = 0
	_, _, _ = procShowWindow.Call(c.console, uintptr(0)) // SW_HIDE = 0
}

func (c *WindowController) ShowWindow() {
	c.Status = 1
	_, _, _ = procShowWindow.Call(c.console, uintptr(1)) // SW_SHOW = 1
}

func (c *WindowController) ToggleWindow() {
	c.Status ^= 1
	_, _, _ = procShowWindow.Call(c.console, uintptr(c.Status))
}
