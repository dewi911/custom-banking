package rest

import (
	"custom-banking/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type InsuranceHandler struct {
	service InsuranceService
}

func NewInsuranceHandler(service InsuranceService) *InsuranceHandler {
	return &InsuranceHandler{service: service}
}

func (h *InsuranceHandler) InitRoutes(api *gin.RouterGroup) {
	insurance := api.Group("/insurance")
	{
		insurance.POST("", h.create)
		insurance.GET("", h.list)
		insurance.GET("/:id", h.getByID)
		insurance.GET("/user", h.getByUserID)
		insurance.PATCH("/:id/status", h.updateStatus)

		insurance.POST("/:id/claims", h.createClaim)
		insurance.GET("/:id/claims", h.getClaimsByInsuranceID)
		insurance.GET("/claims/:claimId", h.getClaimByID)
		insurance.PATCH("/claims/:claimId/status", h.updateClaimStatus)
		insurance.GET("/claims", h.listClaims)
	}
}

// @Summary Create a new insurance policy
// @Description Creates a new insurance policy for the authenticated user
// @Tags insurance
// @Accept json
// @Produce json
// @Param input body models.InsuranceRequest true "Insurance request info"
// @Success 201 {object} models.Insurance
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance [post]
func (h *InsuranceHandler) create(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request models.InsuranceRequest
	if err := c.BindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.Type == "" || request.InsuredItem == "" || request.CoverageAmount <= 0 || request.Premium <= 0 || request.DurationDays <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "all fields are required and must be valid")
		return
	}

	request.UserID = userID

	insurance, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		logrus.WithError(err).Error("insuranceHandler.create: error creating insurance")
		if errors.Is(err, errors.New("insufficient funds in payment account")) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error creating insurance")
		return
	}

	c.JSON(http.StatusCreated, insurance)
}

// @Summary Get an insurance policy by ID
// @Description Retrieves an insurance policy by its ID
// @Tags insurance
// @Produce json
// @Param id path int true "Insurance ID"
// @Success 200 {object} models.Insurance
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/{id} [get]
func (h *InsuranceHandler) getByID(c *gin.Context) {
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

	insurance, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.getByID: error getting insurance with id %d", id)
		newErrorResponse(c, http.StatusNotFound, "insurance not found")
		return
	}

	if insurance.UserID != userID && !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "forbidden")
		return
	}

	c.JSON(http.StatusOK, insurance)
}

// @Summary Get all insurance policies for the authenticated user
// @Description Retrieves all insurance policies for the authenticated user
// @Tags insurance
// @Produce json
// @Success 200 {array} models.Insurance
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/user [get]
func (h *InsuranceHandler) getByUserID(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	insurances, err := h.service.GetByUserID(c.Request.Context(), int64(userID))
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.getByUserID: error getting insurances for user %d", userID)
		newErrorResponse(c, http.StatusInternalServerError, "error getting insurances")
		return
	}

	c.JSON(http.StatusOK, insurances)
}

// @Summary Update the status of an insurance policy
// @Description Updates the status of an insurance policy (admin only)
// @Tags insurance
// @Accept json
// @Produce json
// @Param id path int true "Insurance ID"
// @Param input body map[string]string true "Status update"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/{id}/status [patch]
func (h *InsuranceHandler) updateStatus(c *gin.Context) {
	if !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "admin access required")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	validStatuses := []string{
		models.InsuranceStatusActive,
		models.InsuranceStatusCancelled,
		models.InsuranceStatusExpired,
	}

	isValid := false
	for _, status := range validStatuses {
		if input.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		newErrorResponse(c, http.StatusBadRequest, "invalid status")
		return
	}

	err = h.service.UpdateStatus(c.Request.Context(), id, input.Status)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.updateStatus: error updating status for insurance %d", id)
		if errors.Is(err, errors.New("insurance not found")) {
			newErrorResponse(c, http.StatusNotFound, "insurance not found")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error updating insurance status")
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "success", Message: "status updated successfully"})
}

