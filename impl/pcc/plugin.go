package pcc

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/gemfire/cloudcache-management-cf-plugin/domain"
	"github.com/gemfire/cloudcache-management-cf-plugin/impl/common"
	"github.com/gemfire/cloudcache-management-cf-plugin/util"
	"github.com/gemfire/cloudcache-management-cf-plugin/util/requests"
)

// BasicPlugin declares the dataset that commands work on
type BasicPlugin struct {
	commandData domain.CommandData
}

// Run is the main entry point for the CF plugin interface
// It is run once for each CF plugin command executed
func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}
	var err error
	c.commandData.Target, c.commandData.UserCommand, err = requests.GetTargetAndClusterCommand(args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	pluginConnection, err := NewPluginConnectionProvider(cliConnection)
	if err != nil {
		fmt.Printf(util.GenericErrorMessage, err.Error())
		os.Exit(1)
	}
	c.commandData.ConnnectionData, err = pluginConnection.GetConnectionData([]string{c.commandData.Target})
	if err != nil {
		fmt.Printf(util.GenericErrorMessage, err.Error())
		os.Exit(1)
	}

	// From this point common code can handle the processing of the command
	common.ProcessCommand(&c.commandData, args)

	return
}

// GetMetadata provides metadata about the CF plugin including a helptext for the user
func (c *BasicPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "pcc",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "pcc",
				HelpText: "Commands to interact with geode cluster.\n",
				UsageDetails: plugin.Usage{
					Usage: "	cf  pcc  <target>  <command>  [*options]  (* = optional)\n" +
						"\n\ttarget: a pcc_instance.\n" +
						"\tcommand: use 'cf pcc <target> commands' to see a list of supported commands \n" +
						"\toptions: see help for individual commands for options.",
				},
			},
		},
	}
}