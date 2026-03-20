package domain

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// PaginatedResult wraps a paginated response with metadata.
type PaginatedResult struct {
	Items      any `json:"items"`
	TotalItems int `json:"total_items"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}
