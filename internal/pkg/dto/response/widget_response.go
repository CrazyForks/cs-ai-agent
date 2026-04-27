package response

type WidgetConfigResponse struct {
	ChannelID      string `json:"channelId"`
	ChannelType    string `json:"channelType"`
	ExternalSource string `json:"externalSource"`
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	ThemeColor     string `json:"themeColor"`
	Position       string `json:"position"`
	Width          string `json:"width"`
}
