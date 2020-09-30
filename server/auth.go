package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/robert-pkg/gateway/conf"
	http_client "github.com/robert-pkg/gateway/http-client"
	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/rpc"
	rpc_http "github.com/robert-pkg/micro-go/rpc/client/http"
	"github.com/robert-pkg/micro-go/rpc/metadata"
)

type cacheToken struct {
	UserID     int64
	DeviceType string
	Token      string
	ExpireTS   int64 // 过期时间戳
}

type auth struct {
	verifyTokenClient    *rpc_http.Client
	verifyToenMethodName string

	cacheTokenMap map[string][]*cacheToken

	getChannel       chan cacheToken
	getResultChannel chan *cacheToken
	addChannel       chan *cacheToken
}

// initVerifyTokenClient .
func (auth *auth) initAuth(cfg *conf.Config) error {

	verifyTokenServerName, verifyTokenMethod := cfg.GetVerifyTokenInfo()

	gClient, err := http_client.GetClient(verifyTokenServerName)
	if err != nil {
		return err
	}

	auth.verifyTokenClient = gClient
	auth.verifyToenMethodName = verifyTokenMethod

	auth.cacheTokenMap = make(map[string][]*cacheToken)

	auth.getChannel = make(chan cacheToken)
	auth.getResultChannel = make(chan *cacheToken)
	auth.addChannel = make(chan *cacheToken)

	go auth.run()
	return nil
}

func (auth *auth) run() {
	for {
		select {
		case <-auth.getChannel:

			auth.getResultChannel <- nil

		case res := <-auth.addChannel:
			auth.add2cacheImp(res)
		}
	}

}

func (auth *auth) add2cacheImp(cache *cacheToken) {

	key := fmt.Sprintf("%d-%s", cache.UserID, cache.DeviceType)
	tList, ok := auth.cacheTokenMap[key]
	if !ok {
		// 不存在，则新增
		auth.cacheTokenMap[key] = []*cacheToken{cache}
		return
	}

	for _, r := range tList {
		if r.Token == cache.Token {
			// 已存在
			return
		}
	}

	tList = append(tList, cache)

	if len(tList) > 10 {
		// 按ExpireTS从小到大排序
		sort.Slice(tList, func(i, j int) bool {
			return tList[i].ExpireTS < tList[j].ExpireTS
		})

		// 只保留10个
		tList = tList[len(tList)-10:]
	}

	auth.cacheTokenMap[key] = tList

}

func (auth *auth) getFromCache(userID int64, deviceType string, token string) (isExist bool, isValid bool, err error) {

	cache := cacheToken{
		UserID:     userID,
		DeviceType: deviceType,
		Token:      token,
		ExpireTS:   0,
	}

	auth.getChannel <- cache
	if res, ok := <-auth.getResultChannel; !ok {
		return false, false, errors.New("gateway is closed")
	} else if res == nil {
		return false, false, nil
	} else {
		if res.ExpireTS <= time.Now().Unix() {
			// token 已过期
			return true, false, nil
		}
	}

	return true, true, nil
}

func (auth *auth) VerifyToken(ctx context.Context, header *messageHeader) (httpCode int, resp string) {

	// verify from cache
	if isCacheExist, isCacheValid, err := auth.getFromCache(header.UserID, header.DeviceType, header.Token); err != nil {
		return http.StatusInternalServerError, err.Error()
	} else if isCacheExist {
		if isCacheValid {
			return http.StatusOK, "OK"
		}

		// token 无效
		return http.StatusUnauthorized, "无效token"
	}

	if isValid, expireTS, err := auth.verifyImp(ctx, header.UserID, header.DeviceType, header.Token); err != nil {
		return http.StatusInternalServerError, err.Error()
	} else if isValid {

		if expireTS <= time.Now().Unix() {
			// token 过期
			return http.StatusUnauthorized, "无效token"
		}

		cacheToken := &cacheToken{
			UserID:     header.UserID,
			DeviceType: header.DeviceType,
			Token:      header.Token,
			ExpireTS:   expireTS,
		}
		auth.addChannel <- cacheToken

		return http.StatusOK, "OK"

	}

	// token 无效
	return http.StatusUnauthorized, "无效token"
}

func (auth *auth) verifyImp(ctx context.Context, userID int64, deviceType string, token string) (bool, int64, error) {
	var verifyTokenReq struct {
		UserID     int64  `json:"user_id,omitempty"`
		DeviceType string `json:"device_type,omitempty"`
		Token      string `json:"token,omitempty"`
	}

	verifyTokenReq.UserID = userID
	verifyTokenReq.Token = token
	verifyTokenReq.DeviceType = deviceType

	var verifyTokenReply struct {
		IsValid  bool  `json:"is_valid,omitempty"`
		ExpireTS int64 `json:"expire_ts,omitempty"`
	}

	header := map[string]string{
		rpc.SkipTrace: "true",
	}
	ctx = metadata.NewContext(ctx, header)

	// 去验证token，属于内部调用，无需header
	if err := auth.verifyTokenClient.Call(ctx, auth.verifyToenMethodName, verifyTokenReq, &verifyTokenReply); err != nil {
		if err == rpc_http.ErrNoAvailableConn {
			log.Error("err", "err", err)
		}

		return false, 0, err
	}

	return verifyTokenReply.IsValid, verifyTokenReply.ExpireTS, nil
}
