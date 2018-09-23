// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/tensorflow/tensorflow/tensorflow/go/core/example"
)

const delta = uint32(0xa282ead8)

var crcTable = crc32.MakeTable(crc32.Castagnoli)

func mask(v uint32) uint32 {
	return ((v >> 15) | (v << 17)) + delta
}

func unmask(v uint32) uint32 {
	r := v - delta
	return ((r >> 17) | (r << 15))
}

func crc(bs []byte, n int64) uint32 {
	return mask(crc32.Checksum(bs[:n], crcTable))
}

// TFRecordWriter writes records to a file
type TFRecordWriter struct {
	wc io.WriteCloser
}

// NewTFRecordWriter creates a new TFRecord file
func NewTFRecordWriter(path string) (*TFRecordWriter, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	return &TFRecordWriter{wc: file}, nil
}

// Append writes a tf.Example to the TFRecord
func (w *TFRecordWriter) Append(e *example.Example) error {
	p, err := proto.Marshal(e)
	if err != nil {
		return err
	}
	_, err = w.Write(p)
	return err
}

func (w *TFRecordWriter) Write(p []byte) (int, error) {
	length := uint64(len(p))
	if length == 0 {
		return 0, fmt.Errorf("empty data slice")
	}

	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs[0:], length)
	lengthCrc := crc(bs, 8)
	dataCrc := crc(p, int64(length))

	b := bytes.NewBuffer(make([]byte, 0, length+8+4+4))
	binary.Write(b, binary.LittleEndian, length)
	binary.Write(b, binary.LittleEndian, lengthCrc)
	b.Write(p[:])
	binary.Write(b, binary.LittleEndian, dataCrc)

	recordBytes := b.Bytes()
	n, err := w.wc.Write(recordBytes)
	log.Printf("Wrote %d bytes", len(recordBytes))
	return n, err
}

// Close closes the output stream
func (w *TFRecordWriter) Close() error {
	if w.wc != nil {
		w.wc.Close()
	}
	return nil
}

// Flush flushes remaining bytes
func (w *TFRecordWriter) Flush() error {
	return nil
}

func IntFeature(x int64) *example.Feature {
	return &example.Feature{
		Kind: &example.Feature_Int64List{
			Int64List: &example.Int64List{
				Value: []int64{x},
			},
		},
	}
}

func FloatFeature(x float32) *example.Feature {
	return &example.Feature{
		Kind: &example.Feature_FloatList{
			FloatList: &example.FloatList{
				Value: []float32{x},
			},
		},
	}
}

func BytesFeature(p []byte) *example.Feature {
	return &example.Feature{
		Kind: &example.Feature_BytesList{
			BytesList: &example.BytesList{
				Value: [][]byte{p},
			},
		},
	}
}

func StringFeature(s string) *example.Feature {
	return BytesFeature([]byte(s))
}
