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
	"flag"
	"log"

	"github.com/julienschmidt/httprouter"
)

// Label contains features to be added to tf.Example
type Label struct {
	Height   int64   `json:"image/height"`
	Width    int64   `json:"image/width"`
	Filename string  `json:"image/filename"`
	SourceID string  `json:"image/source_id"`
	Format   string  `json:"image/format"`
	Xmin     float32 `json:"image/object/bbox/xmin"`
	Xmax     float32 `json:"image/object/bbox/xmax"`
	Ymin     float32 `json:"image/object/bbox/ymin"`
	Ymax     float32 `json:"image/object/bbox/ymax"`
	Text     string  `json:"image/object/class/text"`
	Label    int64   `json:"image/object/class/label"`
}

func main() {
	var (
		bucket string
		prefix string
		output string
	)
	flag.StringVar(&bucket, "bucket", "", "GCS bucket with images")
	flag.StringVar(&prefix, "prefix", "", "(optional) GCS prefix")
	flag.StringVar(&output, "output", "test.tfrecord", "TFRecord path")
	flag.Parse()
	requireArg(bucket, "bucket")

	images, err := NewImages(bucket, prefix)
	if err != nil {
		log.Fatal(err)
	}

	tfrecord, err := NewTFRecordWriter(output)
	if err != nil {
		log.Fatal(err)
	}

	server := &server{
		router:   httprouter.New(),
		port:     8080,
		tfrecord: tfrecord,
		images:   images,
	}
	server.start()
}
