package rest

import (
	"custom-banking/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type CardTransfersHandler struct {
	service CardTransfersService
}

func NewCardTransfersHandler(service CardTransfersService) *CardTransfersHandler {
	return &CardTransfersHandler{
		service: service,
	}
}

func (h *CardTransfersHandler) Register(api *gin.Engine, middlewares ...gin.HandlerFunc) {
	cards := api.Group("/cards")
	{
		cards.POST("/transfer", h.TransferBetweenCards)
		cards.GET("/info/:number", h.GetCardByNumber)
		cards.GET("/transaction/:id", h.GetTransactionDetails)
		cards.GET("/:id/transactions", h.ListCardTransactions)
		cards.GET("/user/:user_id", h.GetCardsByUserID)
	}
}

// TransferBetweenCards godoc
// @Summary Transfer money between cards
// @Description Transfer money from one card to another
// @Tags cards
// @Accept json
// @Produce json
// @Param request body models.CardTransferRequest true "Transfer details"
// @Success 200 {object} models.CardTransferResult
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cards/transfer [post]
func (h *CardTransfersHandler) TransferBetweenCards(c *gin.Context) {
	var request models.CardTransferRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.FromCardNumber == "" || request.ToCardNumber == "" {
		newErrorResponse(c, http.StatusBadRequest, "card numbers are required")
		return
	}

	if request.Amount <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "amount must be positive")
		return
	}

	result, err := h.service.TransferBetweenCards(request)
	if err != nil {
		if errors.Is(err, errors.New("insufficient funds on source card")) ||
			errors.Is(err, errors.New("source card is not active")) ||
			errors.Is(err, errors.New("destination card is not active")) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		logrus.WithError(err).Error("CardTransfersHandler.TransferBetweenCards: error transferring between cards")
		newErrorResponse(c, http.StatusInternalServerError, "failed to process transfer")
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCardByNumber godoc
// @Summary Get card information
// @Description Get detailed information about a card by its number
// @Tags cards
// @Accept json
// @Produce json
// @Param number path string true "Card number"
// @Success 200 {object} models.CardInfo
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cards/info/{number} [get]
func (h *CardTransfersHandler) GetCardByNumber(c *gin.Context) {
	cardNumber := c.Param("number")
	if cardNumber == "" {
		newErrorResponse(c, http.StatusBadRequest, "card number is required")
		return
	}

	card, err := h.service.GetCardByNumber(cardNumber)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersHandler.GetCardByNumber: error getting card")
		newErrorResponse(c, http.StatusNotFound, "card not found")
		return
	}

	c.JSON(http.StatusOK, card)
}

// GetTransactionDetails godoc
// @Summary Get transaction details
// @Description Get detailed information about a transaction
// @Tags cards
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.CardTransferResult
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cards/transaction/{id} [get]
func (h *CardTransfersHandler) GetTransactionDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid transaction id")
		return
	}

	transaction, err := h.service.GetTransactionDetails(id)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersHandler.GetTransactionDetails: error getting transaction")
		newErrorResponse(c, http.StatusNotFound, "transaction not found")
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// ListCardTransactions godoc
// @Summary List card transactions
// @Description Get a list of transactions for a specific card
// @Tags cards
// @Accept json
// @Produce json
// @Param id path int true "Card ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {array} models.CardTransferResult
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cards/{id}/transactions [get]
func (h *CardTransfersHandler) ListCardTransactions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid card id")
		return
	}

	page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	transactions, err := h.service.ListCardTransactions(id, page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersHandler.ListCardTransactions: error listing transactions")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve transactions")
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetCardsByUserID godoc
// @Summary List user cards
// @Description Get a list of cards for a specific user
// @Tags cards
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param account_id query int false "Filter by account ID"
// @Param card_type query string false "Filter by card type"
// @Param active query bool false "Filter by active status"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} paginatedResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cards/user/{user_id} [get]
func (h *CardTransfersHandler) GetCardsByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid user id")
		return
	}

	params := models.CardListParams{
		Page:     1,
		PageSize: 10,
	}

	accountIDStr := c.Query("account_id")
	if accountIDStr != "" {
		accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
		if err == nil {
			params.AccountID = accountID
		}
	}

	params.CardType = c.Query("card_type")

	activeStr := c.Query("active")
	if activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			params.Active = &active
		}
	}

	page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if err == nil && page > 0 {
		params.Page = page
	}

	pageSize, err := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)
	if err == nil && pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}

	cards, totalCount, err := h.service.GetCardsByUserID(userID, params)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersHandler.GetCardsByUserID: error getting cards")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve cards")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       cards,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	})
}
