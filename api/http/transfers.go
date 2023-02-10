package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
	"github.com/orlandorode97/simple-bank/pkg/token"
	"github.com/orlandorode97/simple-bank/store"
)

func (s *Server) addTransferRoutes(r *gin.RouterGroup) {
	transfers := r.Group("/transfers")

	transfers.POST("/", s.createTransfer)
}

type createTransferRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64 `json:"to_account_id" binding:"required,min=1"`
	Amount        int64 `json:"amount" binding:"required,gt=1"`
	CurrencyID    int64 `json:"currency_id" binding:"required,oneof=1 2 3 4 5"`
}

func (s *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := store.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	payload := ctx.MustGet(string(authorizationKey)).(*token.Payload)

	fromAccount, valid := s.validAccount(ctx, req.FromAccountID, req.CurrencyID)
	if !valid {
		return
	}

	if fromAccount.Owner != payload.Username {
		err := errors.New("form account does not belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	_, valid = s.validAccount(ctx, req.ToAccountID, req.CurrencyID)
	if !valid {
		return
	}

	transfer, err := s.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	ctx.JSON(http.StatusCreated, transfer)
}

// validAccount  valids the account and the account's currency
func (s *Server) validAccount(ctx *gin.Context, accountID int64, currencyID int64) (*simplebanksql.Account, bool) {
	account, err := s.store.GetAccount(ctx, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return &account, false
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return &account, false
	}

	if account.CurrencyID != currencyID {
		err = fmt.Errorf("account [%d] currency mismatch %v - %v", accountID, account.CurrencyID, currencyID)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return &account, false
	}

	return &account, true
}
