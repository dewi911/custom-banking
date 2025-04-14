package rest

import (
	"custom-banking/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type StakingHandler struct {
	service StakingService
}

func NewStakingHandler(service StakingService) *StakingHandler {
	return &StakingHandler{
		service: service,
	}
}

func (h *StakingHandler) Register(api *gin.Engine, middlewares ...gin.HandlerFunc) {
	staking := api.Group("/staking").Use(middlewares...)
	{
		staking.POST("", h.CreateStaking)
		staking.GET("/:id", h.GetStakingByID)
		staking.GET("/user/:user_id", h.GetStakingsByUserID)
		staking.GET("", h.ListStakings)
		staking.POST("/:id/withdraw", h.WithdrawStaking)
		staking.GET("/:id/interest", h.GetStakingInterest)
		staking.GET("/:id/interests", h.GetStakingInterests)
		staking.POST("/calculate", h.CalculateProjectedInterest)
	}
}

// CreateStaking godoc
// @Summary Create a new staking
// @Description Create a new staking deposit
// @Tags staking
// @Accept json
// @Produce json
// @Param request body models.StakingRequest true "Staking details"
// @Success 201 {object} models.Staking
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/staking [post]
func (h *StakingHandler) CreateStaking(c *gin.Context) {
	var request models.StakingRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	request.UserID = int64(userID)

	staking, err := h.service.Create(c, request)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.CreateStaking: error creating staking")
		if err.Error() == "insufficient funds in account" {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "failed to create staking")
		return
	}

	c.JSON(http.StatusCreated, staking)
}

// GetStakingByID godoc
// @Summary Get staking by ID
// @Description Get detailed information about a staking by its ID
// @Tags staking
// @Accept json
// @Produce json
// @Param id path int true "Staking ID"
// @Success 200 {object} models.Staking
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/staking/{id} [get]
func (h *StakingHandler) GetStakingByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid staking ID")
		return
	}

	staking, err := h.service.GetByID(c, id)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.GetStakingByID: error getting staking")
		newErrorResponse(c, http.StatusNotFound, "staking not found")
		return
	}

	c.JSON(http.StatusOK, staking)
}

// GetStakingsByUserID godoc
// @Summary Get user stakings
// @Description Get all stakings for a specific user
// @Tags staking
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Staking
// @Failure 500 {object} errorResponse
// @Router /api/v1/staking/user/{user_id} [get]
func (h *StakingHandler) GetStakingsByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	stakings, err := h.service.GetByUserID(c, userID)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.GetStakingsByUserID: error getting stakings")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve stakings")
		return
	}

	c.JSON(http.StatusOK, stakings)
}

// ListStakings godoc
// @Summary List stakings
// @Description Get a list of stakings with filtering and pagination
// @Tags staking
// @Accept json
// @Produce json
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by staking status"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} paginatedResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/staking [get]
func (h *StakingHandler) ListStakings(c *gin.Context) {
	params := models.StakingListParams{
		Page:     1,
		PageSize: 30,
	}

	userIDStr := c.Query("user-id")
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

	stakings, totalCount, err := h.service.List(c, params)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.ListStakings: error listing stakings")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve stakings")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       stakings,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	})
}

// WithdrawStaking godoc
// @Summary Withdraw staking
// @Description Withdraw funds from a staking deposit
// @Tags staking
// @Accept json
// @Produce json
// @Param id path int true "Staking ID"
// @Param request body models.StakingWithdrawRequest true "Withdrawal details"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/staking/{id}/withdraw [post]
func (h *StakingHandler) WithdrawStaking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid staking ID")
		return
	}

	var request models.StakingWithdrawRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	request.StakingID = id

	if request.AccountID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "account ID is required")
		return
	}

	err = h.service.Withdraw(c, request)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.WithdrawStaking: error withdrawing staking")

		if err.Error() == "staking not found" {
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "account not found" {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		if err.Error() == "cannot withdraw from a staking with status" {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, "failed to withdraw staking")
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status:  "success",
		Message: "Staking withdrawn successfully",
	})
}

// GetStakingInterest godoc
// @Summary Get staking interest
// @Description Get the total earned interest for a staking
// @Tags staking
// @Accept json
// @Produce json
// @Param id path int true "Staking ID"
// @Success 200 {object} interestResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/staking/{id}/interest [get]
func (h *StakingHandler) GetStakingInterest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid staking ID")
		return
	}

	_, err = h.service.GetByID(c, id)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.GetStakingInterest: staking not found")
		newErrorResponse(c, http.StatusNotFound, "staking not found")
		return
	}

	interest, err := h.service.GetEarnedInterest(c, id)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.GetStakingInterest: error getting interest")
		newErrorResponse(c, http.StatusInternalServerError, "failed to calculate interest")
		return
	}

	c.JSON(http.StatusOK, interestResponse{
		StakingID: id,
		Interest:  interest,
	})
}

// GetStakingInterests godoc
// @Summary Get staking interest history
// @Description Get all interest records for a staking
// @Tags staking
// @Accept json
// @Produce json
// @Param id path int true "Staking ID"
// @Success 200 {array} models.StakingInterest
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/staking/{id}/interests [get]
func (h *StakingHandler) GetStakingInterests(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid staking ID")
		return
	}

	_, err = h.service.GetByID(c, id)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.GetStakingInterests: staking not found")
		newErrorResponse(c, http.StatusNotFound, "staking not found")
		return
	}

	interests, err := h.service.GetInterestsByStakingID(c, id)
	if err != nil {
		logrus.WithError(err).Error("StakingHandler.GetStakingInterests: error getting interests")
		newErrorResponse(c, http.StatusInternalServerError, "failed to retrieve interest history")
		return
	}

	c.JSON(http.StatusOK, interests)
}

// CalculateProjectedInterest godoc
// @Summary Calculate projected interest
// @Description Calculate projected interest for a potential staking
// @Tags staking
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Calculation parameters"
// @Success 200 {object} map[string]float64
// @Failure 400 {object} errorResponse
// @Router /api/v1/staking/calculate [post]
func (h *StakingHandler) CalculateProjectedInterest(c *gin.Context) {
	var request struct {
		Amount       float64 `json:"amount" binding:"required"`
		Days         int64   `json:"days" binding:"required"`
		InterestRate float64 `json:"interest_rate,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.Amount <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "amount must be positive")
		return
	}

	if request.Days <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "days must be positive")
		return
	}

	if request.InterestRate <= 0 {
		switch {
		case request.Days <= 30:
			request.InterestRate = 3.0 // 3% for 1-30 days
		case request.Days <= 90:
			request.InterestRate = 5.0 // 5% for 31-90 days
		case request.Days <= 180:
			request.InterestRate = 7.0 // 7% for 91-180 days
		default:
			request.InterestRate = 10.0 // 10% for 180+ days
		}
	}

	interest := h.service.CalculateProjectedInterest(c, request.Amount, request.Days, request.InterestRate)

	c.JSON(http.StatusOK, map[string]float64{
		"interest": interest,
	})
}
