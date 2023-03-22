package ldt

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"latale_tool/filereader"
	"os"
	"path/filepath"
	"strings"
)

const (
	TYPE_UNSIGNED_INT = 0
	TYPE_STRING       = 1
	TYPE_BOOL         = 2
	TYPE_INT          = 3
	TYPE_FLOAT        = 4
)

// LDT LDT
type LDT struct {
	file   *os.File
	reader io.Reader
}

func (l *LDT) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(234)
		return err
	}
	filename = filepath.Clean(filename)
	fname := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)) + ".csv"
	os.Remove(fname)
	fmt.Println(fname)
	outFile, err := os.Create(fname)

	// outFile.WriteString("\xEF\xBB\xBF")
	outFile.Write([]byte{0xEF, 0xBB, 0xBF})
	csvWriter := csv.NewWriter(outFile)
	// csvWriter.Comma = '\t'

	l.file = file

	reader := &filereader.Reader{
		File: file,
	}

	reader.Seek(4, os.SEEK_SET)

	propCount, err := reader.ReadUInt32()
	if err != nil {
		fmt.Println(100)
		return err
	}

	itemCount, err := reader.ReadUInt32()

	propNames := make([]string, 0, propCount)
	for i := uint32(0); i < propCount; i++ {
		b, err := reader.ReadBytes(64)
		if err != nil {
			return fmt.Errorf("%d %w", i, err)
		}
		propName := string(bytes.Split(b, []byte{0})[0])
		propNames = append(propNames, propName)
	}

	reader.Seek(8204, os.SEEK_SET)
	propTypes := make([]uint32, 0, propCount)
	for i := uint32(0); i < propCount; i++ {
		propType, _ := reader.ReadUInt32()

		if propType > 4 {
			return fmt.Errorf("unknown type %d", propType)
		}

		propTypes = append(propTypes, propType)
	}

	headers := make([]string, 0, propCount+1)
	headers = append(headers, "item_id")
	for i, name := range propNames {
		t := propTypes[i]
		typeStr := ""
		switch t {
		case TYPE_BOOL:
			typeStr = "bool"
		case TYPE_FLOAT:
			typeStr = "float"
		case TYPE_INT:
			typeStr = "int"
		case TYPE_UNSIGNED_INT:
			typeStr = "uint"
		case TYPE_STRING:
			typeStr = "string"
		}

		s := fmt.Sprintf("%s|%s", name, typeStr)
		headers = append(headers, s)
	}

	csvWriter.Write(headers)
	csvWriter.Flush()

	reader.Seek(8716, os.SEEK_SET)
	for i := uint32(0); i < itemCount; i++ {
		itemID, _ := reader.ReadUInt32()
		props, err := l.readItem(reader, propTypes)
		if err != nil {
			fmt.Println(itemCount)
			return err
		}

		if itemID == 0 {
			continue
		}

		rowData := make([]string, 0, propCount+1)
		rowData = append(rowData, fmt.Sprint(itemID))
		for _, v := range props {
			rowData = append(rowData, fmt.Sprint(v))
		}
		csvWriter.Write(rowData)
	}
	csvWriter.Flush()

	return nil
}

func (l *LDT) readItem(reader *filereader.Reader, propTypes []uint32) ([]interface{}, error) {
	res := make([]interface{}, 0, len(propTypes))
	for _, propType := range propTypes {
		var err error
		var data interface{}
		switch propType {
		case TYPE_UNSIGNED_INT:
			data, err = reader.ReadUInt32()
		case TYPE_BOOL:
			data, err = reader.ReadUInt32()
			data = data.(uint32) != 0
		case TYPE_INT:
			data, err = reader.ReadInt32()
		case TYPE_FLOAT:
			data, err = reader.ReadFloat()
		case TYPE_STRING:
			l, err := reader.ReadUInt16()
			if err != nil {
				return nil, err
			}
			data, err = reader.ReadString(uint32(l))
		}

		if err != nil {
			fmt.Println(len(propTypes))
			return nil, err
		}

		res = append(res, data)
	}
	return res, nil
}
