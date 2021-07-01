package types

import (
	"github.com/cosmos/cosmos-sdk/types/query"
)

func NewQueryAllFundsRequest(req *query.PageRequest) *QueryAllFundsRequest {
	return &QueryAllFundsRequest{Pagination: req}
}
