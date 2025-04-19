package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginationFeedQuery struct {
	Limit  int64      `json:"limit" validate:"gte=1,lte=20"`
	Offset int64      `json:"offset" validate:"gte=0"`
	Sort   string     `json:"sort" validate:"oneof=asc desc"`
	Tags   []string   `json:"tags" validate:"max=20"`
	Search string     `json:"search" validate:"max=100"`
	Since  *time.Time `json:"since"`
	Until  *time.Time `json:"until"`
}

func (fq PaginationFeedQuery) Parse(r *http.Request) (PaginationFeedQuery, error) {
	query := r.URL.Query()

	limit := query.Get("limit")
	if limit != "" {
		l, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return fq, err
		}
		fq.Limit = int64(l)
	}

	offset := query.Get("offset")
	if offset != "" {
		o, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return fq, err
		}
		fq.Offset = int64(o)
	}

	sort := query.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	tags := query.Get("tags")
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	} else {
		fq.Tags = []string{}
	}

	search := query.Get("search")
	if search != "" {
		fq.Search = search
	}

	since := query.Get("since")
	if since != "" {
		fq.Since = parseTime(since)
	}

	until := query.Get("until")
	if until != "" {
		fq.Until = parseTime(until)
	}

	return fq, nil
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return nil
	}
	return &t
}
