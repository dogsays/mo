package logger

type buffer struct {
	buf []byte
}

func (b *buffer) Write(d byte) {
	b.buf = append(b.buf, d)
}

func (b *buffer) WriteInt(s int) {
	itoa(&b.buf, s, -1)
}

func (b *buffer) WriteString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *buffer) Reset() {
	b.buf = b.buf[:0]
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
