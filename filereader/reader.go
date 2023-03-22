package filereader

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/axgle/mahonia"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

type Reader struct {
	File        *os.File
	bufioReader *bufio.Reader
}

func (r *Reader) Seek(offset int64, whence int) {
	r.File.Seek(offset, whence)
	r.bufioReader = bufio.NewReaderSize(r.File, 8*1024)
}

func (r *Reader) ReadUInt32() (uint32, error) {
	buf, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(buf), nil
}

func (r *Reader) ReadUInt16() (uint16, error) {
	buf, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(buf), nil
}

func (r *Reader) ReadInt32() (int32, error) {
	d, err := r.ReadUInt32()
	return int32(d), err
}

func (r *Reader) ReadBool() (bool, error) {
	buf, err := r.ReadBytes(1)
	if err != nil {
		return false, err
	}

	return buf[0] != 0, nil
}

func (r *Reader) ReadFloat() (float32, error) {
	buf, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}

	return math.Float32frombits(binary.LittleEndian.Uint32(buf)), nil
}

func (r *Reader) ReadString(n uint32) (string, error) {
	if n == 0 {
		return "", nil
	}
	buf, err := r.ReadBytes(int(n))
	if err != nil {
		return "", nil
	}

	// buf, err = GetUTF8(buf)
	// if err != nil {
	// 	return "", err
	// }
	return ConvertToString(string(buf), "gbk", "utf-8"), nil
}

func (r *Reader) ReadBytes(n int) ([]byte, error) {
	buf := make([]byte, n)

	total := 0
	for total < n {
		readN, err := r.bufioReader.Read(buf[total:])
		if err != nil {
			return nil, err
		}
		total += readN
	}

	return buf, nil
}

func GetUTF8(data []byte) (result []byte, err error) {
	encode, name, _ := charset.DetermineEncoding(data, "")
	fmt.Println(name)
	if name == "utf-8" {
		result = data
		return
	}
	newReader := transform.NewReader(bytes.NewReader(data), encode.NewDecoder())
	result, err = ioutil.ReadAll(newReader)

	return
}

func ConvertToString(src string, srcCode string, tagCode string) string {

	srcCoder := mahonia.NewDecoder(srcCode)

	srcResult := srcCoder.ConvertString(src)

	tagCoder := mahonia.NewDecoder(tagCode)

	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)

	result := string(cdata)

	return result

}
