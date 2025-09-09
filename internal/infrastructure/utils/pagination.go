package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationData struct {
	HasOtherPages      bool
	HasPrevious        bool
	HasNext            bool
	PreviousPageNumber int
	NextPageNumber     int
	CurrentPageNumber  int
	TotalPages         int
	CustomRange        []int
}

// Paginate calculates pagination data
func Paginate(c *gin.Context, totalItems int, itemsPerPage int) PaginationData {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	totalPages := int(math.Ceil(float64(totalItems) / float64(itemsPerPage)))

	if page < 1 {
		page = 1
	}
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	startPage := page - 2
	if startPage < 1 {
		startPage = 1
	}

	endPage := page + 2
	if endPage > totalPages {
		endPage = totalPages
	}

	// Ensure at least 5 pages in custom range if possible
	if endPage-startPage+1 < 5 && totalPages >= 5 {
		if startPage == 1 {
			endPage = 5
		} else if endPage == totalPages {
			startPage = totalPages - 4
		}
	}

	var customRange []int
	for i := startPage; i <= endPage; i++ {
		customRange = append(customRange, i)
	}

	return PaginationData{
		HasOtherPages:      totalPages > 1,
		HasPrevious:        page > 1,
		HasNext:            page < totalPages,
		PreviousPageNumber: page - 1,
		NextPageNumber:     page + 1,
		CurrentPageNumber:  page,
		TotalPages:         totalPages,
		CustomRange:        customRange,
	}
}
