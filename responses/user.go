package responses

type UserResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type UserResponsePagination struct {
	Status     int                    `json:"status"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data"`
	Pagination Pagination             `json:"pagination"`
}

type PaginationArgs struct {
	Limit int
	Skip  int
	Sort  string
	Order int
}

type Pagination struct {
	Limit int    `json:"limit"`
	Skip  int    `json:"skip"`
	Sort  string `json:"sort"`
	Total int    `json:"total"`
}
