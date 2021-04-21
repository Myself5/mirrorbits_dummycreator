package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Size    int64
	ModTime string
	Sha1    string
	Sha256  string
	Md5     string
}

func main() {

	// You can get individual args with normal indexing.
	path := os.Args[1]
	dummyPath := path

	fi, err := os.Stat(path)
	if err != nil {
		return
	}

	if len(os.Args) >= 3 {
		trim := os.Args[2]

		if trim != "" {
			_, err := os.Stat(trim)
			dummyPath = strings.ReplaceAll(dummyPath, filepath.Dir(trim), "")
			if err != nil {
				return
			}
		}
	}

	h, err := HashFile(path)
	if err != nil {
		return
	}

	h.Size = fi.Size()
	h.ModTime = fi.ModTime().String()

	fileJson, _ := json.Marshal([]FileInfo{h})

	os.MkdirAll(filepath.Join("dummy/", filepath.Dir(dummyPath)), os.FileMode(0755))
	err = ioutil.WriteFile(filepath.Join("dummy/", dummyPath), fileJson, 0644)
	fmt.Println(filepath.Base(dummyPath) + " dummy created")
}

// HashFile generates a human readable hash of the given file path
func HashFile(path string) (hashes FileInfo, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	var writers []io.Writer

	hsha1 := newHasher(sha1.New(), &hashes.Sha1)
	defer hsha1.Close()
	writers = append(writers, hsha1)
	hsha256 := newHasher(sha256.New(), &hashes.Sha256)
	defer hsha256.Close()
	writers = append(writers, hsha256)
	hmd5 := newHasher(md5.New(), &hashes.Md5)
	defer hmd5.Close()
	writers = append(writers, hmd5)

	if len(writers) == 0 {
		return
	}

	w := io.MultiWriter(writers...)

	_, err = io.Copy(w, reader)
	if err != nil {
		return
	}

	return
}

type hasher struct {
	hash.Hash
	output *string
}

func newHasher(hash hash.Hash, output *string) hasher {
	return hasher{
		Hash:   hash,
		output: output,
	}
}

func (h hasher) Close() error {
	*h.output = hex.EncodeToString(h.Sum(nil))
	return nil
}
