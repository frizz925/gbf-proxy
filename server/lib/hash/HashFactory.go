package hash

import (
	"crypto/sha1"
	"hash"
)

type HashFactory interface {
	New() hash.Hash
}

type Sha1HashFactory struct{}

func (*Sha1HashFactory) New() hash.Hash {
	return sha1.New()
}
