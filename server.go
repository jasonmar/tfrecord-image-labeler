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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/julienschmidt/httprouter"
	"github.com/tensorflow/tensorflow/tensorflow/go/core/example"
)

type server struct {
	router   *httprouter.Router
	port     int
	tfrecord *TFRecordWriter
	images   *ImageIterator
}

func (s *server) handleImages() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")

		image := s.images.Next()
		js, err := json.Marshal(image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)
	}
}

func (s *server) handleImage() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "no-cache")

		imgID := ps.ByName("imgID")
		p, err := s.images.getImage(imgID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(p)
	}
}

func (s *server) handleStatic() httprouter.Handle {
	box := packr.NewBox("./static")
	fs := http.FileServer(box)
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")
		fs.ServeHTTP(w, r)
	}
}

func (s *server) handleLabel() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		decoder := json.NewDecoder(r.Body)
		var (
			err        error
			label      Label
			encodedJpg []byte
		)
		err = decoder.Decode(&label)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		source, filename := splitPath(label.SourceID)

		encodedJpg, err = s.images.getImage(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		encodedSha256 := hash(encodedJpg)

		example := &example.Example{
			Features: &example.Features{
				Feature: map[string]*example.Feature{
					"image/height":             IntFeature(label.Height),
					"image/width":              IntFeature(label.Width),
					"image/filename":           StringFeature(filename),
					"image/source_id":          StringFeature(source),
					"image/key/sha256":         StringFeature(encodedSha256),
					"image/encoded":            BytesFeature(encodedJpg),
					"image/format":             StringFeature(label.Format),
					"image/object/bbox/xmin":   FloatFeature(label.Xmin),
					"image/object/bbox/xmax":   FloatFeature(label.Xmax),
					"image/object/bbox/ymin":   FloatFeature(label.Ymin),
					"image/object/bbox/ymax":   FloatFeature(label.Ymax),
					"image/object/class/text":  StringFeature(label.Text),
					"image/object/class/label": IntFeature(label.Label),
				},
			},
		}

		err = s.tfrecord.Append(example)
		if err != nil {
			log.Printf("%v", err)
		}
	}
}

func (s *server) start() {
	s.routes()
	log.Printf("Running on port %d", s.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router))
}
