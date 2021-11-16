package cmd

import (
	"CloudManager/pkg/client"
	"CloudManager/pkg/cmdexec"
	"CloudManager/pkg/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

type QbertImpl struct {
	fqdn string
}

type Cluster struct {
	Uuid          string `json:"uuid"`
	CreatedAt     string `json:"created_at"`
	CloudProvider string `json:"cloudProviderType"`
}

func CheckDeleteCluster() {

	//Loads config that is passed via the

	ctx, err = util.LoadConfig("/etc/pf9/config.json")
	if err != nil {
		fmt.Println("An error has occured ", err)
		return
	}

	executor, err := getExecutor(ctx.ProxyURL)
	if err != nil {
		fmt.Println("Error connecting to host ", err.Error())
		return
	}

	c, err = client.NewClient(ctx.Fqdn, executor, ctx.AllowInsecure, false)
	if err != nil {
		fmt.Println("Unable to load clients needed for the Cmd. Error: ", err.Error())
		return
	}

	// Validate the user credentials entered during config set and will loop back again if invalid
	if err := validateUserCredentials(ctx, c); err != nil {
		clearContext(&util.Context)
		fmt.Println("Error", err)
		return
	}

	auth, err := c.Keystone.GetAuth(ctx.Username, ctx.Password, ctx.Tenant, ctx.MfaToken)
	if err != nil {
		fmt.Println("Failed to get keystone ", err.Error())
	}
	projectId := auth.ProjectID
	token := auth.Token

	clusters, err := GetAllClusters(executor, ctx.Fqdn, projectId, token)

	if err != nil {
		fmt.Println("An error has occured", err)
		return
	}

	tNow := time.Now().UTC()

	for i, _ := range clusters {

		betterFormat := clusters[i].CreatedAt
		layout := "2006-01-02T15:04:05.000Z"
		t, err := time.Parse(layout, betterFormat)

		if err != nil {
			fmt.Println(err)
		}

		duration := tNow.Sub(t)

		if duration.Hours() > 36 {

			fmt.Println("Deleting cluster (Age ", duration, "): ", clusters[i])

			err = DeleteCluster(ctx.Fqdn, clusters[i].Uuid, projectId, token)

			if err != nil {
				fmt.Println("Error deleting cluster")
			} else {
				fmt.Println("Clusted deleted")
			}

		}

	}

}

func GetAllClusters(execc cmdexec.Executor, fqdn, projectID, token string) ([]Cluster, error) {

	zap.S().Debug("Getting cluster status")
	tkn := fmt.Sprintf(`"X-Auth-Token: %v"`, token)
	cmd := fmt.Sprintf(`curl -sH %v -X GET %v/qbert/v3/%v/clusters`, tkn, fqdn, projectID)

	output, err := exec.Command("/bin/sh", "-c", cmd).Output()

	if err != nil {
		zap.S().Debug("Error fetching clusters")
		return nil, err
	}

	var awsClusters []Cluster

	var allClusters []Cluster
	json.Unmarshal([]byte(output), &allClusters)

	for _, cluster := range allClusters {
		if cluster.CloudProvider == "aws" {
			awsClusters = append(awsClusters, cluster)
		}
	}

	return awsClusters, nil

}

func DeleteCluster(fqdn, clusterID, projectID, token string) error {
	zap.S().Debugf("Deleting the %s cluster: ", clusterID)

	deleteEndpoint := fmt.Sprintf(
		"%s/qbert/v3/%s/clusters/%s",
		fqdn, projectID, clusterID)

	client := http.Client{}

	req, err := http.NewRequest("DELETE", deleteEndpoint, strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("Unable to create a request: %w", err)
	}
	req.Header.Set("X-Auth-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Unable to DELETE request through client: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respString, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			zap.S().Info("Error occured while converting response body to string")
		}
		zap.S().Debug(string(respString))
		return fmt.Errorf("%v", string(respString))
	}
	return nil
}
