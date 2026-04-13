package gokakera

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"
)

func (k *Kakera) computeChecksum(data []byte) (string, error) {
	switch k.config.ChecksumType {
	case "crc32":
		return computeCRC32(data), nil
	case "sha256":
		return computeSHA256(data), nil
	default:
		return "", fmt.Errorf("unsupported checksum type: %s", k.config.ChecksumType)
	}
}

func computeCRC32(data []byte) string {
	sum := crc32.ChecksumIEEE(data)
	return fmt.Sprintf("%08x", sum)
}

func computeSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
