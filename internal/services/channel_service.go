package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"cs-agent/internal/wxwork"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"github.com/silenceper/wechat/v2"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/work/kf"
)

var ChannelService = newChannelService()

func newChannelService() *channelService {
	return &channelService{}
}

type channelService struct {
}

type wechatMPOAuthState struct {
	ChannelID string `json:"channelId"`
	ExpiresAt int64  `json:"expiresAt"`
}

type WechatMPOAuthResult struct {
	ChannelID    string
	ExternalID   string
	ExternalName string
}

func (s *channelService) Get(id int64) *models.Channel {
	return repositories.ChannelRepository.Get(sqls.DB(), id)
}

func (s *channelService) Take(where ...interface{}) *models.Channel {
	return repositories.ChannelRepository.Take(sqls.DB(), where...)
}

func (s *channelService) Find(cnd *sqls.Cnd) []models.Channel {
	return repositories.ChannelRepository.Find(sqls.DB(), cnd)
}

func (s *channelService) FindOne(cnd *sqls.Cnd) *models.Channel {
	return repositories.ChannelRepository.FindOne(sqls.DB(), cnd)
}

func (s *channelService) FindPageByParams(params *params.QueryParams) (list []models.Channel, paging *sqls.Paging) {
	return repositories.ChannelRepository.FindPageByParams(sqls.DB(), params)
}

func (s *channelService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Channel, paging *sqls.Paging) {
	return repositories.ChannelRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *channelService) Count(cnd *sqls.Cnd) int64 {
	return repositories.ChannelRepository.Count(sqls.DB(), cnd)
}

func (s *channelService) Create(t *models.Channel) error {
	return repositories.ChannelRepository.Create(sqls.DB(), t)
}

func (s *channelService) Update(t *models.Channel) error {
	return repositories.ChannelRepository.Update(sqls.DB(), t)
}

func (s *channelService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.ChannelRepository.Updates(sqls.DB(), id, columns)
}

func (s *channelService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.ChannelRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *channelService) CreateChannel(req request.CreateChannelRequest, operator *dto.AuthPrincipal) (*models.Channel, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildChannelModel(0, req)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := repositories.ChannelRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *channelService) UpdateChannel(req request.UpdateChannelRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("接入渠道不存在")
	}
	item, err := s.buildChannelModel(req.ID, req.CreateChannelRequest)
	if err != nil {
		return err
	}
	return repositories.ChannelRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"channel_type":     item.ChannelType,
		"channel_id":       item.ChannelID,
		"ai_agent_id":      item.AIAgentID,
		"name":             item.Name,
		"config_json":      item.ConfigJSON,
		"status":           item.Status,
		"remark":           item.Remark,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *channelService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil || item.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("接入渠道不存在")
	}
	if status != int(enums.StatusOk) && status != int(enums.StatusDisabled) {
		return errorsx.InvalidParam("状态值不合法")
	}
	return s.Updates(id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *channelService) DeleteChannel(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil || item.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("接入渠道不存在")
	}
	return s.Updates(id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *channelService) ParseWxWorkKFChannelConfig(raw string) (*dto.WxWorkKFChannelConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &dto.WxWorkKFChannelConfig{}, nil
	}
	cfg := &dto.WxWorkKFChannelConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, err
	}
	cfg.OpenKfID = strings.TrimSpace(cfg.OpenKfID)
	return cfg, nil
}

