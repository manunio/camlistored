package main

import (
	"crypto/sha1"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const ref = "sha1-9242dbadb7827d697fab034a1e73f366b451ce4d"

func TestParseBlobRef(t *testing.T) {

	tests := map[string]struct {
		input  string
		output *BlobRef
		fails  bool
	}{
		"fails on missing ref": {
			input:  "",
			output: nil,
			fails:  true,
		},

		"fails on invalid hash": {
			input:  "sha11-9242dbadb7827d697fab034a1e73f366b451ce4d",
			output: nil,
			fails:  true,
		},
		"fails on invalid digest": {
			input:  "sha1-9242dbadb7827d697fab034a1e73f366b451ce4",
			output: nil,
			fails:  true,
		},
		"passes on valid ref": {
			input: "sha1-9242dbadb7827d697fab034a1e73f366b451ce4d",
			output: &BlobRef{
				HashName: "sha1",
				Digest:   "9242dbadb7827d697fab034a1e73f366b451ce4d",
			},
			fails: false,
		},
	}

	for _, test := range tests {
		blobRef := ParseBlobRef(test.input)
		if test.fails {
			assert.Nil(t, blobRef, test)
		} else {
			assert.NotNil(t, blobRef, test)
		}
		assert.Equal(t, test.output, blobRef, test)
	}
}

func TestParsePath(t *testing.T) {
	tests := map[string]struct {
		input  string
		output *BlobRef
		fails  bool
	}{
		"fails on missing path": {
			input:  "",
			output: nil,
			fails:  true,
		}, "fails on invalid path": {
			input:  "/camli/sha11-9242dbadb7827d697fab034a1e73f366b451ce4d",
			output: nil,
			fails:  true,
		}, "fails on invalid digest": {
			input:  "/camli/sha1-9242dbadb7827d697fab034a1e73f366b451ce4",
			output: nil,
			fails:  true,
		},
		"passes on valid path": {
			input: "/camli/sha1-9242dbadb7827d697fab034a1e73f366b451ce4d",
			output: &BlobRef{
				HashName: "sha1",
				Digest:   "9242dbadb7827d697fab034a1e73f366b451ce4d",
			},
			fails: false,
		},
	}

	for _, test := range tests {
		blobRef := ParsePath(test.input)
		if test.fails {
			assert.Nil(t, blobRef, test)
		} else {
			assert.NotNil(t, blobRef, test)
		}
		assert.Equal(t, test.output, blobRef, test)
	}

}

func TestBlobRef_IsSupported(t *testing.T) {
	blobRef := ParseBlobRef(ref)
	assert.True(t, blobRef.IsSupported())
}

func TestBlobRef_Hash(t *testing.T) {
	blobRef := ParseBlobRef(ref)
	assert.Equal(t, sha1.New(), blobRef.Hash())
}

func TestBlobRef_DirectoryName(t *testing.T) {
	blobRef := ParseBlobRef(ref)
	digest := strings.Split(ref, "-")[1]
	assert.Equal(t, *storageRoot+"/"+digest[0:3]+"/"+digest[3:6], blobRef.DirectoryName())
}

func TestBlobRef_FileBaseName(t *testing.T) {
	blobRef := ParseBlobRef(ref)
	refSplit := strings.Split(ref, "-")
	hashName := refSplit[0]
	digest := refSplit[1]
	assert.Equal(t, hashName+"-"+digest+".dat", blobRef.FileBaseName())
}

func TestBlobRef_FileName(t *testing.T) {
	blobRef := ParseBlobRef(ref)
	refSplit := strings.Split(ref, "-")
	hashName := refSplit[0]
	digest := refSplit[1]
	directoryName := *storageRoot + "/" + digest[0:3] + "/" + digest[3:6]
	assert.Equal(t, directoryName+"/"+hashName+"-"+digest+".dat", blobRef.FileName())
}

func TestBlobRef_String(t *testing.T) {
	blobRef := ParseBlobRef(ref)
	refSplit := strings.Split(ref, "-")
	hashName := refSplit[0]
	digest := refSplit[1]
	assert.Equal(t, hashName+"-"+digest, blobRef.String())
}
