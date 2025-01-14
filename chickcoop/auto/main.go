package main

import (
	"chickcoop/constant"
	"chickcoop/request"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

var logger *zap.Logger

func main() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ = config.Build()

	viper.SetConfigFile("./conf.toml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	tokens := viper.GetStringSlice("auth.tokens")
	proxies := viper.GetStringSlice("proxies.data")
	for i, token := range tokens {
		proxy := proxies[i%len(proxies)]
		go process(token, proxy, i)

	}
	select {}
}

func process(query, proxy string, idx int) {
	client := resty.New().SetProxy(proxy).
		SetHeader("authorization", query).
		SetHeader("accept", "*/*").
		SetHeader("accept-language", "en-US,en;q=0.5").
		SetHeader("origin", "https://game.chickcoop.io").
		SetHeader("priority", "u=1, i").
		SetHeader("referer", "https://game.chickcoop.io/").
		SetHeader("sec-ch-ua", `"Not)A;Brand";v="99", "Brave";v="127", "Chromium";v="127"`).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-platform", `"macOS"`).
		SetHeader("sec-fetch-dest", "empty").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-site", "same-site").
		SetHeader("sec-gpc", "1").
		SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

	lastUpgradeEgg := time.Now()
	lastSpin := time.Now()
	lastDailyCheckin := time.Now()
	captchaURL := viper.GetString("captcha.server")
	for {

		// ================ FARM ================
		var state request.Response
		res, err := client.
			R().
			SetResult(&state).
			Get(constant.GetStateAPI)
		if err != nil {
			logger.Error("Get state error", zap.Any("idx", idx), zap.Error(err))
			continue
		}
		// UPGRADE EGG + AUO

		// ZONE FARM

		//if time.Now().Sub(lastUpgradeEgg) > time.Hour {
		if time.Now().Sub(lastUpgradeEgg) > time.Hour {
			lastUpgradeEgg = time.Now()
			res, err := client.
				R().
				SetBody(request.ResearchAPIRequest{ResearchType: constant.ResearchTypeEggValue}).
				SetResult(&state).
				Post(constant.ResearchAPI)
			if err != nil {
				logger.Error("Try to upgrade egg value error", zap.Any("idx", idx), zap.Error(err))
			} else {
				logger.Info("Try to upgrade egg value : account ", zap.Any("idx", idx), zap.Any("state", res))

			}

			res, err = client.
				R().
				SetBody(request.ResearchAPIRequest{ResearchType: constant.ResearchTypeLayingRate}).
				SetResult(&state).
				Post(constant.ResearchAPI)
			if err != nil {
				logger.Error("Try to upgrade laying rate error", zap.Any("idx", idx), zap.Error(err))
			} else {
				logger.Info("Try to upgrade laying rate : account ", zap.Any("idx", idx), zap.Any("state", res))

			}
			// AUTO HATCH

			if !state.Data.Eggs.HatchingRate.IsAuto {
				res, err = client.
					R().
					SetResult(&state).
					Post(constant.AutoHatchAPI)
				if err != nil {
					logger.Error("Try to enable auto error", zap.Any("idx", idx), zap.Error(err))
				} else {
					logger.Info("Try to enable auto : account ", zap.Any("idx", idx), zap.Any("state", res))

				}
			}
			//else {
			//	res, err := client.
			//		R().
			//		SetResult(&state).
			//		Post(constant.AutoHatchAPI)
			//	if err != nil {
			//		logger.Error("Try to upgrade auto  hatch error", zap.Any("idx", idx), zap.Error(err))
			//	} else {
			//		logger.Info("Try to upgrade auto  hatch : account ", zap.Any("idx", idx), zap.Any("state", res))
			//
			//	}
			//}

		}

		// DAILY
		if time.Now().Sub(lastDailyCheckin) > 12*time.Hour {
			body := struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				GemsReward int    `json:"gemsReward"`
				Achieved   bool   `json:"achieved"`
				Rewarded   bool   `json:"rewarded"`
			}{
				ID:         "daily.checkin",
				Name:       "Check in game",
				GemsReward: 5,
				Achieved:   true,
				Rewarded:   false,
			}

			lastUpgradeEgg = time.Now()
			res, err := client.
				R().
				SetBody(body).
				Post(constant.ClaimDailyAPI)
			if err != nil {
				logger.Error("Try to check in error", zap.Any("idx", idx), zap.Error(err))
			} else {
				logger.Info("Try to check in : account ", zap.Any("idx", idx), zap.Any("state", res))

			}
			lastDailyCheckin = time.Now()

		}
		// =========== SPIN ===============

		if time.Now().Sub(lastSpin) > time.Hour {
			var spinResult request.SpinResult
			lastUpgradeEgg = time.Now()
			res, err := client.
				R().
				SetBody(request.SpinRequest{Mode: constant.SpinModeFree}).
				SetResult(&spinResult).
				Post(constant.SpinAPI)
			if err != nil {
				logger.Error("Try to tak a spin error", zap.Any("idx", idx), zap.Error(err))
			} else {
				logger.Info("Try to take a spin : account ", zap.Any("idx", idx), zap.Any("state", res))

			}

			if spinResult.Ok {
				res, err = client.
					R().
					SetBody(spinResult).
					Post(constant.ClaimSpinAPI)
				if err != nil {
					logger.Error("Try to claim the spin result error", zap.Any("idx", idx), zap.Error(err))
				} else {
					logger.Info("Try to the spin result : account ", zap.Any("idx", idx), zap.Any("state", res))

				}
			}
			lastSpin = time.Now()

		}
		// =========== random gift ===============

		if state.Data.RandomGift != nil {
			res, err := client.
				R().
				SetBody(state).
				SetResult(&state).
				Post(constant.GiftClaimAPI)
			if err != nil {
				logger.Error("claim gift error", zap.Any("idx", idx), zap.Error(err))
				continue
			}
			logger.Info("Claim gift: account ", zap.Any("idx", idx), zap.Any("gift", res))
		}

		// =========== upgrade egg ===============

		if state.Data.Discovery.AvailableToUpgrade {
			state.Data.Discovery.Level++
			state.Data.Discovery.AvailableToUpgrade = false

			res, err := client.
				R().
				SetBody(state).
				SetResult(&state).
				Post(constant.UpgradelevelAPI)
			if err != nil {
				logger.Error("Upgrade egg error", zap.Any("idx", idx), zap.Error(err))
				continue
			}
			logger.Info("Upgrade egg level: account ", zap.Any("idx", idx), zap.Any("egg", res))

		}
		// =========== upgrade capacity ===============

		if state.Data.FarmCapacity.NeedToUpgrade {

			res, err := client.
				R().
				SetBody(request.ResearchAPIRequest{ResearchType: constant.ResearchTypeFarmCapacity}).
				SetResult(&state).
				Post(constant.ResearchAPI)
			if err != nil {
				logger.Error("Upgrade farm capacity error", zap.Any("idx", idx), zap.Error(err))
				continue
			}
			logger.Info("Upgrade farm capacity : account ", zap.Any("idx", idx), zap.Any("state", res))

		}

		// =========== manual hatch ===============
		state.Data.Chickens.Quantity++

		res, err = client.
			R().
			SetBody(state).
			SetResult(&state).
			Post(constant.HatchAPI)
		if err != nil {
			logger.Error("Hatch error", zap.Any("idx", idx), zap.Error(err))
			continue
		}
		if state.Ok {
			logger.Info("Hatch account ", zap.Any("idx", idx), zap.Any("status", state.Ok))
		} else {
			logger.Error("Hatch account error", zap.Any("idx", idx), zap.Any("error", res))

			//========== TRY TO VERIFY CAPTCHA =================
			var challenge request.ChallengeResponse
			_, err := client.
				R().
				SetResult(&challenge).
				Get(constant.GetChallenge)
			if err != nil {
				logger.Error("Get challenge error", zap.Any("idx", idx), zap.Error(err))
				continue
			}
			if challenge.Ok {
				logger.Info("Bypass captcha - doing : ", zap.Any("idx", idx))
				res, err = resty.New().
					R().
					SetBody(challenge.Data.Challenge).
					SetResult(&challenge).
					Post(captchaURL)
				logger.Info("Detect captcha - result : ", zap.Any("idx", idx), zap.Any("res", res.String()))

				res, err = client.
					R().
					SetBody(res.String()).
					SetResult(&challenge).
					Post(constant.VerifyChallenge)
				logger.Info("Bypass captcha - result : ", zap.Any("idx", idx), zap.Any("res", res.String()))
				//========== TRY TO VERIFY CAPTCHA =================

			}

		}
		// ================ END FARM ================
		// ================ LAND ====================

		var landState request.LandState
		res, err = client.
			R().
			SetResult(&landState).
			Get(constant.GetLandStateAPI)
		if err != nil {
			logger.Error("Get land state error", zap.Any("idx", idx), zap.Error(err))
			continue
		}

		for _, s := range landState.Data.Land.Normal {
			if s.State != "available" {
				continue
			}
			wateringBody := request.WateringRequest{
				TileID: s.ID,
			}

			if s.Plant != nil && time.Now().UnixMilli() >= s.Plant.HarvestAt {
				res, err = client.
					R().
					SetBody(wateringBody).
					SetResult(&landState).
					Post(constant.HarvestAPI)
				if err != nil {
					logger.Error("Get land state error", zap.Any("idx", idx), zap.Error(err))
					continue
				}
			}

			if s.Plant == nil {
				body := request.PurchaseSeedRequest{
					PlantID:  "flower.Daisy",
					Quantity: 1,
				}
				res, err = client.
					R().
					SetBody(body).
					SetResult(&landState).
					Post(constant.PurchaseSeedsAPI)
				if err != nil {
					logger.Error("PurchaseSeedsAPI error", zap.Any("idx", idx), zap.Error(err))
					continue
				}
				logger.Info("PurchaseSeedsAPI success", zap.Any("idx", idx), zap.Any("res", res))

				putInBody := request.PutInRequest{
					PlantID: "flower.Daisy",
					TileID:  s.ID,
				}
				res, err = client.
					R().
					SetBody(putInBody).
					SetResult(&landState).
					Post(constant.PutInAPI)
				if err != nil {
					logger.Error("PutInAPI error", zap.Any("idx", idx), zap.Error(err))
					continue
				}
				logger.Info("Putin success", zap.Any("idx", idx), zap.Any("res", res))

				res, err = client.
					R().
					SetBody(wateringBody).
					SetResult(&landState).
					Post(constant.WateringAPI)
				if err != nil {
					logger.Error("WateringAPI error", zap.Any("idx", idx), zap.Error(err))
					continue
				}
				logger.Info("Watering success", zap.Any("idx", idx), zap.Any("res", res))

			}

		}

		// ================ END LAND ================

		time.Sleep(3 * time.Second)
	}
}