// @Summary List insurance policies with filtering
// @Description Lists insurance policies with filtering and pagination
// @Tags insurance
// @Produce json
// @Param user_id query int false "Filter by user ID (admin only)"
// @Param type query string false "Filter by insurance type"
// @Param status query string false "Filter by status"
// @Param insured_item query string false "Filter by insured item (partial match)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} paginatedResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance [get]
func (h *InsuranceHandler) list(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params models.InsuranceListParams

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if !isAdmin(c) {
			newErrorResponse(c, http.StatusForbidden, "admin access required to filter by user_id")
			return
		}
		userIDParam, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid user_id param")
			return
		}
		params.UserID = userIDParam
	} else {
		if !isAdmin(c) {
			params.UserID = userID
		}
	}

	params.Type = c.Query("type")

	params.Status = c.Query("status")

	params.InsuredItem = c.Query("insured_item")

	params.Page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil || page < 1 {
			newErrorResponse(c, http.StatusBadRequest, "invalid page param")
			return
		}
		params.Page = page
	}

	params.PageSize = 10
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
		if err != nil || pageSize < 1 || pageSize > 100 {
			newErrorResponse(c, http.StatusBadRequest, "invalid page_size param")
			return
		}
		params.PageSize = pageSize
	}

	insurances, totalCount, err := h.service.List(c.Request.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("insuranceHandler.list: error listing insurances")
		newErrorResponse(c, http.StatusInternalServerError, "error listing insurances")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       insurances,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	})
}

// @Summary Create a new insurance claim
// @Description Creates a new claim for an existing insurance policy
// @Tags insurance-claims
// @Accept json
// @Produce json
// @Param id path int true "Insurance ID"
// @Param input body models.InsuranceClaimRequest true "Claim request info"
// @Success 201 {object} models.InsuranceClaim
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/{id}/claims [post]
func (h *InsuranceHandler) createClaim(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	insuranceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid insurance id param")
		return
	}

	insurance, err := h.service.GetByID(c.Request.Context(), insuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.createClaim: error getting insurance with id %d", insuranceID)
		newErrorResponse(c, http.StatusNotFound, "insurance not found")
		return
	}

	if insurance.UserID != userID {
		newErrorResponse(c, http.StatusForbidden, "you can only create claims for your own insurance policies")
		return
	}

	var request models.InsuranceClaimRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logrus.WithError(err).Error("insuranceHandler.createClaim: invalid input body")
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if request.Description == "" || request.Amount <= 0 || request.ClaimDate.IsZero() {
		newErrorResponse(c, http.StatusBadRequest, "description, amount, and claim date are required")
		return
	}

	request.InsuranceID = insuranceID

	claim, err := h.service.CreateClaim(c.Request.Context(), request)
	if err != nil {
		logrus.WithError(err).Error("insuranceHandler.createClaim: error creating claim")
		if errors.Is(err, errors.New("claim date must be within insurance period")) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, errors.New("cannot file claim for insurance with status")) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error creating claim")
		return
	}

	c.JSON(http.StatusCreated, claim)
}

// @Summary Get a claim by ID
// @Description Retrieves a claim by its ID
// @Tags insurance-claims
// @Produce json
// @Param claimId path int true "Claim ID"
// @Success 200 {object} models.InsuranceClaim
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/claims/{claimId} [get]
func (h *InsuranceHandler) getClaimByID(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	claimID, err := strconv.ParseInt(c.Param("claimId"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid claim id param")
		return
	}

	claim, err := h.service.GetClaimByID(c.Request.Context(), claimID)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.getClaimByID: error getting claim with id %d", claimID)
		newErrorResponse(c, http.StatusNotFound, "claim not found")
		return
	}

	insurance, err := h.service.GetByID(c.Request.Context(), claim.InsuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.getClaimByID: error getting insurance with id %d", claim.InsuranceID)
		newErrorResponse(c, http.StatusInternalServerError, "error getting insurance details")
		return
	}

	if insurance.UserID != userID && !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "forbidden")
		return
	}

	c.JSON(http.StatusOK, claim)
}

