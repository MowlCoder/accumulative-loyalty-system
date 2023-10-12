package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/httputils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/jsonutil"
)

type goodRewardsService interface {
	SaveNewGoodReward(ctx context.Context, match string, reward float64, rewardType string) (*domain.GoodReward, error)
}

type GoodsHandler struct {
	goodRewardsService goodRewardsService
}

func NewGoodsHandler(goodRewardsService goodRewardsService) *GoodsHandler {
	return &GoodsHandler{
		goodRewardsService: goodRewardsService,
	}
}

type saveNewGoodRewardBody struct {
	Match      string  `json:"match"`
	Reward     float64 `json:"reward"`
	RewardType string  `json:"reward_type"`
}

func (b *saveNewGoodRewardBody) Valid() bool {
	if len(b.Match) == 0 || b.Reward <= 0 || !domain.IsValidRewardType(b.RewardType) {
		return false
	}

	return true
}

// SaveNewGoodReward godoc
// @Summary Save new good reward
// @Tags goods
// @Accept json
// @Produce json
// @Param dto body saveNewGoodRewardBody true "Add new Good Reward"
// @Success 200 {object} domain.GoodReward
// @Failure 400 {object} httputils.HTTPError
// @Failure 409 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Router /goods [post]
func (h *GoodsHandler) SaveNewGoodReward(w http.ResponseWriter, r *http.Request) {
	var body saveNewGoodRewardBody

	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	if !body.Valid() {
		httputils.SendJSONErrorResponse(w, http.StatusBadRequest, "invalid body")
		return
	}

	reward, err := h.goodRewardsService.SaveNewGoodReward(r.Context(), body.Match, body.Reward, body.RewardType)

	if err != nil {
		if errors.Is(err, domain.ErrMatchKeyAlreadyExists) {
			httputils.SendJSONErrorResponse(w, http.StatusConflict, err.Error())
			return
		}

		log.Println("[SaveNewGoodReward]", err)
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, reward)
}
