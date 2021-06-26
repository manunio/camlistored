package main

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"regexp"
)

var getPutPattern = regexp.MustCompile(`^/camli/(sha1)-([a-f0-9]+)$`)
var blobRefPattern = regexp.MustCompile(`^([a-z0-9]+)-([a-f0-9]+)$`)

// BlobRef ...
type BlobRef struct {
	HashName string
	Digest   string
}

var expectedDigestSize = map[string]int{
	"md5":  32,
	"sha1": 40,
}

func blobIfValid(hashName, digest string) *BlobRef {
	expectedSize := expectedDigestSize[hashName]
	if expectedSize != 0 && len(digest) != expectedSize {
		return nil
	}
	return &BlobRef{hashName, digest}
}

func blobFromPattern(r *regexp.Regexp, s string) *BlobRef {
	matches := r.FindAllStringSubmatch(s, -1)
	if len(matches) != 1 || len(matches[0]) != 3 {
		return nil
	}
	return blobIfValid(matches[0][1], matches[0][2])
}

// ParseBlobRef ...
func ParseBlobRef(ref string) *BlobRef {
	return blobFromPattern(blobRefPattern, ref)
}

// ParsePath ...
func ParsePath(path string) *BlobRef {
	return blobFromPattern(getPutPattern, path)
}

// IsSupported ...
func (o *BlobRef) IsSupported() bool {
	if o.HashName == "sha1" {
		return true
	}
	return false
}

// Hash ...
func (o *BlobRef) Hash() hash.Hash {
	if o.HashName == "sha1" {
		return sha1.New()
	}
	return nil
}

// FileBaseName ...
func (o *BlobRef) FileBaseName() string {
	return fmt.Sprintf("%s-%s.dat", o.HashName, o.Digest)
}

// DirectoryName ...
func (o *BlobRef) DirectoryName() string {
	return fmt.Sprintf("%s/%s/%s", *storageRoot, o.Digest[0:3], o.Digest[3:6])
}

// FileName ...
func (o *BlobRef) FileName() string {
	return fmt.Sprintf("%s/%s-%s.dat", o.DirectoryName(), o.HashName, o.Digest)
}
