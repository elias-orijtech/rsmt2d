package rsmt2d_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/celestiaorg/rsmt2d"
)

func TestEdsRepairRoundtripSimple(t *testing.T) {
	bufferSize := 64
	tests := []struct {
		name string
		// Size of each share, in bytes
		shareSize int
		codec     rsmt2d.Codec
	}{
		{"leopardFF8", bufferSize, rsmt2d.NewLeoRSFF8Codec()},
		{"leopardFF16", bufferSize, rsmt2d.NewLeoRSFF16Codec()},
		{"infectiousGF8", bufferSize, rsmt2d.NewRSGF8Codec()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ones := bytes.Repeat([]byte{1}, bufferSize)
			twos := bytes.Repeat([]byte{2}, bufferSize)
			threes := bytes.Repeat([]byte{3}, bufferSize)
			fours := bytes.Repeat([]byte{4}, bufferSize)

			// Compute parity shares
			eds, err := rsmt2d.ComputeExtendedDataSquare(
				[][]byte{
					ones, twos,
					threes, fours,
				},
				tt.codec,
				rsmt2d.NewDefaultTree,
			)
			if err != nil {
				t.Errorf("ComputeExtendedDataSquare failed: %v", err)
			}

			rowRoots := eds.RowRoots()
			colRoots := eds.ColRoots()

			// Save all shares in flattened form.
			flattened := make([][]byte, 0, eds.Width()*eds.Width())
			for i := uint(0); i < eds.Width(); i++ {
				flattened = append(flattened, eds.Row(i)...)
			}

			// Delete some shares, just enough so that repairing is possible.
			flattened[0], flattened[2], flattened[3] = nil, nil, nil
			flattened[4], flattened[5], flattened[6], flattened[7] = nil, nil, nil, nil
			flattened[8], flattened[9], flattened[10] = nil, nil, nil
			flattened[12], flattened[13] = nil, nil

			// Re-import the data square.
			eds, err = rsmt2d.ImportExtendedDataSquare(flattened, tt.codec, rsmt2d.NewDefaultTree)
			if err != nil {
				t.Errorf("ImportExtendedDataSquare failed: %v", err)
			}

			// Repair square.
			err = eds.Repair(
				rowRoots,
				colRoots,
			)
			if err != nil {
				// err contains information to construct a fraud proof
				// See extendeddatacrossword_test.go
				t.Errorf("RepairExtendedDataSquare failed: %v", err)
			}
		})
	}
}

func TestEdsRepairTwice(t *testing.T) {
	bufferSize := 64
	tests := []struct {
		name string
		// Size of each share, in bytes
		shareSize int
		codec     rsmt2d.Codec
	}{
		{"leopardFF8", bufferSize, rsmt2d.NewLeoRSFF8Codec()},
		{"leopardFF16", bufferSize, rsmt2d.NewLeoRSFF16Codec()},
		{"infectiousGF8", bufferSize, rsmt2d.NewRSGF8Codec()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ones := bytes.Repeat([]byte{1}, bufferSize)
			twos := bytes.Repeat([]byte{2}, bufferSize)
			threes := bytes.Repeat([]byte{3}, bufferSize)
			fours := bytes.Repeat([]byte{4}, bufferSize)

			// Compute parity shares
			eds, err := rsmt2d.ComputeExtendedDataSquare(
				[][]byte{
					ones, twos,
					threes, fours,
				},
				tt.codec,
				rsmt2d.NewDefaultTree,
			)
			if err != nil {
				t.Errorf("ComputeExtendedDataSquare failed: %v", err)
			}

			rowRoots := eds.RowRoots()
			colRoots := eds.ColRoots()

			// Save all shares in flattened form.
			flattened := make([][]byte, 0, eds.Width()*eds.Width())
			for i := uint(0); i < eds.Width(); i++ {
				flattened = append(flattened, eds.Row(i)...)
			}

			// Delete some shares, just enough so that repairing is possible, then remove one more.
			missing := make([]byte, bufferSize)
			copy(missing, flattened[1])
			flattened[0], flattened[1], flattened[2], flattened[3] = nil, nil, nil, nil
			flattened[4], flattened[5], flattened[6], flattened[7] = nil, nil, nil, nil
			flattened[8], flattened[9], flattened[10] = nil, nil, nil
			flattened[12], flattened[13] = nil, nil

			// Re-import the data square.
			eds, err = rsmt2d.ImportExtendedDataSquare(flattened, tt.codec, rsmt2d.NewDefaultTree)
			if err != nil {
				t.Errorf("ImportExtendedDataSquare failed: %v", err)
			}

			// Repair square.
			err = eds.Repair(
				rowRoots,
				colRoots,
			)
			if !errors.Is(err, rsmt2d.ErrUnrepairableDataSquare) {
				// Should fail since insufficient data.
				t.Errorf("RepairExtendedDataSquare did not fail with `%v`, got `%v`", rsmt2d.ErrUnrepairableDataSquare, err)
			}
			// Re-insert missing share and try again.
			flattened[1] = make([]byte, bufferSize)
			copy(flattened[1], missing)

			// Re-import the data square.
			eds, err = rsmt2d.ImportExtendedDataSquare(flattened, tt.codec, rsmt2d.NewDefaultTree)
			if err != nil {
				t.Errorf("ImportExtendedDataSquare failed: %v", err)
			}

			err = eds.Repair(
				rowRoots,
				colRoots,
			)
			if err != nil {
				// Should now pass, since sufficient data.
				t.Errorf("RepairExtendedDataSquare failed: %v", err)
			}

		})
	}
}
