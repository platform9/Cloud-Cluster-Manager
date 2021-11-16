// Copyright © 2020 The pf9ctl authors

package main

import (
	"CloudManager/cmd"
	"fmt"
	"time"
)

func main() {

	for {

		cmd.CheckDeleteCluster()
		fmt.Println("Sleeping")
		time.Sleep(time.Hour)
	}

}
