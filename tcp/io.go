package tcp

import (
	"bufio"
	"io"
)

type ReaderImpl struct {
	reader             io.Reader
	rawReader          io.Reader
	buf                *bufio.Reader
	snapshot           *[]byte
	maxSnapshotBufSize int
}

// NewReader creates a reader.
// bufferSize: 0 for unbuffered reader. Otherwise use buffered reader.
func NewReader(rd io.Reader, bufferSize int) *ReaderImpl {
	ret := ReaderImpl{
		rawReader: rd,
	}

	if bufferSize > 0 {
		ret.buf = bufio.NewReaderSize(rd, bufferSize)
		ret.reader = ret.buf
	} else {
		ret.reader = rd
	}

	return &ret
}

func (self *ReaderImpl) Read(p []byte) (n int, err error) {
	if self.snapshot == nil {
		return self.reader.Read(p)
	}

	n, err = self.reader.Read(p)

	if n > 0 {
		buf := *self.snapshot
		l := len(buf)
		bytesToAppend := l + n - self.maxSnapshotBufSize

		if bytesToAppend <= 0 {
			*self.snapshot = append(buf, p[:n]...)
			return
		}

		*self.snapshot = append(buf, p[:bytesToAppend]...)
		err = bufio.ErrBufferFull
	}
	return
}

func (self *ReaderImpl) ReadFull(p []byte) (n int, err error) {
	read := 0
	l := len(p)
	for n < l {
		read, err = self.Read(p[n:])
		n += read

		if err != nil {
			return
		}
	}

	return
}

func (self *ReaderImpl) ReadAll(p *[]byte) (ret []byte, err error) {
	var buf []byte

	if p != nil {
		buf = *p
	}

	l, n := 0, 0

	c := cap(buf)
	for {
		if l == c {
			buf = append(buf, 0)[:l]
			c = cap(buf)
		}

		n, err = self.Read(buf[l:c])
		l += n
		buf = buf[:l]
		if err != nil {
			if err == io.EOF {
				err = nil
			}

			if p != nil {
				*p = buf
			}
			return buf[:l], err
		}
	}
}

var MaxReadSliceLength int = 8192

// ReadSlice returns a buffer that contains bytes up to (including) delimiter.
// Note:
// 1. the buffer may get overwritten when next time Read is called for buffered reader
// 2. it is inefficient for unbuffered reader, as it reads one byte at a time
func (self *ReaderImpl) ReadSlice(delim byte, p *[]byte) (ret []byte, err error) {
	if self.buf != nil {
		return self.buf.ReadSlice(delim)
	}

	var buf []byte

	if p != nil {
		buf = *p
	}

	l, n := 0, 0

	c := cap(buf)

	for {
		if l > MaxReadSliceLength {
			return buf[:l], bufio.ErrBufferFull
		}

		if l == c {
			buf = append(buf, 0)[:l]
			c = cap(buf)
		}

		n, err = self.reader.Read(buf[l : l+1])
		l += n
		buf = buf[:l]

		if err != nil {
			if p != nil {
				*p = buf
			}
			return buf[:l], err
		}

		if n > 0 && buf[l-1] == delim {
			if p != nil {
				*p = buf
			}
			return buf[:l], nil
		}
	}
}

func (self *ReaderImpl) SetSnapshot(p *[]byte, maxBufSize int) {
	var buf []byte

	if p != nil {
		self.snapshot = p
		buf = *p
	} else {
		self.snapshot = &buf
	}

	self.maxSnapshotBufSize = maxBufSize

}

func (self *ReaderImpl) GetSnapshot(clearSnapshot bool) (ret []byte) {
	if self.snapshot != nil {
		ret = *self.snapshot

		if clearSnapshot {
			self.snapshot = nil
			self.maxSnapshotBufSize = 0
		}
	}

	return ret
}

func (self *ReaderImpl) Reset(rd io.Reader) {
	self.reader = rd
	if self.buf != nil {
		self.buf.Reset(rd)
	}
}

type WriterImpl struct {
	writer io.Writer
	buf    *bufio.Writer
}

// NewWriter creates a writer.
// bufferSize: 0 for unbuffered reader. Otherwise use buffered reader.
func NewWriter(wr io.Writer, bufferSize int) *WriterImpl {
	ret := WriterImpl{
		writer: wr,
	}

	if bufferSize > 0 {
		ret.buf = bufio.NewWriterSize(wr, bufferSize)
	}

	return &ret
}

func (self *WriterImpl) Write(p []byte) (n int, err error) {
	if self.buf == nil {
		return self.writer.Write(p)
	}

	return self.buf.Write(p)
}

func (self *WriterImpl) WriteAll(p ...[]byte) (n int, err error) {
	var wr io.Writer
	if self.buf == nil {
		wr = self.writer
	} else {
		wr = self.buf
	}

	var written int
	for _, value := range p {
		written, err = wr.Write(value)

		n += written
		if err != nil {
			return
		}
	}
	return
}

func (self *WriterImpl) Flush() error {
	if self.buf == nil {
		// No flush for unbuffered
		return nil
	}

	return self.buf.Flush()
}

func (self *WriterImpl) Reset(wr io.Writer) {
	self.writer = wr
	if self.buf != nil {
		self.buf.Reset(wr)
	}
}
