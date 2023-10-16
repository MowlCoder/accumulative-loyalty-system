package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	servicemock "github.com/MowlCoder/accumulative-loyalty-system/internal/handlers/mocks"
)

func TestGoodsHandler_SaveNewGoodReward(t *testing.T) {
	ctrl := gomock.NewController(t)
	goodsRewardsService := servicemock.NewMockgoodRewardsService(ctrl)
	goodsHandler := NewGoodsHandler(goodsRewardsService)

	type TestCase struct {
		Name               string
		Body               *saveNewGoodRewardBody
		PrepareServiceFunc func(ctx context.Context, service *servicemock.MockgoodRewardsService, body *saveNewGoodRewardBody)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: &saveNewGoodRewardBody{
				Match:      "Bork",
				Reward:     10.0,
				RewardType: domain.PercentRewardType,
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockgoodRewardsService, body *saveNewGoodRewardBody) {
				service.
					EXPECT().
					SaveNewGoodReward(ctx, body.Match, body.Reward, body.RewardType).
					Return(&domain.GoodReward{Match: body.Match, Reward: body.Reward, RewardType: body.RewardType}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "invalid (invalid body)",
			Body:               &saveNewGoodRewardBody{},
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "invalid (invalid json)",
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name: "invalid (match key already exists)",
			Body: &saveNewGoodRewardBody{
				Match:      "Bork",
				Reward:     10.0,
				RewardType: domain.PercentRewardType,
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockgoodRewardsService, body *saveNewGoodRewardBody) {
				service.
					EXPECT().
					SaveNewGoodReward(ctx, body.Match, body.Reward, body.RewardType).
					Return(nil, domain.ErrMatchKeyAlreadyExists)
			},
			ExpectedStatusCode: http.StatusConflict,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var rawBody []byte
			var err error

			if testCase.Body != nil {
				rawBody, err = json.Marshal(*testCase.Body)
				require.NoError(t, err)
			}

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(rawBody))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), goodsRewardsService, testCase.Body)
			}

			goodsHandler.SaveNewGoodReward(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}
