package rest

import (
	"custom-banking/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type LoanHandler struct {
	service LoanService
}

func NewLoanHandler(service LoanService) *LoanHandler {
	return &LoanHandler{
		service: service,
	}
}

func (h *LoanHandler) Register(api *gin.RouterGroup) {
	loans := api.Group("/loans")
	{
		loans.POST("", h.CreateLoan)
		loans.GET("/:id", h.GetLoanByID)
		loans.GET("/user/:user_id", h.GetLoansByUserID)
		loans.GET("", h.ListLoans)
		loans.PATCH("/:id/status", h.UpdateLoanStatus)
		loans.POST("/:id/payments", h.MakePayment)
		loans.GET("/:id/payments", h.GetPaymentsByLoanID)
	}
}

// CreateLoan godoc
// @Summary Create a new loan
// @Description Create a new loan application
// @Tags loans
// @Accept json
// @Produce json
// @Param request body models.LoanRequest true "Loan details"
// @Success 201 {object} models.Loan
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/loans [post]
func (h *LoanHandler) CreateLoan(c *gin.Context) {
	var request models.LoanRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.Amount <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "loan amount must be positive")
		return
	}

	if request.UserID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "user ID is required")
		return
	}

	if request.CurrencyID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "currency ID is required")
		return
	}

	if request.MonthsDuration <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "loan duration must be positive")
		return
	}

	loan, err := h.service.Create(request)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.CreateLoan: error creating loan")
		newErrorResponse(c, http.StatusInternalServerError, "failed to create loan")
		return
	}

	c.JSON(http.StatusCreated, loan)
}

// GetLoanByID godoc
// @Summary Get loan by ID
// @Description Get detailed information about a loan by its ID
// @Tags loans
// @Accept json
// @Produce json
// @Param id path int true "Loan ID"
// @Success 200 {object} models.Loan
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/loans/{id} [get]
func (h *LoanHandler) GetLoanByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid loan ID")
		return
	}

	loan, err := h.service.GetByID(id)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.GetLoanByID: error getting loan")
		newErrorResponse(c, http.StatusNotFound, "loan not found")
		return
	}

	c.JSON(http.StatusOK, loan)
}

// GetLoansByUserID godoc
// @Summary Get user loans
// @Description Get all loans for a specific user
// @Tags loans
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Loan
// @Failure 500 {object} errorResponse
// @Router /api/v1/loans/user/{user_id} [get]
func (h *LoanHandler) GetLoansByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	loans, err := h.service.GetByUserID(userID)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.GetLoansByUserID: error getting loans")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve loans")
		return
	}

	c.JSON(http.StatusOK, loans)
}

// ListLoans godoc
// @Summary List loans
// @Description Get a list of loans with filtering and pagination
// @Tags loans
// @Accept json
// @Produce json
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by loan status"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} paginatedResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/loans [get]
func (h *LoanHandler) ListLoans(c *gin.Context) {
	params := models.LoanListParams{
		Page:     1,
		PageSize: 10,
	}

	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			params.UserID = userID
		}
	}

	params.Status = c.Query("status")

	page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if err == nil && page > 0 {
		params.Page = page
	}

	pageSize, err := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)
	if err == nil && pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}

	loans, totalCount, err := h.service.List(params)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.ListLoans: error listing loans")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve loans")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       loans,
		TotalCount: int64(totalCount),
		Page:       params.Page,
		PageSize:   params.PageSize,
	})
}

// UpdateLoanStatus godoc
// @Summary Update loan status
// @Description Update the status of a loan
// @Tags loans
// @Accept json
// @Produce json
// @Param id path int true "Loan ID"
// @Param request body statusRequest true "Status details"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/loans/{id}/status [patch]
func (h *LoanHandler) UpdateLoanStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid loan ID")
		return
	}

	var request statusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.Status == "" {
		newErrorResponse(c, http.StatusBadRequest, "status is required")
		return
	}

	// First check if loan exists
	_, err = h.service.GetByID(id)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.UpdateLoanStatus: error getting loan")
		newErrorResponse(c, http.StatusNotFound, "loan not found")
		return
	}

	// Update loan status
	err = h.service.UpdateStatus(id, request.Status)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.UpdateLoanStatus: error updating loan status")
		newErrorResponse(c, http.StatusInternalServerError, "failed to update loan status")
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "success",
	})
}

// MakePayment godoc
// @Summary Make loan payment
// @Description Make a payment for a loan
// @Tags loans
// @Accept json
// @Produce json
// @Param id path int true "Loan ID"
// @Param request body models.LoanPaymentRequest true "Payment details"
// @Success 200 {object} models.LoanPayment
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/loans/{id}/payments [post]
func (h *LoanHandler) MakePayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid loan ID")
		return
	}

	var request models.LoanPaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	// Set loan ID from path
	request.LoanID = id

	// Validate required fields
	if request.Amount <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "payment amount must be positive")
		return
	}

	if request.AccountID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "account ID is required")
		return
	}

	// Process payment
	payment, err := h.service.MakePayment(request)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.MakePayment: error making payment")
		if err.Error() == "loan not found" {
			newErrorResponse(c, http.StatusNotFound, err.Error())
		} else {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPaymentsByLoanID godoc
// @Summary Get loan payments
// @Description Get payment history for a loan
// @Tags loans
// @Accept json
// @Produce json
// @Param id path int true "Loan ID"
// @Success 200 {array} models.LoanPayment
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/loans/{id}/payments [get]
func (h *LoanHandler) GetPaymentsByLoanID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid loan ID")
		return
	}

	// First check if loan exists
	_, err = h.service.GetByID(id)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.GetPaymentsByLoanID: error getting loan")
		newErrorResponse(c, http.StatusNotFound, "loan not found")
		return
	}

	payments, err := h.service.GetPaymentsByLoanID(id)
	if err != nil {
		logrus.WithError(err).Error("LoanHandler.GetPaymentsByLoanID: error getting payments")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve payments")
		return
	}

	c.JSON(http.StatusOK, payments)
}
