// buffer/buffer.go
// Package buffer provides an API for working with file buffers.

package buffer

import "errors"

var ErrInvalidPos = errors.New("invalid cursor position")
var ErrMaxSizeExceeded = errors.New("max size exceeded")
var ErrLineEmpty = errors.New("line already empty")

// Buffer struct.
type Buffer struct {
	// Internal buffer data.
	maxSize     int
	lineMaxSize int
	data        []*BufferLine
}

// Create a new buffer.
func NewBuffer(maxSize, lineMaxSize int) *Buffer {
	return &Buffer{
		maxSize:     maxSize,
		lineMaxSize: lineMaxSize,
		data:        make([]*BufferLine, 0),
	}
}

// Get max size.
func (b *Buffer) MaxSize() int {
	return b.maxSize
}

// Get max line size.
func (b *Buffer) MaxLineSize() int {
	return b.lineMaxSize
}

// Get line size.
func (b *Buffer) Size() int {
	return len(b.data)
}

// Get line data.
func (b *Buffer) Data() []*BufferLine {
	return b.data
}

// Set line data.
func (b *Buffer) SetData(lines []*BufferLine) {
	b.data = lines
}

// Insert a char. Returns the new position of the cursor.
func (b *Buffer) InsertOne(char rune, row, col int) (int, int, error) {
	if row > b.Size() {
		return row, col, ErrInvalidPos
	}
	if char == rune('\n') {
		// New line.
		if b.Size()+1 > b.MaxSize() {
			return row, col, ErrMaxSizeExceeded
		}
		origData := make([]*BufferLine, len(b.data))
		copy(origData, b.data)
		b.data = make([]*BufferLine, len(b.data)+1)

		// Create the new line and copy over the previous lines.
		copy(b.data[:row+1], origData[:row+1])
		copy(b.data[row+2:], origData[row+1:])

		// Split the line.
		splitLine := origData[row].data
		b.data[row].data = splitLine[:col]
		b.data[row+1] = NewBufferLine(b.MaxLineSize())
		b.data[row+1].data = splitLine[col:]

		// Return.
		return row + 1, 0, nil
	}

	// Add a char.
	return row, col + 1, b.data[row].Insert([]rune{char}, col)
}

// Delete a char. Returns the new position of the cursor.
func (b *Buffer) DeleteOne(row, col int) (int, int, error) {
	if row > b.Size() {
		return row, col, ErrInvalidPos
	}
	if col == 0 {
		// Move up one line.
		if row == 0 {
			return row, col, ErrInvalidPos
		}
		if b.data[row-1].Size()+b.data[row].Size() > b.MaxLineSize() {
			// Max size exceeded.
			return row, col, ErrMaxSizeExceeded
		}

		origData := make([]*BufferLine, len(b.data))
		copy(origData, b.data)
		b.data = make([]*BufferLine, len(b.data)-1)

		// Copy over the lines.
		copy(b.data[:row-1], origData[:row-1])
		copy(b.data[row-1:], origData[row:])

		// Join the lines.
		b.data[row-1].data = append(origData[row-1].data, origData[row].data...)

		// Return.
		return row - 1, len(origData[row-1].data), nil
	}

	// Delete a char.
	return row, col - 1, b.data[row].Delete(1, col)
}

// Buffer line struct.
type BufferLine struct {
	// Internal buffer data.
	maxSize int
	data    []rune
}

// Create a new buffer line.
func NewBufferLine(maxSize int) *BufferLine {
	return &BufferLine{
		maxSize: maxSize,
		data:    make([]rune, 0, maxSize),
	}
}

// Get max line size.
func (b *BufferLine) MaxSize() int {
	return b.maxSize
}

// Get line size.
func (b *BufferLine) Size() int {
	return len(b.data)
}

// Get line data.
func (b *BufferLine) Data() []rune {
	return b.data
}

// Insert into the line.
func (b *BufferLine) Insert(data []rune, pos int) error {
	if pos > b.Size() {
		return ErrInvalidPos
	}
	if len(data)+b.Size() > b.MaxSize() {
		return ErrMaxSizeExceeded
	}

	// Increase the size of the buffer.
	origData := make([]rune, len(b.data))
	copy(origData, b.data)
	b.data = make([]rune, len(origData)+len(data))

	// Copy over and set the new data.
	copy(b.data[:pos], origData[:pos])
	copy(b.data[pos:pos+len(data)], data)
	copy(b.data[pos+len(data):], origData[pos:])

	// Return.
	return nil
}

// Delete from the line.
func (b *BufferLine) Delete(num int, pos int) error {
	if pos > b.Size() {
		return ErrInvalidPos
	}
	if pos-num < 0 {
		return ErrLineEmpty
	}

	// Decrease the size of the buffer.
	origData := make([]rune, len(b.data))
	copy(origData, b.data)
	b.data = make([]rune, len(origData)-num)

	// Copy over the new data.
	copy(b.data[:pos-num], origData[:pos-num])
	copy(b.data[pos-num:], origData[pos:])

	// Return.
	return nil
}
