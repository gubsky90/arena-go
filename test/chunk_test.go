package test

import (
	"testing"

	"github.com/thebagchi/arena-go/res"
)

func TestNewChunkSizingLogic(t *testing.T) {
	const (
		MIN_BLOCK_SIZE = 16
		MIN_DATA_SIZE  = 256 * 1024
	)

	var (
		initialRequestedSize        = uint64(16_000_000)
		requestedDataSize           = res.RoundPow2(initialRequestedSize)
		dataSize                    = requestedDataSize
		bitsSize             uint64 = dataSize / (4 * MIN_BLOCK_SIZE)
		roundedDataSize      uint64 = dataSize
	)

	if remainder := roundedDataSize % MIN_DATA_SIZE; remainder != 0 {
		roundedDataSize = roundedDataSize + (MIN_DATA_SIZE - remainder)
	}

	var (
		totalChunkSize  uint64 = bitsSize + roundedDataSize
		blocksAvailable uint64 = roundedDataSize / MIN_BLOCK_SIZE
	)

	t.Logf("Initial requested size: %d bytes", initialRequestedSize)
	t.Logf("Requested data size: %d bytes", requestedDataSize)
	t.Logf("Data size: %d bytes", dataSize)
	t.Logf("Bits size: %d bytes", bitsSize)
	t.Logf("Rounded data size: %d bytes", roundedDataSize)
	t.Logf("Total chunk size: %d bytes", totalChunkSize)
	t.Logf("Blocks available: %d", blocksAvailable)

	if blocksAvailable < 1_000_000 {
		t.Fatalf("insufficient blocks: %d available, need 1,000,000", blocksAvailable)
	}
}
