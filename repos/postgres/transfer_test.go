package postgres

import (
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

func TestTransferTx(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		fmt.Println("yei")
	})
}