func (s *channelService) ListWxWorkKFAccounts() ([]response.WxWorkKFAccountResponse, error) {
	if !wxwork.Enabled() || wxwork.GetWorkCli() == nil {
		return nil, errorsx.BusinessError(1, "企业微信未启用或配置不完整")
	}
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		return nil, err
	}

	const limit = 100
	accounts := make([]response.WxWorkKFAccountResponse, 0)
	for offset := 0; ; offset += limit {
		result, err := cli.AccountPaging(&kf.AccountPagingRequest{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range result.AccountList {
			openKfID := strings.TrimSpace(item.OpenKFID)
			if openKfID == "" {
				continue
			}
			accounts = append(accounts, response.WxWorkKFAccountResponse{
				OpenKfID:        openKfID,
				Name:            strings.TrimSpace(item.Name),
				Avatar:          strings.TrimSpace(item.Avatar),
				ManagePrivilege: item.ManagePrivilege,
			})
		}
		if len(result.AccountList) < limit {
			break
		}
	}
	return accounts, nil
}

func (s *channelService) ParseWebChannelConfig(raw string) (*dto.WebChannelConfig, error) {
	raw = strings.TrimSpace(raw)
	cfg := &dto.WebChannelConfig{
		Title:      "在线客服",
		Subtitle:   "欢迎咨询",
		ThemeColor: "#2563eb",
		Position:   "right",
		Width:      "380px",
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), cfg); err != nil {
			return nil, err
		}
	}
	cfg.Title = strings.TrimSpace(cfg.Title)
	if cfg.Title == "" {
		cfg.Title = "在线客服"
	}
	cfg.Subtitle = strings.TrimSpace(cfg.Subtitle)
	cfg.ThemeColor = strings.TrimSpace(cfg.ThemeColor)
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = "#2563eb"
	}
	cfg.Position = strings.TrimSpace(cfg.Position)
	if cfg.Position == "" {
		cfg.Position = "right"
	}
	if cfg.Position != "left" && cfg.Position != "right" {
		return nil, errorsx.InvalidParam("Web渠道配置 position 只能为 left 或 right")
	}
	cfg.Width = strings.TrimSpace(cfg.Width)
	if cfg.Width == "" {
		cfg.Width = "380px"
	}
	return cfg, nil
}

func (s *channelService) ParseWechatMPChannelConfig(raw string) (*dto.WechatMPChannelConfig, error) {
	raw = strings.TrimSpace(raw)
	cfg := &dto.WechatMPChannelConfig{
		Title:      "公众号客服",
		Subtitle:   "欢迎咨询",
		ThemeColor: "#2563eb",
		OAuthScope: "snsapi_base",
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), cfg); err != nil {
			return nil, err
		}
	}
	cfg.Title = strings.TrimSpace(cfg.Title)
	if cfg.Title == "" {
		cfg.Title = "公众号客服"
	}
	cfg.Subtitle = strings.TrimSpace(cfg.Subtitle)
	cfg.ThemeColor = strings.TrimSpace(cfg.ThemeColor)
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = "#2563eb"
	}
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.AppSecret = strings.TrimSpace(cfg.AppSecret)
	cfg.OAuthScope = strings.TrimSpace(cfg.OAuthScope)
	if cfg.OAuthScope == "" {
		cfg.OAuthScope = "snsapi_base"
	}
	if cfg.OAuthScope != "snsapi_base" && cfg.OAuthScope != "snsapi_userinfo" {
		return nil, errorsx.InvalidParam("微信公众号渠道配置 oauthScope 只能为 snsapi_base 或 snsapi_userinfo")
	}
	return cfg, nil
}

func (s *channelService) BuildWechatMPOAuthURL(ctx iris.Context, channelID string) (string, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return "", errorsx.InvalidParam("channelId不能为空")
	}
	channel := repositories.ChannelRepository.GetByChannelID(sqls.DB(), channelID)
	if channel == nil || channel.Status != enums.StatusOk || channel.ChannelType != enums.ChannelTypeWechatMP {
		return "", errorsx.InvalidParam("微信公众号渠道不存在或已停用")
	}
	cfg, err := s.ParseWechatMPChannelConfig(channel.ConfigJSON)
	if err != nil {
		return "", errorsx.InvalidParam("微信公众号渠道配置不合法")
	}
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return "", errorsx.InvalidParam("微信公众号渠道缺少 appId 或 appSecret")
	}

	state, err := s.signWechatMPOAuthState(wechatMPOAuthState{
		ChannelID: channel.ChannelID,
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
	}, cfg.AppSecret)
	if err != nil {
		return "", err
	}
	redirectURI := buildAbsoluteURL(ctx, "/api/channel/wechat_mp/oauth/callback", nil)
	oa := wechat.NewWechat().GetOfficialAccount(&offConfig.Config{
		AppID:     cfg.AppID,
		AppSecret: cfg.AppSecret,
	})
	return oa.GetOauth().GetRedirectURL(redirectURI, cfg.OAuthScope, state)
}

