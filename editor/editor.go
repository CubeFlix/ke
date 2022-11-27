// editor/editor.go
// Package editor provides the editor interface

package editor

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/cubeflix/edit/buffer"
	"github.com/gdamore/tcell"
)

const (
	MaxBufferSize = 1e5
	MaxLineSize   = 1e5
)

// Editor struct.
type Editor struct {
	// Terminal screen object.
	screen tcell.Screen

	// The file to use.
	file string

	// The size of the editor window.
	width  int
	height int

	// The current x and y position of the cursor. The location of the viewport
	// relative to the buffer is calculated by ensuring that the current line
	// is in the center of the screen and the current position on the line is
	// always within view.
	cursorX int
	cursorY int
	top     int
	left    int

	// Are we currently running.
	running bool

	// File buffer.
	buffer *buffer.Buffer
}

// Create a new editor.
func NewEditor(file string) (*Editor, error) {
	// Create the new tcell screen.
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	// Return.
	return &Editor{
		screen: screen,
		file:   file,
		buffer: buffer.NewBuffer(MaxBufferSize, MaxLineSize),
	}, nil
}

// Initialize the editor.
func (e *Editor) Init() error {
	// Initialize the screen.
	if err := e.screen.Init(); err != nil {
		return err
	}

	// Check if the file exists.
	_, err := os.Stat(e.file)
	if os.IsNotExist(err) {
		// File does not exist.
		line := buffer.NewBufferLine(MaxLineSize)
		e.buffer.SetData([]*buffer.BufferLine{line})
	} else {
		// Read the file.
		file, err := os.Open(e.file)
		if err != nil {
			return err
		}
		defer file.Close()

		// Read the file into the buffer.
		scanner := bufio.NewReader(file)
		lines := make([]*buffer.BufferLine, 0)
		for {
			str, err := scanner.ReadString('\n')
			fmt.Println(str, err)
			if err != nil {
				if err == io.EOF {
					// EOF, break.
					line := buffer.NewBufferLine(MaxLineSize)
					line.Insert([]rune(str), 0)
					lines = append(lines, line)
					break
				}
				return err
			}
			line := buffer.NewBufferLine(MaxLineSize)
			line.Insert([]rune(str[:len(str)-2]), 0)
			lines = append(lines, line)
		}
		fmt.Println(lines)

		// Set the buffer lines.
		e.buffer.SetData(lines)
	}

	// Get the screen size.
	e.width, e.height = e.screen.Size()

	e.screen.Clear()

	// Return.
	return nil
}

// Handle user input events.
func (e *Editor) HandleEvents() {
	e.running = true
	e.Render()
	for e.running {
		event := e.screen.PollEvent()

		switch event := event.(type) {
		case *tcell.EventKey:
			e.handleKeyPress(event)
		case *tcell.EventResize:
			e.width, e.height = event.Size()
			e.Render()
		}
	}
}

