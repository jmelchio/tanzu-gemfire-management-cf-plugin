package cfservice

import (
	"bytes"
	"errors"
	"os/exec"
)

const invalidPCCInstanceMessage string = `You entered %s which not a deployed PCC instance.
To deploy this as an instance, enter: 

	cf create-service p-cloudcache <region_plan> %s

For help see: cf create-service --help
`
const invalidServiceKeyResponse string = `The cf service-key response is invalid.

For help see: cf create-service-key --help
`

// Cf receiver for CfService implementation
type Cf struct{}

// Cmd implementation for CfService interface
func (c *Cf) Cmd(name string, options ...string) (string, error) {
	options = append([]string{name}, options...)
	cmd := exec.Command("cf", options...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		switch options[0] {
		case "service-keys":
			return "", errors.New(invalidPCCInstanceMessage)
		case "service-key":
			return "", errors.New(invalidServiceKeyResponse)
		default:
			return "", err
		}
	}
	return (&out).String(), nil
}