func (s *channelService) CompleteWechatMPOAuth(ctx context.Context, code, state string) (*WechatMPOAuthResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errorsx.InvalidParam("code不能为空")
	}
	payload, err := decodeWechatMPOAuthStatePayload(state)
	if err != nil {
		return nil, errorsx.InvalidParam("OAuth state不合法")
	}
	channel := repositories.ChannelRepository.GetByChannelID(sqls.DB(), payload.ChannelID)
	if channel == nil || channel.Status != enums.StatusOk || channel.ChannelType != enums.ChannelTypeWechatMP {
		return nil, errorsx.InvalidParam("微信公众号渠道不存在或已停用")
	}
	cfg, err := s.ParseWechatMPChannelConfig(channel.ConfigJSON)
	if err != nil {
		return nil, errorsx.InvalidParam("微信公众号渠道配置不合法")
	}
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil, errorsx.InvalidParam("微信公众号渠道缺少 appId 或 appSecret")
	}
	if _, err := s.verifyWechatMPOAuthState(state, cfg.AppSecret); err != nil {
		return nil, err
	}

	oa := wechat.NewWechat().GetOfficialAccount(&offConfig.Config{
		AppID:     cfg.AppID,
		AppSecret: cfg.AppSecret,
	})
	oauth := oa.GetOauth()
	token, err := oauth.GetUserAccessTokenContext(ctx, code)
	if err != nil {
		return nil, err
	}
	externalName := ""
	if strings.Contains(token.Scope, "snsapi_userinfo") {
		if info, infoErr := oauth.GetUserInfoContext(ctx, token.AccessToken, token.OpenID, "zh_CN"); infoErr == nil {
			externalName = strings.TrimSpace(info.Nickname)
		}
	}
	if strings.TrimSpace(token.OpenID) == "" {
		return nil, errorsx.InvalidParam("微信授权未返回 openid")
	}
	return &WechatMPOAuthResult{
		ChannelID:    channel.ChannelID,
		ExternalID:   strings.TrimSpace(token.OpenID),
		ExternalName: externalName,
	}, nil
}

func (s *channelService) GetEnabledWxWorkKFChannelByOpenKfID(openKfID string) *models.Channel {
	openKfID = strings.TrimSpace(openKfID)
	if openKfID == "" {
		return nil
	}
	channels := s.Find(sqls.NewCnd().
		Eq("channel_type", enums.ChannelTypeWxWorkKF).
		Eq("status", enums.StatusOk).
		Asc("id"))
	for i := range channels {
		cfg, err := s.ParseWxWorkKFChannelConfig(channels[i].ConfigJSON)
		if err != nil {
			continue
		}
		if cfg != nil && cfg.OpenKfID == openKfID {
			return &channels[i]
		}
	}
	return nil
}

func (s *channelService) GetEnabledChannel(ctx iris.Context) *models.Channel {
	channelID := irisx.GetChannelID(ctx)
	channel := repositories.ChannelRepository.GetByChannelID(sqls.DB(), channelID)
	if channel == nil {
		return nil
	}
	if channel.Status != enums.StatusOk {
		return nil
	}
	return channel
}

func (s *channelService) buildChannelModel(id int64, req request.CreateChannelRequest) (*models.Channel, error) {
	channelType := strings.TrimSpace(req.ChannelType)
	if channelType != enums.ChannelTypeWeb && channelType != enums.ChannelTypeWechatMP && channelType != enums.ChannelTypeWxWorkKF {
		return nil, errorsx.InvalidParam("渠道类型不合法")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("渠道名称不能为空")
	}
	if req.AIAgentID <= 0 {
		return nil, errorsx.InvalidParam("请选择 AI Agent")
	}
	aiAgent := AIAgentService.Get(req.AIAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI Agent 不存在或未启用")
	}
	status := enums.Status(req.Status)
	if req.Status == 0 {
		status = enums.StatusOk
	}
	if status != enums.StatusOk && status != enums.StatusDisabled {
		return nil, errorsx.InvalidParam("渠道状态不合法")
	}

	channelID := ""
	if id > 0 {
		current := s.Get(id)
		if current == nil || current.Status == enums.StatusDeleted {
			return nil, errorsx.InvalidParam("接入渠道不存在")
		}
		channelID = strings.TrimSpace(current.ChannelID)
	}
	configJSON := strings.TrimSpace(req.ConfigJSON)
	switch channelType {
	case enums.ChannelTypeWeb:
		if channelID == "" {
			channelID = strs.UUID()
		}
		if exists := s.Take("channel_id = ? AND status <> ? AND id <> ?", channelID, enums.StatusDeleted, id); exists != nil {
			return nil, errorsx.InvalidParam("渠道标识已存在")
		}
		cfg, err := s.ParseWebChannelConfig(configJSON)
		if err != nil {
			return nil, errorsx.InvalidParam("Web渠道配置不合法")
		}
		configBytes, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		configJSON = string(configBytes)
	case enums.ChannelTypeWechatMP:
		if channelID == "" {
			channelID = strs.UUID()
		}
		if exists := s.Take("channel_id = ? AND status <> ? AND id <> ?", channelID, enums.StatusDeleted, id); exists != nil {
			return nil, errorsx.InvalidParam("渠道标识已存在")
		}
		cfg, err := s.ParseWechatMPChannelConfig(configJSON)
		if err != nil {
			return nil, errorsx.InvalidParam("微信公众号渠道配置不合法")
		}
		if cfg == nil || cfg.AppID == "" || cfg.AppSecret == "" {
			return nil, errorsx.InvalidParam("微信公众号渠道配置缺少 appId 或 appSecret")
		}
		configBytes, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		configJSON = string(configBytes)
	case enums.ChannelTypeWxWorkKF:
		if channelID == "" {
			channelID = strs.UUID()
		}
		if exists := s.Take("channel_id = ? AND status <> ? AND id <> ?", channelID, enums.StatusDeleted, id); exists != nil {
			return nil, errorsx.InvalidParam("渠道标识已存在")
		}
		cfg, err := s.ParseWxWorkKFChannelConfig(configJSON)
		if err != nil {
			return nil, errorsx.InvalidParam("企业微信渠道配置不合法")
		}
		if cfg == nil || cfg.OpenKfID == "" {
			return nil, errorsx.InvalidParam("企业微信渠道配置缺少 openKfId")
		}
		if channel := s.GetEnabledWxWorkKFChannelByOpenKfID(cfg.OpenKfID); channel != nil && channel.ID != id {
			return nil, errorsx.InvalidParam("openKfId 已被其他渠道使用")
		}
	}

	return &models.Channel{
		ChannelType: channelType,
		ChannelID:   channelID,
		AIAgentID:   req.AIAgentID,
		Name:        name,
		ConfigJSON:  configJSON,
		Status:      status,
		Remark:      strings.TrimSpace(req.Remark),
	}, nil
}

