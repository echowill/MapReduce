package data

import (
	"fmt"
	"os/exec"
	"strconv"
	"testing"
)

func TestCreateData(t *testing.T) {
	for i := 0; i < 100; i++ {
		command := "./createData.sh " + "sensor " + strconv.Itoa(i) + " 100"
		cmd := exec.Command("/bin/bash", "-c", command)
		out, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(out))
		}

	}
}
