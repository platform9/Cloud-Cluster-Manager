// Copyright © 2020 The pf9ctl authors

package cmd

import (
	"CloudManager/pkg/cmdexec"
	"CloudManager/pkg/node"
	"CloudManager/pkg/ssh"
	"CloudManager/pkg/util"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	user           string
	password       string
	sshKey         string
	ips            []string
	skipChecks     bool
	disableSwapOff bool
	FoundRemote    = false
)

// getExecutor creates the right Executor
func getExecutor(proxyURL string) (cmdexec.Executor, error) {

	return cmdexec.LocalExecutor{ProxyUrl: proxyURL}, nil
}

// To check if Remote Host needs Password to access Sudo and prompt for Sudo Password if exists.
func SudoPasswordCheck(exec cmdexec.Executor) error {

	_, err := exec.RunWithStdout("-l | grep '(ALL) PASSWD: ALL'")
	if err == nil {
		// To bail out if Sudo Password entered is invalid multiple times.
		loopcounter := 1
		for true {
			loopcounter += 1
			fmt.Printf("Enter Sudo password for Remote Host: ")
			sudopassword, _ := terminal.ReadPassword(0)
			ssh.SudoPassword = string(sudopassword)
			// Validate Sudo Password entered.
			if ssh.SudoPassword == "" || validateSudoPassword(exec) == util.Invalid {
				fmt.Printf("\nInvalid Sudo Password provided of Remote Host\n")
				if loopcounter >= 4 {
					fmt.Printf("\n")
					zap.S().Debug("Invalid Sudo Password entered multiple times")
					return fmt.Errorf("Invalid password entered multiple times")
				} else {
					continue
				}
			} else {
				break
			}
		}
		fmt.Printf("\n")
	}

	return nil
}

func validateSudoPassword(exec cmdexec.Executor) string {

	_ = node.CheckSudo(exec)
	// Validate Sudo Password entered for Remote Host from stderr.
	if strings.Contains(cmdexec.StdErrSudoPassword, util.InvalidPassword) {
		return util.Invalid
	}
	return util.Valid
}
