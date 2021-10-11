package client_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/regen/types/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/group/client/testsuite"
	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	suite.Run(t, testsuite.NewIntegrationTestSuite(cfg))
}
