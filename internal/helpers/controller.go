package helpers

import (
	"strconv"

	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
)

// GetPaginationArgs extracts pagination parameters from HTTP request
func GetPaginationArgs(c *gin.Context) util.PaginationArgs {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	sort := c.DefaultQuery("sort", "created_at_desc")

	return util.PaginationArgs{
		Limit: limit,
		Skip:  skip,
		Sort:  sort,
	}
}