// Handle a key press.
func (e *Editor) handleKeyPress(event *tcell.EventKey) (render bool) {
	if event.Key() == tcell.KeyEscape {
		// Exit.
		e.Exit()
		return false
	}
	defer func() {
		if render {
			e.Render()
		}
	}()
	if event.Key() == tcell.KeyDown {
		// Down.
		if e.cursorY >= e.buffer.Size()-1 {
			e.screen.Beep()
			return true
		}
		e.cursorY += 1

		// If the next line is too short, move the cursor X.
		newSize := e.buffer.Data()[e.cursorY].Size()
		if e.cursorX >= newSize {
			e.cursorX = newSize
			if e.cursorX < e.left {
				// Out of viewing area.
				e.left = e.cursorX
			}
		}
	} else if event.Key() == tcell.KeyUp {
		// Up.
		if e.cursorY == 0 {
			e.screen.Beep()
			return true
		}
		e.cursorY -= 1

		// If the next line is too short, move the cursor X.
		newSize := e.buffer.Data()[e.cursorY].Size()
		if e.cursorX >= newSize {
			e.cursorX = newSize
			if e.cursorX < e.left {
				// Out of viewing area.
				e.left = e.cursorX
			}
		}
	} else if event.Key() == tcell.KeyLeft {
		// Left.
		if e.cursorX == 0 {
			if e.cursorY == 0 {
				e.screen.Beep()
				return true
			}
			e.cursorY -= 1
			e.cursorX = e.buffer.Data()[e.cursorY].Size()
			if e.cursorX >= e.left+e.width {
				// Out of viewing area.
				e.left = e.cursorX - e.width
			}
			return true
		}
		e.cursorX -= 1
	} else if event.Key() == tcell.KeyRight {
		// Right.
		if e.cursorX >= e.buffer.Data()[e.cursorY].Size() {
			if e.cursorY >= e.buffer.Size()-1 {
				e.screen.Beep()
				return true
			}
			e.cursorY += 1
			e.cursorX = 0
			e.left = 0
			return true
		}
		e.cursorX += 1
	} else if event.Key() == tcell.KeyEnter {
		// Insert new line.
		var err error
		e.cursorY, e.cursorX, err = e.buffer.InsertOne('\n', e.cursorY, e.cursorX)
		if err != nil {
			e.screen.Beep()
			return true
		}
	} else if event.Key() == tcell.KeyBackspace {
		// Backspace.
		var err error
		e.cursorY, e.cursorX, err = e.buffer.DeleteOne(e.cursorY, e.cursorX)
		if err != nil {
			e.screen.Beep()
			return true
		}
	} else if event.Key() == tcell.KeyCtrlS {
		// Save.
		err := e.Save()
		e.Render()
		if err != nil {
			style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
			for i := range err.Error() {
				e.screen.SetContent(i, e.height-1, rune(err.Error()[i]), nil, style)
			}
			e.screen.Beep()
			e.screen.Sync()
			return false
		}
		style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
		for i := range "saved" {
			e.screen.SetContent(i, e.height-1, rune("saved"[i]), nil, style)
		}
		e.screen.Sync()
		return false
	} else {
		// Insert.
		var err error
		e.cursorY, e.cursorX, err = e.buffer.InsertOne(event.Rune(), e.cursorY, e.cursorX)
		if err != nil {
			e.screen.Beep()
			return true
		}
	}
	return true
}

// Render the buffer.
func (e *Editor) Render() error {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	e.screen.Clear()

	// Calculate the top position.
	// If the cursor is out of the current viewport, move the viewport.
	if e.cursorY >= e.top+e.height {
		e.top += 1
	} else if e.cursorY < e.top {
		e.top -= 1
	}

	// Calculate the left position.
	// If the cursor is out of the current viewport, move the viewport.
	if e.cursorX >= e.left+e.width {
		e.left += 1
	} else if e.cursorX < e.left {
		e.left -= 1
	}

	// Draw the text.
	for i := 0; i < e.height; i++ {
		line := e.top + i
		if line > e.buffer.Size()-1 {
			break
		}

		lineData := e.buffer.Data()[line]
		if e.left >= lineData.Size() {
			continue
		}
		display := lineData.Data()[e.left:]
		for j := range display {
			e.screen.SetContent(j, i, display[j], nil, style)
		}
	}

	e.screen.ShowCursor(e.cursorX-e.left, e.cursorY-e.top)

	// Sync and return.
	e.screen.Sync()
	return nil
}

// Save.
func (e *Editor) Save() error {
	// Open the file.
	file, err := os.OpenFile(e.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the buffer.
	for i := range e.buffer.Data() {
		if _, err := file.WriteString(string(e.buffer.Data()[i].Data())); err != nil {
			return err
		}
		if i <= e.buffer.Size()-1 {
			file.WriteString("\r\n")
		}
	}

	// Return.
	return nil
}

// Exit the editor.
func (e *Editor) Exit() {
	e.screen.Fini()
	e.running = false
}
