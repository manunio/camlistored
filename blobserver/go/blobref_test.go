package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
