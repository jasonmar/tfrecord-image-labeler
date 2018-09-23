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
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// ImageIterator loops through images
type ImageIterator struct {
	bucket    string
	prefix    string
	basePath  string
	filenames []string
	i         int
	limit     int
	ctx       context.Context
	client    *storage.Client
	cache     map[string][]byte
}

// Image contains the uri of an image
type Image struct {
	URI string `json:"uri"`
	ID  string `json:"id"`
}

// NewImages creates an ImageIterator
func NewImages(bucket, prefix string) (*ImageIterator, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	images := &ImageIterator{
		bucket:   bucket,
		prefix:   prefix,
		basePath: "/image",
		ctx:      ctx,
		client:   client,
		limit:    1024,
		cache:    make(map[string][]byte),
	}
	images.listImages()
	return images, nil
}

// Next returns the next image
func (it *ImageIterator) Next() *Image {
	i := it.i
	if i >= len(it.filenames) {
		i = 0
	}
	filename := it.filenames[i]
	img := &Image{
		URI: fmt.Sprintf("%s/%s", it.basePath, filename),
		ID:  filename,
	}
	i++
	it.i = i
	return img
}

// getImage reads a file from GCS
func (it *ImageIterator) getImageGCS(imgID string) ([]byte, error) {
	log.Printf("fetching gs://%s/%s", it.bucket, imgID)
	r, err := it.client.Bucket(it.bucket).Object(imgID).NewReader(it.ctx)
	if err != nil {
		return nil, err
	}

	p, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	log.Printf("read %d bytes from GCS", len(p))
	return p, nil
}

// getImage reads a file from GCS
func (it *ImageIterator) getImage(imgID string) ([]byte, error) {
	var (
		p     []byte
		found bool
		err   error
	)

	p, found = it.cache[imgID]
	if found {
		log.Printf("read %d bytes from cache", len(p))
		return p, nil
	}

	p, err = it.getImageGCS(imgID)
	if err != nil {
		return nil, err
	}
	it.limitCacheSize()
	it.cache[imgID] = p
	return p, nil
}

// limitCacheSize randomly deletes 10% of cached images
func (it *ImageIterator) limitCacheSize() {
	if len(it.cache) >= it.limit {
		i := 0
		for k := range it.cache {
			delete(it.cache, k)
			i++
			if i > it.limit/10 {
				break
			}
		}
	}
}

func setLabel(ctx context.Context, client *storage.Client, bucket string, prefix string) {
	obj := client.Bucket(bucket).Object(prefix)
	uattrs := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{"labeled": "true"},
	}
	_, err := obj.Update(ctx, uattrs)
	if err != nil {
		log.Printf("%v", err)
	}
}

// listImages updates list of images in bucket
func (it *ImageIterator) listImages() {
	log.Printf("listing gs://%s/%s", it.bucket, it.prefix)
	q := &storage.Query{
		Prefix:   it.prefix,
		Versions: false,
	}
	objects := it.client.Bucket(it.bucket).Objects(it.ctx, q)
	parts := make([]string, 0, it.limit)
	i := 0
	for {
		if i >= it.limit {
			break
		}
		obj, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if obj != nil && isImage(obj) {
			n := len(it.prefix)
			if n > 0 {
				parts = append(parts, obj.Name[n+1:])
			} else {
				parts = append(parts, obj.Name)
			}
			i++
		}
	}
	log.Printf("%v", parts)
	it.filenames = parts
}

func isImage(objectAttrs *storage.ObjectAttrs) bool {
	for _, imageSuffix := range []string{"jpg", "jpeg"} {
		if strings.HasSuffix(objectAttrs.Name, imageSuffix) {
			return true
		}
	}
	return false
}
