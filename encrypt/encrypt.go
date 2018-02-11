package encrypt

import (
	"bytes"
	"encoding/binary"
	"github.com/sencydai/qyh15c/crc16"
	"math/rand"
	"time"
)

type Encrypt struct {
	sKey        uint32
	sSelfSalt   uint32
	sTargetSalt uint32
	sKeyBuff    []byte
}

func NewEncrypt() *Encrypt {
	en := &Encrypt{}
	en.sSelfSalt = en.makeSalt()
	en.sKeyBuff = make([]byte, 4)
	return en
}

func (e *Encrypt) makeSalt() uint32 {
	return uint32(rand.Float32() * float32(time.Now().Unix()))
}

func (e *Encrypt) Encode(inBuff []byte, offset, length int64) int64 {
	if offset >= int64(len(inBuff)) {
		return 0
	}
	var end int64
	if length > 0 {
		end = offset + length
		if end > int64(len(inBuff)) {
			end = int64(len(inBuff))
		}
	} else {
		end = int64(len(inBuff))
	}
	pos := offset
	for i := offset; i < end; i++ {
		inBuff[pos] = inBuff[pos] ^ e.sKeyBuff[i%4]
		pos++
	}
	return end - offset
}

func (e *Encrypt) Decode(inBuff []byte, offset, length int64) int64 {
	return e.Encode(inBuff, offset, length)
}

func (e *Encrypt) GetCRC16(data []byte, length int64) int16 {
	return crc16.Update(data, 0, length)
}

func (e *Encrypt) GetCRC16ByPos(data []byte, offset, length int64) int16 {
	return crc16.Update(data, offset, length)
}

func (e *Encrypt) GetCheckKey() int16 {
	reader := bytes.NewBuffer([]byte{})
	binary.Write(reader, binary.LittleEndian, e.sKey)
	return crc16.Update(reader.Bytes(), 0, 0)
}

func (e *Encrypt) GetSelfSalt() uint32 {
	return e.sSelfSalt
}

func (e *Encrypt) GetTargetSalt() uint32 {
	return e.sTargetSalt
}

func (e *Encrypt) SetTargetSalt(v uint32) {
	e.sTargetSalt = v
	e.makeKey()
}

func (e *Encrypt) makeKey() {
	e.sKey = (e.sSelfSalt ^ e.sTargetSalt) + 8254
	for i := 0; i < 4; i++ {
		e.sKeyBuff[i] = byte((e.sKey & (0xff << uint(i<<3))) >> uint(i<<3))
	}
}
