package i2cssd1306

import (
	"bitbucket.org/gmcbay/i2c"
	// "image/jpeg"
	// "os"
)

const (
	SSD1306_I2C_ADDRESS         = 0x3C // 011110+SA0+RW - 0x3C or 0x3D
	SSD1306_SETCONTRAST         = 0x81
	SSD1306_DISPLAYALLON_RESUME = 0xA4
	SSD1306_DISPLAYALLON        = 0xA5
	SSD1306_NORMALDISPLAY       = 0xA6
	SSD1306_INVERTDISPLAY       = 0xA7
	SSD1306_DISPLAYOFF          = 0xAE
	SSD1306_DISPLAYON           = 0xAF
	SSD1306_SETDISPLAYOFFSET    = 0xD3
	SSD1306_SETCOMPINS          = 0xDA
	SSD1306_SETVCOMDETECT       = 0xDB
	SSD1306_SETDISPLAYCLOCKDIV  = 0xD5
	SSD1306_SETPRECHARGE        = 0xD9
	SSD1306_SETMULTIPLEX        = 0xA8
	SSD1306_SETLOWCOLUMN        = 0x00
	SSD1306_SETMEMORYMODEVERT   = 0x01
	SSD1306_SETMEMORYMODEPAGE   = 0x02
	SSD1306_SETHIGHCOLUMN       = 0x10
	SSD1306_SETSTARTLINE        = 0x40
	SSD1306_MEMORYMODE          = 0x20
	SSD1306_COLUMNADDR          = 0x21
	SSD1306_PAGEADDR            = 0x22
	SSD1306_COMSCANINC          = 0xC0
	SSD1306_COMSCANDEC          = 0xC8
	SSD1306_SEGREMAP            = 0xA0
	SSD1306_CHARGEPUMP          = 0x8D
	SSD1306_EXTERNALVCC         = 0x1
	SSD1306_SWITCHCAPVCC        = 0x2

	// Scrolling constants
	SSD1306_ACTIVATE_SCROLL                      = 0x2F
	SSD1306_DEACTIVATE_SCROLL                    = 0x2E
	SSD1306_SET_VERTICAL_SCROLL_AREA             = 0xA3
	SSD1306_RIGHT_HORIZONTAL_SCROLL              = 0x26
	SSD1306_LEFT_HORIZONTAL_SCROLL               = 0x27
	SSD1306_VERTICAL_AND_RIGHT_HORIZONTAL_SCROLL = 0x29
	SSD1306_VERTICAL_AND_LEFT_HORIZONTAL_SCROLL  = 0x2A
)

type bitmap struct {
	cols        int
	rows        int
	bytesPerCol int
	data        []byte
}

func (b *bitmap) Init(cols int, rows int) {
	b.cols = cols
	b.rows = rows
	b.bytesPerCol = rows / 8
	b.data = make([]byte, cols*b.bytesPerCol)
}

func (b *bitmap) Clear() {
	for i, _ := range b.data {
		b.data[i] = 0
	}
}

func (b *bitmap) DrawPixel(x int, y int, on bool) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	memCol := x
	memRow := y / 8
	bitMask := 1 << (uint(y) % 8)
	offset := memRow + (b.rows / 8 * memCol)
	if on {
		b.data[offset] |= byte(bitMask)
	} else {
		b.data[offset] &= byte((0xff - bitMask))
	}
}

func (b *bitmap) ClearBlock(x0 int, y0 int, dx int, dy int) {
	xLength := x0 + dx
	yLength := y0 + dy
	for x := x0; x < xLength; x++ {
		for y := y0; y < yLength; y++ {
			b.DrawPixel(x, y, false)
		}
	}
}

type SSD1306 struct {
	bus    *i2c.I2CBus
	bmap   *bitmap
	addr   byte
	height int
	width  int
	pages  int
}

func NewDevice() *SSD1306 {
	return &SSD1306{}
}

func (l *SSD1306) Init(busNumber byte, addr byte, height int, width int) error {
	var err error
	l.bus, err = i2c.Bus(busNumber)
	l.bmap = &bitmap{}
	l.bmap.Init(width, height)
	l.addr = addr
	l.height = height
	l.width = width
	l.pages = height / 8
	return err
}

func (l *SSD1306) command(c int) {
	control := 0x00 // Co = 0, DC = 0
	l.bus.WriteByte(l.addr, byte(control), byte(c))
}

func (l *SSD1306) InitDevice() {
	l.command(SSD1306_DISPLAYOFF)
	l.command(SSD1306_SETDISPLAYCLOCKDIV)
	l.command(0x80) // the suggested ratio 0x80
	l.command(SSD1306_SETMULTIPLEX)
	l.command(0x3F)
	l.command(SSD1306_SETCOMPINS)
	l.command(0x12)

	l.command(SSD1306_SETDISPLAYOFFSET)
	l.command(0x0) // no offset

	l.command(SSD1306_SETSTARTLINE | 0x0) // line #0

	l.command(SSD1306_CHARGEPUMP)
	l.command(0x14)

	l.command(SSD1306_MEMORYMODE)
	l.command(0x00) // 0x0 act like ks0108
	l.command(SSD1306_SEGREMAP | 0x1)
	l.command(SSD1306_COMSCANDEC)
	l.command(SSD1306_SETCONTRAST)
	l.command(0x8F)

	l.command(SSD1306_SETPRECHARGE)
	l.command(0xF1)

	l.command(SSD1306_SETVCOMDETECT)
	l.command(0x40)

	l.command(SSD1306_NORMALDISPLAY)
	l.command(SSD1306_DISPLAYON)
}

func (s *SSD1306) Clear() {
	s.bmap.Clear()
	s.Display()
}

func (l *SSD1306) data(bytes []byte) {
	control := 0x40 // Co = 0, DC = 0
	l.bus.WriteByteBlock(l.addr, byte(control), bytes)
}

func (s *SSD1306) Display() {
	// s.displayBlock(0, 0, s.width, 0)

	s.command(SSD1306_COLUMNADDR)
	s.command(0)           // Column start address. (0 = reset)
	s.command(s.width - 1) // Column end address.
	s.command(SSD1306_PAGEADDR)
	s.command(0)           // Page start address. (0 = reset)
	s.command(s.pages - 1) // Page end address.
	length := len(s.bmap.data)
	for i := 0; i < length; i += 16 {
		s.data(s.bmap.data[i : i+16])
	}
}

func (s *SSD1306) DeactivateScroll() {
	s.command(SSD1306_DEACTIVATE_SCROLL)
}

func (s *SSD1306) ActivateScroll() {
	s.command(SSD1306_ACTIVATE_SCROLL)
}

func (s *SSD1306) GetPages() int {
	return s.pages
}

func (s *SSD1306) WriteData(d byte, pos int) {
	s.bmap.data[pos] = d
}

func (s *SSD1306) SetStartLine(pos int) {
	s.command(SSD1306_SETSTARTLINE | pos)
}

func (s *SSD1306) SetAndActiveScroll(speed int) {

}
