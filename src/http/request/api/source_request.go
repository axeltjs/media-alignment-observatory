package apirequest

type CreateSourceRequest struct {
	Name     string `json:"name" binding:"required"`
	RSSUrl   string `json:"rss_url" binding:"required,url"`
	BaseUrl  string `json:"base_url" binding:"required,url"`
	Category string `json:"category"`
}
