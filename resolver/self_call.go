package resolver

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

// init checks if a function in this package has called itself
// is it has, the requested function is executed, and the program exits
// if it has not, the program continues to the main
func init() {
	isSelfCall := flag.Bool("self", false, "")
	command := flag.String("command", "", "")
	ethName := flag.String("eth-name", "", "")
	containerPid := flag.Int("container-pid", 0, "")
	ethName1 := flag.String("eth-name-1", "", "")
	containerPid1 := flag.Int("container-pid-1", 0, "")
	sTag := flag.Int("s-tag", -1, "")
	cTag := flag.Int("c-tag", -1, "")
	tunnelId := flag.Int("tunnel-id", -1, "")
	peerFabricIp := flag.String("peer-fabric-ip", "", "")
	flag.Parse()

	if !*isSelfCall {
		return
	}

	var output interface{}
	var err error
	switch *command {
	case "setup-local-container-link":
		err = setupLocalContainerLink(*ethName, *containerPid, *ethName1, *containerPid1)
	case "setup-olt-container-link":
		err = setupOLTContainerLink(*ethName, *containerPid, *sTag, *cTag)
	case "setup-remote-container-link":
		err = setupRemoteContainerLink(*ethName, *containerPid, *tunnelId, *peerFabricIp)

	case "delete-shared-olt-interface":
		err = deleteSharedOLTInterface(*sTag)
	case "delete-container-remote-interface":
		err = deleteContainerRemoteInterface(*ethName, *containerPid)
	case "delete-container-peer-interface":
		err = deleteContainerPeerInterface(*ethName, *containerPid)

	case "get-shared-olt-interfaces":
		output, err = getSharedOLTInterfaces()
	case "find-existing-remote-interface":
		output, err = findExistingRemoteInterface(*ethName, *containerPid)
	case "determine-fabric-ip":
		output, err = determineFabricIp()
	}

	//if any errors, print them to stdout
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	if output != nil {
		data, err := json.Marshal(output)
		//if any errors, print them to stdout
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		//print output to stdout
		fmt.Printf("%s", data)
	}
	os.Exit(0)
}

func execSelf(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(os.Args[0], append([]string{"--self", "--command=" + command}, args...)...)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	cmdErr := cmd.Run()

	stdout := outBuffer.Bytes()
	errout := errBuffer.Bytes()

	if len(errout) > 0 {
		fmt.Printf("%s", errout)
	}

	if cmdErr != nil {
		if len(stdout) > 0 {
			return []byte{}, errors.New(string(stdout))
		} else {
			return []byte{}, cmdErr
		}
	}
	return stdout, nil
}
