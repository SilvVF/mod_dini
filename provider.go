package main

import (
	"context"
	"fmt"
)

type Provider interface {
	ID() string
	List(ctx context.Context, p ListParams) (SubmissionList, error)
	Subcategories(ctx context.Context, category string) (ProviderSubcategoryList, error)
	Page(ctx context.Context, submissionID string) (SubmissionPage, error)
}

type ProviderSubcategoryList []struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	IconURL   string `json:"iconUrl"`
	ItemCount int    `json:"itemCount"`
}

type ListParams struct {
	Provider string `json:"provider"`
	Cat      string `json:"cat"`
	Query    string `json:"query"`
	Sort     string `json:"sort"`
	Page     int    `json:"page"`
}

type SubmissionList struct {
	Submissions []Submission `json:"submissions"`
	HasNext     bool         `json:"hasNext"`
	Next        int          `json:"next"`
	Prev        int          `json:"prev"`
}

type SubmissionPage struct {
	ResolvedToolId  *string          `json:"resolvedToolId,omitempty"`
	Submission      Submission       `json:"submission"`
	SubmissionFiles []SubmissionFile `json:"submissionFiles"`
}

func NewListPage(subs []Submission, page, total, perPage int) SubmissionList {
	last := 1
	if perPage > 0 {
		last = max((total+perPage-1)/perPage, 1) // ceil, so a partial last page counts
	}
	next := min(page+1, last)

	return SubmissionList{
		Submissions: subs,
		Next:        next,
		Prev:        max(page-1, 1),
		HasNext:     next != page,
	}
}

type ProviderRegistry map[string]Provider

func NewProviderRegistry(ps ...Provider) ProviderRegistry {
	r := make(ProviderRegistry, len(ps))
	for _, p := range ps {
		r[p.ID()] = p
	}
	return r
}

func (r ProviderRegistry) Get(id string) (Provider, error) {
	p, ok := r[id]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", id)
	}
	return p, nil
}
