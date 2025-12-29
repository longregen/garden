package entity

import "time"

// DashboardStats represents comprehensive statistics for the dashboard
type DashboardStats struct {
	Contacts    CategoryStats `json:"contacts"`
	Sessions    CategoryStats `json:"sessions"`
	Bookmarks   CategoryStats `json:"bookmarks"`
	History     CategoryStats `json:"history"`
	RecentItems []RecentItem  `json:"recentItems"`
}

// CategoryStats represents statistics for a specific category
type CategoryStats struct {
	Total                int64   `json:"total"`
	RecentCount          int64   `json:"recentCount,omitempty"`
	RecentlyActive       int64   `json:"recentlyActive,omitempty"`
	MonthOverMonthChange float64 `json:"monthOverMonthChange"`
}

// RecentItem represents a recent item from any category
type RecentItem struct {
	ID       string    `json:"id"`
	Category string    `json:"category"`
	Name     string    `json:"name"`
	Date     time.Time `json:"date"`
}
