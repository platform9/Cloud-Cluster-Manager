// Copyright © 2020 The pf9ctl authors

package cmd

import (
	"CloudManager/pkg/client"
	"CloudManager/pkg/keystone"
	"CloudManager/pkg/node"
	"CloudManager/pkg/util"
	"errors"
	"fmt"
	"reflect"

	"go.uber.org/zap"
)

var (
	ctx util.Config
	err error
	c   client.Client
	// This flag is used to loop back if user enters invalid credentials during config set.
	credentialFlag bool
	// This flag is true when we set config through ./pf9ctl config set
	SetConfig             bool
	SetConfigByParameters bool
	// This flag is set true when only region is found/entered is invalid.
	RegionInvalid bool
)

const MaxLoopNoConfig = 3

//This function clears the context if it is invalid. Before storing it.
func clearContext(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

var (
	account_url   string
	username      string
	Password      string
	proxyURL      string
	region        string
	tenant        string
	overrideProxy bool
)

// This function will validate the user credentials entered during config set and bail out if invalid
func validateUserCredentials(util.Config, client.Client) error {

	auth, err := c.Keystone.GetAuth(
		ctx.Username,
		ctx.Password,
		ctx.Tenant,
		ctx.MfaToken,
	)

	if err != nil {
		RegionInvalid = false
		return err
	}

	// To validate region.
	endpointURL, err1 := node.FetchRegionFQDN(ctx, auth)
	if endpointURL == "" || err1 != nil {
		RegionInvalid = true
		zap.S().Debug("Invalid Region")
		return errors.New("Invalid Region")
	}

	return nil
}

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
