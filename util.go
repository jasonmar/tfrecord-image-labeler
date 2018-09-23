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
	"crypto/sha256"
	"encoding/base64"
	"log"
	"strings"
)

func splitPath(path string) (string, string) {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path, ""
	}
	return path[:i], path[i+1 : len(path)]
}

func hash(p []byte) string {
	key := make([]byte, 0, 32)
	h := sha256.New()
	h.Write(p)
	h.Sum(key)
	return base64.URLEncoding.EncodeToString(key)
}

func requireArg(arg string, name string) {
	if arg == "" {
		log.Fatalf("Missing required argument: -%s\n", name)
	}
}
