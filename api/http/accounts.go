package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
	"github.com/orlandorode97/simple-bank/pkg/token"
)

func (s *Server) addAccountRoutes(r *gin.RouterGroup) {
	accounts := r.Group("/accounts")

	accounts.POST("/", s.createAccount)
	accounts.GET("/", s.listAccounts)
	accounts.GET("/:id", s.getAccount)
}

type createAccountRequest struct {
	Owner      string `json:"owner" binding:"required"`
	CurrencyID int64  `json:"currency_id" binding:"required,oneof=1 2 3 4 5"`
}

func (s *Server) createAccount(ctx *gin.Context) {
	payload := ctx.MustGet(string(authorizationKey)).(*token.Payload)

	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := simplebanksql.CreateAccountParams{
		Owner:      payload.Username,
		CurrencyID: req.CurrencyID,
		Balance:    0,
	}

	account, err := s.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation":
				ctx.JSON(http.StatusUnprocessableEntity, errorResponse(err))
				return
			case "unique_violation":
				ctx.JSON(http.StatusConflict, errorResponse(err))
				return
			default:
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
		}
	}

	ctx.JSON(http.StatusCreated, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getAccount gets the account that the user owns
func (s *Server) getAccount(ctx *gin.Context) {

	payload := ctx.MustGet(string(authorizationKey)).(*token.Payload)
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := s.store.GetAccount(ctx, req.ID)
	if errors.Is(err, sql.ErrNoRows) {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if account.Owner != payload.Username { // Check if the user owns an account based on the auth token payload username
		err := errors.New("account does not belong to authenticated user")
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// listAccounts lists all accounts that a user owns
func (s *Server) listAccounts(ctx *gin.Context) {
	payload := ctx.MustGet(string(authorizationKey)).(*token.Payload)

	var req listAccountsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := simplebanksql.ListAccountsParams{
		Owner:  payload.Username,
		Limit:  req.PageSize,                    // limit is the page size
		Offset: (req.PageID - 1) * req.PageSize, // records to skip
	}

	accounts, err := s.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"accounts": accounts})
}
