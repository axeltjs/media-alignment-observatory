package apirequest

import "time"

type TemporalDetectRequest struct {
	From           time.Time `json:"from" binding:"required"`
	To             time.Time `json:"to" binding:"required"`
	WindowHours    int       `json:"window_hours"`
}
