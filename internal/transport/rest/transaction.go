package rest

import (
	"custom-banking/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Transaction struct {
	TransactionService TransactionService
}

func NewTransaction(s TransactionService) *Transaction {
	return &Transaction{s}
}

func (t *Transaction) InjectRoutes(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	transaction := r.Group("/account/:id/transaction").Use(middlewares...)
	{
		transaction.GET("/", t.getTransactionList)
	}
}

func (t Transaction) getTransactionList(ctx *gin.Context) {
	// http://localhost:8080/account/1/transaction?order=amount:desc|id:asc&page=1&per-page=100
	// http://localhost:8080/account/1/transaction?page=1&per-page=100

	sPage := ctx.Request.URL.Query().Get("page")
	sPerPage := ctx.Request.URL.Query().Get("per-page")
	rawOrdering := ctx.Request.URL.Query().Get("order")

	pag := models.Paginator{
		Page:    1,
		PerPage: 5,
	}

	if sPage != "" {
		page, err := strconv.Atoi(sPage)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong \"page\" request param", err))
			return
		}

		pag.Page = page
	}

	if sPerPage != "" {
		perPage, err := strconv.Atoi(sPerPage)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong \"per-page\" request param", err))
			return
		}

		pag.PerPage = perPage
	}

	orderings, err := buildOrderingMessage(rawOrdering, []string{"id", "date_updated"})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong ordering query param", err))
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting current user error", err))
		return
	}

	accountID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong \"id\" request param", err))
		return
	}

	domainTransactionsList, err := t.TransactionService.GetTransactionList(ctx, accountID, userID, orderings, pag)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting transactions list error", err))
		return
	}

	messageTransactionsList := make([]models.Transaction, 0, len(domainTransactionsList))
	for _, transaction := range domainTransactionsList {
		messageTransaction := models.Transaction{
			ID:              transaction.ID,
			FromAccount:     transaction.FromAccount,
			ToAccount:       transaction.ToAccount,
			Amount:          transaction.Amount,
			TransactionType: transaction.TransactionType,
			Status:          transaction.Status,
			DateCreated:     transaction.DateCreated,
			DateUpdated:     transaction.DateUpdated,
		}
		messageTransactionsList = append(messageTransactionsList, messageTransaction)
	}

	ctx.JSON(http.StatusOK, messageTransactionsList)
}
