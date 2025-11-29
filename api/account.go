package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/roman-adamchik/simplebank/db/sqlc"
)

type CreateAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR ILS"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req CreateAccountRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Balance:  0,
		Currency: req.Currency,
	}

	acc, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		errorResponse(err)
		return
	}

	ctx.JSON(http.StatusOK, acc)
}
