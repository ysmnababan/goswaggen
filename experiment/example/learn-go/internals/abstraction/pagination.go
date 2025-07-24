package abstraction

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	defaultPageSize = 100
)

type PaginationCursor struct {
	PageSize int     `json:"page_size" query:"page_size" default:"100"`
	Cursor   string  `json:"cursor"    query:"cursor"`
	SortBy   *string `json:"sort_by"   query:"sort_by"                 example:"id"`
	OrderBy  *string `json:"order_by"  query:"order_by"                             enums:"asc,desc"`
}

func (p *PaginationCursor) Limit() int {
	if p == nil || p.PageSize <= 0 || p.PageSize > defaultPageSize {
		return defaultPageSize
	}

	return p.PageSize
}

func (p *PaginationCursor) GetPageSize() int {
	if p == nil || p.PageSize <= 0 || p.PageSize > defaultPageSize {
		return defaultPageSize
	}
	return p.PageSize
}

type Pagination struct {
	Page     int     `json:"page" query:"page" default:"1"`
	PageSize int     `json:"page_size" query:"page_size" default:"100"`
	Cursor   string  `json:"cursor" query:"cursor"`
	SortBy   *string `json:"sort_by" query:"sort_by" example:"id"`
	OrderBy  *string `json:"order_by" query:"order_by" enums:"asc,desc"`
}

func (p *Pagination) Limit() int {
	if p == nil || p.PageSize <= 0 {
		return defaultPageSize
	}
	return p.PageSize
}

func (p *Pagination) Offset() int {
	if p == nil || p.Page <= 0 {
		return 0
	}
	return (p.Page - 1)
}

func (p *Pagination) GetPage() int {
	if p == nil || p.Page <= 0 {
		return 1
	}
	return p.Page
}

func (p *Pagination) GetPageSize() int {
	if p == nil || p.PageSize <= 0 || p.PageSize > defaultPageSize {
		return defaultPageSize
	}
	return p.PageSize
}

func (p *Pagination) Apply(db *gorm.DB) {
	if p != nil {
		db.Limit(p.Limit())
		db.Offset(p.Offset())
		if p.OrderBy != nil {
			db.Order(p.GetSortBy())
		}
	}
}

// Return SortBy value if not nil with OrderBy value with default asc
func (p *Pagination) GetSortBy() string {
	var sort string

	if p == nil {
		return ""
	}

	if p.SortBy == nil {
		identifier := "id"
		p.SortBy = &identifier
	}

	if *p.SortBy == "order" {
		*p.SortBy = fmt.Sprintf("\"%s\"", "order")
	}

	if p.OrderBy == nil {
		sort = "desc"
	} else {
		sort = *p.OrderBy
	}

	switch sort {
	case "asc":
		sort = "asc"
	case "desc":
		sort = "desc"
	default:
		sort = "desc"
	}

	return *p.SortBy + " " + sort
}

func (p *Pagination) GetSorting() *Sorting {
	if p.OrderBy == nil {
		asc := "asc"
		p.OrderBy = &asc
	}
	if p.SortBy != nil {
		return &Sorting{
			*p.SortBy,
			*p.OrderBy,
		}
	}

	return nil
}

type Sorting struct {
	SortBy  string `json:"sort_by" query:"sort_by"`
	OrderBy string `json:"order_by" query:"order_by" enums:"asc,desc" default:"asc"`
}

func NewSorting(sortBy, sort string) *Sorting {
	if sort != "desc" {
		sort = "asc"
	}
	return &Sorting{
		SortBy:  sortBy,
		OrderBy: sort,
	}
}

type PaginationInfo struct {
	*Pagination
	*Sorting
	Count       int    `json:"count"`
	TotalCount  int    `json:"total_count,omitempty"`
	MoreRecords bool   `json:"more_records"`
	NextCursor  string `json:"next_cursor"`
}

func NewPageInfo(count int, moreRecords bool, p *Pagination, s *Sorting) *PaginationInfo {
	if s == nil {
		s = &Sorting{}
	}
	return &PaginationInfo{
		Pagination: &Pagination{
			Page:     p.GetPage(),
			PageSize: p.GetPageSize(),
		},
		Sorting:     s,
		Count:       count,
		MoreRecords: moreRecords,
	}
}