// @Summary Get all claims for an insurance policy
// @Description Retrieves all claims for an insurance policy
// @Tags insurance-claims
// @Produce json
// @Param id path int true "Insurance ID"
// @Success 200 {array} models.InsuranceClaim
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/{id}/claims [get]
func (h *InsuranceHandler) getClaimsByInsuranceID(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	insuranceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid insurance id param")
		return
	}

	insurance, err := h.service.GetByID(c.Request.Context(), insuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.getClaimsByInsuranceID: error getting insurance with id %d", insuranceID)
		newErrorResponse(c, http.StatusNotFound, "insurance not found")
		return
	}

	if insurance.UserID != userID && !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "forbidden")
		return
	}

	claims, err := h.service.GetClaimsByInsuranceID(c.Request.Context(), insuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.getClaimsByInsuranceID: error getting claims for insurance %d", insuranceID)
		newErrorResponse(c, http.StatusInternalServerError, "error getting claims")
		return
	}

	c.JSON(http.StatusOK, claims)
}

// @Summary Update the status of a claim
// @Description Updates the status of an insurance claim
// @Tags insurance-claims
// @Accept json
// @Produce json
// @Param claimId path int true "Claim ID"
// @Param input body map[string]string true "Status update"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/claims/{claimId}/status [patch]
func (h *InsuranceHandler) updateClaimStatus(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	claimID, err := strconv.ParseInt(c.Param("claimId"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid claim id param")
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	validStatuses := []string{
		models.ClaimStatusApproved,
		models.ClaimStatusRejected,
		models.ClaimStatusCancelled,
	}

	isValid := false
	for _, status := range validStatuses {
		if input.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		newErrorResponse(c, http.StatusBadRequest, "invalid status")
		return
	}

	if (input.Status == models.ClaimStatusApproved || input.Status == models.ClaimStatusRejected) && !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "admin access required to approve or reject claims")
		return
	}

	err = h.service.UpdateClaimStatus(c.Request.Context(), claimID, input.Status, userID)
	if err != nil {
		logrus.WithError(err).Errorf("insuranceHandler.updateClaimStatus: error updating status for claim %d", claimID)
		if errors.Is(err, errors.New("claim not found")) {
			newErrorResponse(c, http.StatusNotFound, "claim not found")
			return
		}
		if errors.Is(err, errors.New("only administrators can approve or reject claims")) ||
			errors.Is(err, errors.New("only claim owner can cancel claims")) {
			newErrorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "error updating claim status")
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "success", Message: "claim status updated successfully"})
}

// @Summary List all claims with filtering
// @Description Lists all claims with filtering and pagination (admin only)
// @Tags insurance-claims
// @Produce json
// @Param insurance_id query int false "Filter by insurance ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} paginatedResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/insurance/claims [get]
func (h *InsuranceHandler) listClaims(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var insuranceID int64
	if insuranceIDStr := c.Query("insurance_id"); insuranceIDStr != "" {
		var err error
		insuranceID, err = strconv.ParseInt(insuranceIDStr, 10, 64)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid insurance_id param")
			return
		}

		insurance, err := h.service.GetByID(c.Request.Context(), insuranceID)
		if err != nil {
			newErrorResponse(c, http.StatusNotFound, "insurance not found")
			return
		}

		if insurance.UserID != userID && !isAdmin(c) {
			newErrorResponse(c, http.StatusForbidden, "forbidden")
			return
		}
	} else if !isAdmin(c) {
		newErrorResponse(c, http.StatusForbidden, "must specify insurance_id or be an admin")
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

	claims, totalCount, err := h.service.ListClaims(c.Request.Context(), insuranceID, page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("insuranceHandler.listClaims: error listing claims")
		newErrorResponse(c, http.StatusInternalServerError, "error listing claims")
		return
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       claims,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}
