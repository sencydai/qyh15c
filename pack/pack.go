package pack

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

func NewWriter(datas ...interface{}) *bytes.Buffer {
	writer := bytes.NewBuffer([]byte{})
	Write(writer, datas...)
	return writer
}

func GetBytes(datas ...interface{}) []byte {
	writer := NewWriter(datas...)
	return writer.Bytes()
}

func Read(reader *bytes.Reader, datas ...interface{}) {
	for _, data := range datas {
		switch v := data.(type) {
		case *bool, *int8, *uint8, *int16, *uint16, *int32, *uint32, *int64, *uint64, *float32, *float64:
			err := binary.Read(reader, binary.LittleEndian, v)
			if err != nil {
				panic(err.Error())
			}
		case *string:
			var l uint16
			err := binary.Read(reader, binary.LittleEndian, &l)
			if err != nil {
				panic(err.Error())
			}
			s := make([]byte, l)
			for i := uint16(0); i < l; i++ {
				s[i], err = reader.ReadByte()
				if err != nil {
					panic(err.Error())
				}
			}
			*v = string(s)
			_, err = reader.ReadByte()
			if err != nil {
				panic(err.Error())
			}
		default:
			panic("pack.Read invalid type " + reflect.TypeOf(data).String())
		}
	}
}

func Write(writer *bytes.Buffer, datas ...interface{}) {
	for _, data := range datas {
		switch v := data.(type) {
		case bool, *bool, int8, *int8, uint8, *uint8, int16, *int16, uint16, *uint16, int32, *int32, uint32, *uint32, int64, *int64, uint64, *uint64, float32, *float32, float64, *float64:
			binary.Write(writer, binary.LittleEndian, v)
		case []byte:
			writer.Write(v)
		case string:
			binary.Write(writer, binary.LittleEndian, uint16(len(v)))
			writer.Write([]byte(v))
			binary.Write(writer, binary.LittleEndian, byte(0))
		case *string:
			s := *v
			binary.Write(writer, binary.LittleEndian, uint16(len(s)))
			writer.Write([]byte(s))
			binary.Write(writer, binary.LittleEndian, byte(0))
		default:
			panic("pack.Write invalid type " + reflect.TypeOf(data).String())
		}
	}
}
