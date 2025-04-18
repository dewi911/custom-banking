package rest

import (
	"custom-banking/internal/models"
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Card struct {
	cardService CardService
}

func NewCard(cardService CardService) *Card {
	return &Card{cardService}
}

func (t *Card) InjectRoutes(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	accountCard := r.Group("/account/:id/card").Use(middlewares...)
	{
		accountCard.POST("/", t.createCard)
		accountCard.GET("/", t.getCardListByAccount)
		accountCard.GET("/:card_id", t.getCard)
	}

	card := r.Group("/card").Use(middlewares...)
	{
		card.GET("/", t.GetCardListByUser)
	}
}

func (t *Card) createCard(ctx *gin.Context) {
	accountID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong \"id\" request param", err))
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting current user error", err))
		return
	}

	var req models.CreateCardRequestBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("request body validation error", err))
		return
	}

	domainCard, err := t.cardService.CreateCard(ctx, accountID, userID, req.CashbackPercentage, req.CardType)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("card creation error", err))
		return
	}

	messageCard := models.Card{
		Id:                 domainCard.Id,
		AccountId:          domainCard.AccountId,
		CardNumber:         domainCard.CardNumber,
		CardholderName:     domainCard.CardholderName,
		ExpirationDate:     domainCard.ExpirationDate,
		CvvCode:            domainCard.CvvCode,
		CardType:           domainCard.CardType,
		CashbackPercentage: domainCard.CashbackPercentage,
	}

	ctx.JSON(http.StatusCreated, messageCard)
}

func (t *Card) GetCardListByUser(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting current user error", err))
		return
	}

	domainListCards, err := t.cardService.GetCardListUser(ctx, userID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting cards list error", err))
		return
	}

	messageListCards := make([]models.Card, 0)
	for _, card := range domainListCards {
		messageCard := models.Card{
			Id:                 card.Id,
			AccountId:          card.AccountId,
			CardNumber:         card.CardNumber,
			CardholderName:     card.CardholderName,
			ExpirationDate:     card.ExpirationDate,
			CvvCode:            card.CvvCode,
			CardType:           card.CardType,
			CashbackPercentage: card.CashbackPercentage,
		}
		messageListCards = append(messageListCards, messageCard)
	}

	ctx.JSON(http.StatusOK, messageListCards)
}

func (t *Card) getCardListByAccount(ctx *gin.Context) {
	accountID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong \"id\" request param", err))
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting current user error", err))
		return
	}

	domainListCards, err := t.cardService.GetCardListByAccount(ctx, userID, accountID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting cards list for account error", err))
		return
	}

	messageListCards := make([]models.Card, 0)
	for _, card := range domainListCards {
		messageCard := models.Card{
			Id:                 card.Id,
			AccountId:          card.AccountId,
			CardNumber:         card.CardNumber,
			CardholderName:     card.CardholderName,
			ExpirationDate:     card.ExpirationDate,
			CvvCode:            card.CvvCode,
			CardType:           card.CardType,
			CashbackPercentage: card.CashbackPercentage,
		}
		messageListCards = append(messageListCards, messageCard)
	}

	ctx.JSON(http.StatusOK, messageListCards)
}

func (t *Card) getCard(ctx *gin.Context) {
	accountID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong \"id\" request param", err))
		return
	}

	cardID, err := strconv.Atoi(ctx.Param("card_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("wrong \"card_id\" request param", err))
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting current user error", err))
		return
	}

	domainCard, err := t.cardService.GetCard(ctx, cardID, accountID, userID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			ctx.AbortWithStatusJSON(http.StatusNotFound, NewNotFoundError("card not found", err))
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting card error", err))
		}

		return
	}

	messageCard := models.Card{
		Id:                 domainCard.Id,
		AccountId:          domainCard.AccountId,
		CardNumber:         domainCard.CardNumber,
		CardholderName:     domainCard.CardholderName,
		ExpirationDate:     domainCard.ExpirationDate,
		CvvCode:            domainCard.CvvCode,
		CardType:           domainCard.CardType,
		CashbackPercentage: domainCard.CashbackPercentage,
	}

	ctx.JSON(http.StatusOK, messageCard)
}
