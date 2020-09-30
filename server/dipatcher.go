package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/robert-pkg/gateway/conf"
	http_client "github.com/robert-pkg/gateway/http-client"
	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/rpc/metadata"
	"github.com/robert-pkg/micro-go/utils"
)

const (
	maxSignTimeout int64 = 10 * 60 // 签名中允许的时间戳 误差
)

// dispatcher 调度员
type dispatcher struct {
	auth
}

func (d *dispatcher) checkTimeout(ts int64) bool {
	cur := time.Now().Unix()

	interval := cur - ts
	if interval < 0 {
		interval = -interval
	}

	if interval > maxSignTimeout {
		log.Warnf("isTimeout, request ts:%d, host ts:%d", ts, cur)
		return false
	}

	return true
}

// verifySign 验证 Sign，防止数据被篡改过
func (d *dispatcher) verifySign(header *messageHeader, bodyBytes []byte) (httpCode int, resp string) {
	strTimeStamp := strconv.FormatInt(header.TimeStamp, 10)
	strRawSign := header.AppName + strTimeStamp + string(bodyBytes) + conf.Conf.ServerConfig.SignKey
	md5str, err := utils.GetMd5(strRawSign)
	if err != nil {
		return http.StatusBadRequest, "数据格式错误"
	}

	if strings.Compare(md5str, strings.ToLower(header.Sign)) != 0 {
		return http.StatusBadRequest, "验签失败"
	}

	return http.StatusOK, "OK"
}

func (d *dispatcher) isSkipVerifyToken(serviceName string, methodName string) bool {
	name := serviceName + "/" + methodName
	if conf.Conf.IsSkipToken(name) {
		//log.Info("skip verify token", "name", name)
		return true
	}

	return false
}

// Dispatch
func (d *dispatcher) Call(ctx context.Context, header http.Header, serviceName string, methodName string, bodyBytes []byte) (httpCode int, resp string) {

	var msgHeader messageHeader
	msgHeader.extractHeader(header)
	//log.Infof("header: %v", msgHeader)

	if !d.checkTimeout(msgHeader.TimeStamp) {
		return http.StatusBadRequest, "请求错误, 时间不合法"
	}

	httpCode, resp = d.verifySign(&msgHeader, bodyBytes)
	if httpCode != http.StatusOK {
		return
	}

	if d.isSkipVerifyToken(serviceName, methodName) {
		//log.Infof("skip verify token.")
	} else {
		httpCode, resp = d.VerifyToken(ctx, &msgHeader)
		if httpCode != http.StatusOK {
			return
		}
	}

	httpClient, err := http_client.GetClient(serviceName)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	bestHeader := msgHeader.makeBestHeader()
	ctx = metadata.NewContext(ctx, bestHeader)

	respBytes, err := httpClient.RawCall(ctx, methodName, bodyBytes)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	return http.StatusOK, string(respBytes)
}
