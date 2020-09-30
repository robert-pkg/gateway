package server

import (
	"net/http"
	"strconv"

	"github.com/robert-pkg/micro-go/rpc"
)

type messageHeader struct {
	AppName string
	AppVer  string

	DeviceType string
	Sign       string
	TimeStamp  int64
	Token      string
	UserID     int64
}

func (h *messageHeader) extractHeader(header http.Header) {

	for k, v := range header {
		if len(v) <= 0 {
			continue
		}

		if k == "App_name" {
			h.AppName = v[0]
		} else if k == "App_ver" {
			h.AppVer = v[0]
		} else if k == "Device_type" {
			h.DeviceType = v[0]
		} else if k == "Sign" {
			h.Sign = v[0]
		} else if k == "Ts" {
			h.TimeStamp, _ = strconv.ParseInt(v[0], 10, 64)
		} else if k == "Token" {
			h.Token = v[0]
		} else if k == "Uid" {
			h.UserID, _ = strconv.ParseInt(v[0], 10, 64)
		}
	}

	return
}

func (h *messageHeader) makeBestHeader() map[string]string {

	header := make(map[string]string)
	header[rpc.UserID] = strconv.FormatInt(h.UserID, 10)
	header[rpc.DeviceType] = h.DeviceType

	return header
}
