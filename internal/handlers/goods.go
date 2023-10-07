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

func (h *GoodsHandler) SaveNewGoodReward(w http.ResponseWriter, r *http.Request) {
	var body saveNewGoodRewardBody

	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
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
