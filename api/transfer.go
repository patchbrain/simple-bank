package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"net/http"
)

type createTransferRequest struct {
	FromAccountId int64  `json:"from_account_id" binding:"required"`
	ToAccountId   int64  `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount"  binding:"required,gt=0"` // 大于0
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 检测参数的合法性: 货币类型是否正确，id是否存在等
	if !validateTransfer(ctx, s.Store, req.FromAccountId, req.Currency) {
		return
	}
	if !validateTransfer(ctx, s.Store, req.ToAccountId, req.Currency) {
		return
	}

	res, err := s.Store.TxTransfer(ctx, db.TxTransferParam{
		FromAccountId:  req.FromAccountId,
		ToAccountId:    req.ToAccountId,
		TransferAmount: req.Amount,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func validateTransfer(ctx *gin.Context, store db.Store, accountId int64, currency string) bool {
	account, err := store.GetAccount(ctx, accountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Currency != currency {
		err = fmt.Errorf("currency is mismatched: %s vs %s", currency, account.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}
