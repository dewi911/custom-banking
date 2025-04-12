package rest

import (
	"custom-banking/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type CashbackHandler struct {
	service CashbackService
}

func NewCashbackHandler(service CashbackService) *CashbackHandler {
	return &CashbackHandler{service: service}
}

func (h *CashbackHandler) InitRoutes(api *gin.RouterGroup) {
	cashback := api.Group("/cashback")
	{
		cashback.POST("/settings", h.configureCashback)
		cashback.GET("/settings/:cardId", h.getCashbackSettings)
		cashback.DELETE("/settings/:cardId", h.deactivateCashback)
		cashback.GET("/transactions", h.getCashbackTransactions)
		cashback.GET("/pending", h.getPendingCashback)

		// Admin-only routes
		cashback.POST("/process-payouts", h.processPayouts)
	}
}

// @Summary Configure cashback settings for a card
// @Description Configures or updates cashback settings for a specific card
// @Tags cashback
// @Accept json
// @Produce json
// @Param input body models.CashbackSettingsRequest true "Cashback settings request"
// @Success 200 {object} models.CashbackSettings
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cashback/settings [post]
func (h *CashbackHandler) configureCashback(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request models.CashbackSettingsRequest
	if err := c.BindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.CardID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "card ID is required")
		return
	}

	rate := 0.5
	if request.CashbackRate > 0 {
		rate = request.CashbackRate
	}

	minAmount := 0.0
	if request.MinTransactionAmount > 0 {
		minAmount = request.MinTransactionAmount
	}

	maxAmount := 100.0
	if request.MaxCashbackPerTransaction > 0 {
		maxAmount = request.MaxCashbackPerTransaction
	}

	err = h.service.ConfigureCashback(c.Request.Context(), userID, request.CardID, rate, minAmount, maxAmount)
	if err != nil {
		logrus.WithError(err).Error("cashbackHandler.configureCashback: error configuring cashback")
		if errors.Is(err, errors.New("card does not belong to the user")) {
			newErrorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error configuring cashback settings")
		return
	}

	settings, err := h.service.GetCashbackSettings(c.Request.Context(), request.CardID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "cashback settings configured successfully"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// @Summary Get cashback settings for a card
// @Description Retrieves the current cashback settings for a specific card
// @Tags cashback
// @Produce json
// @Param cardId path int true "Card ID"
// @Success 200 {object} models.CashbackSettings
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cashback/settings/{cardId} [get]
func (h *CashbackHandler) getCashbackSettings(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID, err := strconv.ParseInt(c.Param("cardId"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid card id param")
		return
	}

	settings, err := h.service.GetCashbackSettings(c.Request.Context(), cardID)
	if err != nil {
		logrus.WithError(err).Errorf("cashbackHandler.getCashbackSettings: error getting settings for card %d", cardID)
		newErrorResponse(c, http.StatusInternalServerError, "error getting cashback settings")
		return
	}

	if settings.UserID != userID {
		newErrorResponse(c, http.StatusForbidden, "you don't have access to these cashback settings")
		return
	}

	c.JSON(http.StatusOK, settings)
}

// @Summary Deactivate cashback for a card
// @Description Deactivates cashback for a specific card
// @Tags cashback
// @Produce json
// @Param cardId path int true "Card ID"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cashback/settings/{cardId} [delete]
func (h *CashbackHandler) deactivateCashback(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID, err := strconv.ParseInt(c.Param("cardId"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid card id param")
		return
	}

	settings, err := h.service.GetCashbackSettings(c.Request.Context(), cardID)
	if err != nil {
		logrus.WithError(err).Errorf("cashbackHandler.deactivateCashback: error getting settings for card %d", cardID)
		newErrorResponse(c, http.StatusInternalServerError, "error getting cashback settings")
		return
	}

	if settings.UserID != userID {
		newErrorResponse(c, http.StatusForbidden, "you don't have access to these cashback settings")
		return
	}

	err = h.service.DeactivateCashback(c.Request.Context(), cardID)
	if err != nil {
		logrus.WithError(err).Errorf("cashbackHandler.deactivateCashback: error deactivating cashback for card %d", cardID)
		newErrorResponse(c, http.StatusInternalServerError, "error deactivating cashback")
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "success", Message: "cashback deactivated successfully"})
}

// @Summary Get cashback transactions
// @Description Retrieves cashback transactions with optional filtering
// @Tags cashback
// @Produce json
// @Param card_id query int false "Filter by card ID"
// @Param status query string false "Filter by status"
// @Param start_date query string false "Filter by start date (format: YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (format: YYYY-MM-DD)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} paginatedResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cashback/transactions [get]
func (h *CashbackHandler) getCashbackTransactions(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var filter models.CashbackTransactionFilter

	if cardIDStr := c.Query("card_id"); cardIDStr != "" {
		cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid card_id param")
			return
		}
		filter.CardID = cardID
	}

	filter.Status = c.Query("status")

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid start_date format, use YYYY-MM-DD")
			return
		}
		filter.StartDate = startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid end_date format, use YYYY-MM-DD")
			return
		}
		endDate = endDate.Add(24*time.Hour - 1*time.Second)
		filter.EndDate = endDate
	}

	filter.Page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil || page < 1 {
			newErrorResponse(c, http.StatusBadRequest, "invalid page param")
			return
		}
		filter.Page = page
	}

	filter.PageSize = 10
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
		if err != nil || pageSize < 1 || pageSize > 100 {
			newErrorResponse(c, http.StatusBadRequest, "invalid page_size param")
			return
		}
		filter.PageSize = pageSize
	}

	transactions, totalCount, err := h.service.GetCashbackTransactions(c.Request.Context(), userID, filter)
	if err != nil {
		logrus.WithError(err).Error("cashbackHandler.getCashbackTransactions: error getting transactions")
		newErrorResponse(c, http.StatusInternalServerError, "error getting cashback transactions")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       transactions,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	})
}

// @Summary Get pending cashback amount
// @Description Retrieves the total pending cashback amount for the authenticated user
// @Tags cashback
// @Produce json
// @Success 200 {object} map[string]float64
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cashback/pending [get]
func (h *CashbackHandler) getPendingCashback(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	amount, err := h.service.GetPendingCashbackAmount(c.Request.Context(), userID)
	if err != nil {
		logrus.WithError(err).Errorf("cashbackHandler.getPendingCashback: error getting pending amount for user %d", userID)
		newErrorResponse(c, http.StatusInternalServerError, "error getting pending cashback amount")
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_amount": amount})
}

// @Summary Process cashback payouts (Admin only)
// @Description Processes all pending cashback payouts that are due
// @Tags cashback
// @Produce json
// @Success 200 {object} statusResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/cashback/process-payouts [post]
func (h *CashbackHandler) processPayouts(c *gin.Context) {
	if !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "admin access required")
		return
	}

	err := h.service.ProcessPayouts(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("cashbackHandler.processPayouts: error processing payouts")
		newErrorResponse(c, http.StatusInternalServerError, "error processing cashback payouts")
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "success", Message: "cashback payouts processed successfully"})
}
