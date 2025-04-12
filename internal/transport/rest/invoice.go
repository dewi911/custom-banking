package rest

import (
	"custom-banking/internal/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type InvoiceHandler struct {
	service InvoiceService
}

func NewInvoiceHandler(service InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{service: service}
}

func (h *InvoiceHandler) InitRoutes(api *gin.RouterGroup) {
	invoice := api.Group("/invoice")
	{
		invoice.POST("", h.createInvoice)
		invoice.GET("", h.getUserInvoices)
		invoice.GET("/:id", h.getInvoice)
		invoice.GET("/number/:invoiceNumber", h.getInvoiceByNumber)
		invoice.DELETE("/:id", h.cancelInvoice)
		invoice.POST("/pay", h.payInvoice)

		admin := invoice.Group("/admin")
		{
			admin.GET("", h.listInvoices)
		}
	}
}

// @Summary Create a new invoice
// @Description Creates a new invoice for payment
// @Tags invoices
// @Accept json
// @Produce json
// @Param input body models.InvoiceRequest true "Invoice details"
// @Success 201 {object} models.Invoice
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/invoice [post]
func (h *InvoiceHandler) createInvoice(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request models.InvoiceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logrus.WithError(err).Error("invoiceHandler.createInvoice: invalid input body")
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.RecipientName == "" || request.RecipientAccount == "" {
		newErrorResponse(c, http.StatusBadRequest, "recipient name and account are required")
		return
	}

	if request.Currency == "" {
		newErrorResponse(c, http.StatusBadRequest, "currency is required")
		return
	}

	if request.Description == "" {
		newErrorResponse(c, http.StatusBadRequest, "description is required")
		return
	}

	if request.DueDate.IsZero() {
		newErrorResponse(c, http.StatusBadRequest, "due date is required")
		return
	}

	if request.DueDate.Before(time.Now()) {
		newErrorResponse(c, http.StatusBadRequest, "due date must be in the future")
		return
	}

	if len(request.Items) == 0 {
		newErrorResponse(c, http.StatusBadRequest, "at least one invoice item is required")
		return
	}

	invoice, err := h.service.CreateInvoice(c.Request.Context(), userID, request)
	if err != nil {
		logrus.WithError(err).Error("invoiceHandler.createInvoice: error creating invoice")
		newErrorResponse(c, http.StatusInternalServerError, "error creating invoice")
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

// @Summary Get an invoice by ID
// @Description Retrieves an invoice by its ID
// @Tags invoices
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} models.Invoice
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/invoice/{id} [get]
func (h *InvoiceHandler) getInvoice(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	invoice, err := h.service.GetInvoice(c.Request.Context(), id, userID)
	if err != nil {
		logrus.WithError(err).Errorf("invoiceHandler.getInvoice: error getting invoice %d", id)
		if errors.Is(err, errors.New("access denied")) {
			newErrorResponse(c, http.StatusForbidden, "you don't have access to this invoice")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error getting invoice")
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// @Summary Get an invoice by number
// @Description Retrieves an invoice by its invoice number
// @Tags invoices
// @Produce json
// @Param invoiceNumber path string true "Invoice Number"
// @Success 200 {object} models.Invoice
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/invoice/number/{invoiceNumber} [get]
func (h *InvoiceHandler) getInvoiceByNumber(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	invoiceNumber := c.Param("invoiceNumber")
	if invoiceNumber == "" {
		newErrorResponse(c, http.StatusBadRequest, "invoice number is required")
		return
	}

	invoice, err := h.service.GetInvoiceByNumber(c.Request.Context(), invoiceNumber)
	if err != nil {
		logrus.WithError(err).Errorf("invoiceHandler.getInvoiceByNumber: error getting invoice %s", invoiceNumber)
		newErrorResponse(c, http.StatusNotFound, "invoice not found")
		return
	}

	if invoice.UserID != userID && invoice.RecipientID != userID {
		newErrorResponse(c, http.StatusForbidden, "you don't have access to this invoice")
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// @Summary Cancel an invoice
// @Description Cancels an existing invoice (only allowed for the creator and only if the invoice is pending)
// @Tags invoices
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/invoice/{id} [delete]
func (h *InvoiceHandler) cancelInvoice(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	err = h.service.CancelInvoice(c.Request.Context(), id, userID)
	if err != nil {
		logrus.WithError(err).Errorf("invoiceHandler.cancelInvoice: error cancelling invoice %d", id)
		if errors.Is(err, errors.New("access denied")) {
			newErrorResponse(c, http.StatusForbidden, "only the invoice creator can cancel it")
			return
		}
		if errors.Is(err, fmt.Errorf("cannot cancel invoice with status %s", models.InvoiceStatusPaid)) {
			newErrorResponse(c, http.StatusBadRequest, "cannot cancel paid invoice")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error cancelling invoice")
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "success", Message: "invoice cancelled successfully"})
}

// @Summary Get user's invoices
// @Description Retrieves all invoices for the authenticated user
// @Tags invoices
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} paginatedResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/invoice [get]
func (h *InvoiceHandler) getUserInvoices(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page := int64(1)
	if pageStr := c.Query("page"); pageStr != "" {
		pageVal, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil || pageVal < 1 {
			newErrorResponse(c, http.StatusBadRequest, "invalid page param")
			return
		}
		page = pageVal
	}

	pageSize := int64(10)
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSizeVal, err := strconv.ParseInt(pageSizeStr, 10, 64)
		if err != nil || pageSizeVal < 1 || pageSizeVal > 100 {
			newErrorResponse(c, http.StatusBadRequest, "invalid page_size param")
			return
		}
		pageSize = pageSizeVal
	}

	invoices, totalCount, err := h.service.GetUserInvoices(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logrus.WithError(err).Errorf("invoiceHandler.getUserInvoices: error getting invoices for user %d", userID)
		newErrorResponse(c, http.StatusInternalServerError, "error getting invoices")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       invoices,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

// @Summary List invoices (Admin only)
// @Description Lists all invoices with filtering and pagination (admin only)
// @Tags invoices
// @Produce json
// @Param user_id query int false "Filter by user ID"
// @Param recipient_id query int false "Filter by recipient ID"
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
// @Router /api/v1/invoice/admin [get]
func (h *InvoiceHandler) listInvoices(c *gin.Context) {
	if !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "admin access required")
		return
	}

	var filter models.InvoiceFilter

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid user_id param")
			return
		}
		filter.UserID = userID
	}

	if recipientIDStr := c.Query("recipient_id"); recipientIDStr != "" {
		recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid recipient_id param")
			return
		}
		filter.RecipientID = recipientID
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

	invoices, totalCount, err := h.service.ListInvoices(c.Request.Context(), filter)
	if err != nil {
		logrus.WithError(err).Error("invoiceHandler.listInvoices: error listing invoices")
		newErrorResponse(c, http.StatusInternalServerError, "error listing invoices")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       invoices,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	})
}

// @Summary Pay an invoice
// @Description Pay an invoice using a card or account
// @Tags invoices
// @Accept json
// @Produce json
// @Param input body models.InvoicePaymentRequest true "Payment details"
// @Success 200 {object} models.Invoice
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/invoice/pay [post]
func (h *InvoiceHandler) payInvoice(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request models.InvoicePaymentRequest
	if err := c.BindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.InvoiceID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invoice ID is required")
		return
	}

	if request.PaymentMethod != models.PaymentMethodCard && request.PaymentMethod != models.PaymentMethodAccount {
		newErrorResponse(c, http.StatusBadRequest, "payment method must be CARD or ACCOUNT")
		return
	}

	if request.PaymentMethod == models.PaymentMethodCard && request.CardID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "card ID is required for card payment")
		return
	}

	if request.PaymentMethod == models.PaymentMethodAccount && request.AccountID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "account ID is required for account payment")
		return
	}

	invoice, err := h.service.PayInvoice(c.Request.Context(), userID, request)
	if err != nil {
		logrus.WithError(err).Errorf("invoiceHandler.payInvoice: error paying invoice %d", request.InvoiceID)
		if errors.Is(err, errors.New("invoice is already paid")) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, errors.New("invoice is cancelled and cannot be paid")) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, errors.New("invoice is expired and cannot be paid")) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, errors.New("insufficient funds in account")) ||
			errors.Is(err, errors.New("insufficient funds in account linked to card")) {
			newErrorResponse(c, http.StatusBadRequest, "insufficient funds")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error processing payment")
		return
	}

	c.JSON(http.StatusOK, invoice)
}
