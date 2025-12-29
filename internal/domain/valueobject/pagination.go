package valueobject

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
}

type PaginationParams struct {
	Page     int
	PageSize int
}

func NewPaginationParams(page, pageSize int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p PaginationParams) Limit() int {
	return p.PageSize
}

func NewPaginatedResponse[T any](items []T, totalCount int64, params PaginationParams) PaginatedResponse[T] {
	totalPages := int(totalCount) / params.PageSize
	if int(totalCount)%params.PageSize > 0 {
		totalPages++
	}
	if items == nil {
		items = []T{}
	}
	return PaginatedResponse[T]{
		Data:       items,
		Total:      totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}
}
