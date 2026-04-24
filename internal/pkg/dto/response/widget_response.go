package response

type WidgetConfigResponse struct {
	ChannelID   string `json:"channelId"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	WelcomeText string `json:"welcomeText"`
	ThemeColor  string `json:"themeColor"`
	Position    string `json:"position"`
	Width       string `json:"width"`
}
