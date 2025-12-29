package entity

import "time"

// BrowserHistory represents a browser history entry
type BrowserHistory struct {
	ID                         int32
	URL                        string
	Title                      *string
	VisitDate                  *time.Time
	Typed                      *bool
	Hidden                     *bool
	ImportedFromFirefoxPlaceID *int32
	ImportedFromFirefoxVisitID *int32
	Domain                     *string
	CreatedAt                  *time.Time
}

// BrowserHistoryFilters represents filters for querying browser history
type BrowserHistoryFilters struct {
	SearchQuery *string
	StartDate   *time.Time
	EndDate     *time.Time
	Domain      *string
	Page        int32
	PageSize    int32
}

// DomainVisitCount represents aggregated visit count by domain
type DomainVisitCount struct {
	Domain     string
	VisitCount int64
}
