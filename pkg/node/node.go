// Copyright © 2020 The Platform9 Systems Inc.
package node

import (
	"fmt"

	"CloudManager/pkg/cmdexec"
	"CloudManager/pkg/keystone"
	"CloudManager/pkg/util"

	"go.uber.org/zap"
)

// This variable is assigned with StatusCode during hostagent installation
var HostAgent int
var IsRemoteExecutor bool
var homeDir string

const (
	// Response Status Codes
	HostAgentCertless = 200
	HostAgentLegacy   = 404
)

// PrepNode sets up prerequisites for k8s stack
func FetchRegionFQDN(ctx util.Config, auth keystone.KeystoneAuth) (string, error) {

	// "regionInfo" service will have endpoint information. So fetch it's service ID.
	regionInfoServiceID, err := util.GetServiceID(ctx.Fqdn, auth, "regionInfo")
	if err != nil {
		return "", fmt.Errorf("Failed to fetch installer URL, Error: %s", err)
	}
	zap.S().Debug("Service ID fetched : ", regionInfoServiceID)

	// Fetch the endpoint based on region name.
	endpointURL, err := util.GetEndpointForRegion(ctx.Fqdn, auth, ctx.Region, regionInfoServiceID)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch installer URL, Error: %s", err)
	}
	zap.S().Debug("endpointURL fetched : ", endpointURL)
	return endpointURL, nil
}

func CheckSudo(exec cmdexec.Executor) bool {
	_, err := exec.RunWithStdout("-l")
	return err == nil
}
