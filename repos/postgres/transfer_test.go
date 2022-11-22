package postgres

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

func TestTransferTx(t *testing.T) {
	store := NewSimpleBankDB(nil)

	tests := []struct {
		desc           string
		transferParams TransferTxParams
		hasErr         bool
	}{
		{
			desc: "success - transfer completed",
			transferParams: TransferTxParams{
				FromAccountID: 1,
				ToAccountID:   2,
				Amount:        3,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			result, err := store.TransferTx(ctx, tc.transferParams)
			if !tc.hasErr && err == nil {
				t.Fatal(err)
			}
			fmt.Println(result)
		})
	}
}