func (s *channelService) signWechatMPOAuthState(payload wechatMPOAuthState, secret string) (string, error) {
	payload.ChannelID = strings.TrimSpace(payload.ChannelID)
	if payload.ChannelID == "" || payload.ExpiresAt <= 0 {
		return "", errors.New("invalid oauth state payload")
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s|%d", payload.ChannelID, payload.ExpiresAt),
	))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature, nil
}

func (s *channelService) verifyWechatMPOAuthState(raw, secret string) (*wechatMPOAuthState, error) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, errorsx.InvalidParam("OAuth state不合法")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)
	actual, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(actual, expected) {
		return nil, errorsx.InvalidParam("OAuth state签名不合法")
	}
	payload, err := decodeWechatMPOAuthStatePayload(raw)
	if err != nil {
		return nil, errorsx.InvalidParam("OAuth state不合法")
	}
	if time.Now().Unix() > payload.ExpiresAt {
		return nil, errorsx.InvalidParam("OAuth state已过期")
	}
	return payload, nil
}

func decodeWechatMPOAuthStatePayload(raw string) (*wechatMPOAuthState, error) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || parts[0] == "" {
		return nil, errors.New("invalid oauth state")
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	stateParts := strings.Split(string(data), "|")
	if len(stateParts) != 2 {
		return nil, errors.New("invalid oauth state payload")
	}
	expiresAt, err := strconv.ParseInt(stateParts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	payload := &wechatMPOAuthState{
		ChannelID: strings.TrimSpace(stateParts[0]),
		ExpiresAt: expiresAt,
	}
	payload.ChannelID = strings.TrimSpace(payload.ChannelID)
	if payload.ChannelID == "" || payload.ExpiresAt <= 0 {
		return nil, errors.New("invalid oauth state payload")
	}
	return payload, nil
}

func buildAbsoluteURL(ctx iris.Context, path string, values url.Values) string {
	scheme := ctx.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = ctx.GetHeader("X-Scheme")
	}
	if scheme == "" {
		if ctx.Request().TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := ctx.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = ctx.Host()
	}
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
	if values != nil {
		u.RawQuery = values.Encode()
	}
	return u.String()
}

func BuildWechatMPChatRedirectURL(ctx iris.Context, result *WechatMPOAuthResult) string {
	values := url.Values{}
	values.Set("channelId", result.ChannelID)
	values.Set("externalSource", string(enums.ExternalSourceWechatMP))
	values.Set("externalId", result.ExternalID)
	if result.ExternalName != "" {
		values.Set("subject", result.ExternalName)
	}
	return buildAbsoluteURL(ctx, "/kefu/chat/", values)
}
