package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	miniProgramAppID       = "wx9f30f1cea85e1e8c"
	transportPublicKeyPath = ".local/keys/transport-public.pem"
	safePublicKeyPath      = ".local/keys/safe-public.pem"
)

var shanghaiLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type ciphercodeRequest struct {
	BatchID   string `json:"batchid"`
	BatchCode string `json:"batchcode"`
	Sign      string `json:"sign"`
}

type rushFlowOptions struct {
	TemplateID  string `json:"templateid"`
	CollectType int    `json:"collecttype"`
	Mobile      string `json:"mobile"`
}

type rushFlowResult struct {
	Ciphercode ciphercodeResponse
	AsyncRush  asyncRushResponse
}

func runRushFlow(
	ctx context.Context,
	client *http.Client,
	cipherParams requestParams,
	options rushFlowOptions,
) (rushFlowResult, error) {
	if err := waitUntilRushRelease(ctx); err != nil {
		return rushFlowResult{}, err
	}
	return runRushFlowNow(ctx, client, cipherParams, options)
}

func runRushFlowNow(
	ctx context.Context,
	client *http.Client,
	cipherParams requestParams,
	options rushFlowOptions,
) (rushFlowResult, error) {
	cipherRequest, err := parseCiphercodeRequest(cipherParams.Body)
	if err != nil {
		return rushFlowResult{}, err
	}
	cookie := cipherParams.Headers.Get("Cookie")
	if cookie == "" {
		return rushFlowResult{}, fmt.Errorf("ciphercode packet is missing Cookie")
	}

	cipherResult, err := ciphercode(ctx, client, cipherParams)
	if err != nil {
		return rushFlowResult{}, fmt.Errorf("request ciphercode: %w", err)
	}
	if cipherResult.Code != 0 || cipherResult.Data == nil {
		return rushFlowResult{}, fmt.Errorf(
			"ciphercode failed: code=%d, msg=%s",
			cipherResult.Code,
			cipherResult.Msg,
		)
	}
	if cipherResult.Data.Ciphercode == "" || cipherResult.Data.Safesalt == "" {
		return rushFlowResult{}, fmt.Errorf("ciphercode response is missing ciphercode or safesalt")
	}

	safeTimestamp := currentSeconds()
	safeNonce := strconv.FormatInt(safeTimestamp, 10)
	ciphertext, err := encryptSafePassword(
		cipherResult.Data.Ciphercode,
		cipherResult.Data.Safesalt,
		safeTimestamp,
		safeNonce,
		"",
		safePublicKeyPath,
	)
	if err != nil {
		return rushFlowResult{}, fmt.Errorf("encrypt safe password: %w", err)
	}

	collectType := options.CollectType
	if collectType == 0 {
		collectType = 10
	}
	location := Location{
		Nation:   "中国",
		Adcode:   "...",
		Province: "广东省",
		City:     "深圳市",
		CityCode: "...",
		District: "...",
		Township: "...",
		Address:  "...",
		Position: Position{
			Lat: 22.0,
			Lng: 114.0,
		},
		Decode: 200,
	}
	payload := AsyncRushPayload{
		ID:          cipherRequest.BatchID,
		TemplateID:  options.TemplateID,
		CollectType: collectType,
		BatchID:     cipherRequest.BatchID,
		BatchCode:   cipherRequest.BatchCode,
		CipherCode:  cipherResult.Data.Ciphercode,
		SafeSalt:    cipherResult.Data.Safesalt,
		Ciphertext:  ciphertext,
		Mobile:      options.Mobile,
		IsSafeKey:   200,
		T:           0,
		Channel:     miniProgramAppID,
		Position:    location.Position,
		Location:    location,
	}

	envelope, err := encryptTransport(
		payload,
		transportPublicKeyPath,
		currentMilliseconds(),
	)
	if err != nil {
		return rushFlowResult{}, fmt.Errorf("encrypt transport: %w", err)
	}

	asyncHeaders := make(http.Header, len(envelope.Headers)+1)
	asyncHeaders.Set("Cookie", cookie)
	for name, value := range envelope.Headers {
		asyncHeaders.Set(name, value)
	}
	asyncResult, err := asyncRush(ctx, client, requestParams{
		Headers: asyncHeaders,
		Body:    []byte(envelope.Body),
	})
	if err != nil {
		return rushFlowResult{}, fmt.Errorf("request async rush: %w", err)
	}

	return rushFlowResult{
		Ciphercode: cipherResult,
		AsyncRush:  asyncResult,
	}, nil
}

func waitUntilRushRelease(ctx context.Context) error {
	now := time.Now().In(shanghaiLocation)
	release := rushReleaseTime(now)
	if !now.Before(release) {
		return nil
	}

	timer := time.NewTimer(release.Sub(now))
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait for 12:00 release: %w", ctx.Err())
	}
}

func rushReleaseTime(now time.Time) time.Time {
	localNow := now.In(shanghaiLocation)
	return time.Date(
		localNow.Year(),
		localNow.Month(),
		localNow.Day(),
		12,
		0,
		0,
		0,
		shanghaiLocation,
	)
}

func parseCiphercodeRequest(body []byte) (ciphercodeRequest, error) {
	var request ciphercodeRequest
	if err := json.Unmarshal(body, &request); err != nil {
		return ciphercodeRequest{}, fmt.Errorf("decode ciphercode request: %w", err)
	}
	if request.BatchID == "" || request.BatchCode == "" || request.Sign == "" {
		return ciphercodeRequest{}, fmt.Errorf(
			"ciphercode request is missing batchid, batchcode, or sign",
		)
	}

	expectedSign := md5Hex(request.BatchID + "@coupon" + request.BatchCode)
	if !strings.EqualFold(request.Sign, expectedSign) {
		return ciphercodeRequest{}, fmt.Errorf(
			"ciphercode sign mismatch: expected=%s, actual=%s",
			expectedSign,
			request.Sign,
		)
	}
	request.Sign = strings.ToLower(request.Sign)
	return request, nil
}
