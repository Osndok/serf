package command

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"strings"
)

// StatusCommand is a Command implementation that instructs
// the Serf agent to either report it's string status or to
// change it. If changed, the other serfs will eventually
// report the newly provided string.
type StatusCommand struct {
	Ui cli.Ui
}

func (c *StatusCommand) Help() string {
	helpText := `
Usage: serf status [expectedStatus [newStatus]]

  Without arguments, the 'status' command causes the agent to print out the serfs currently
  stored status string. With a few exceptions, the serf will always currently see itself as
  online, so this value cannot be interpreted to mean "what the others see my status as"...
  or detect when just this serf is disconnected.

  With one argument (the expected status) the status command will act as a test, returning an
  error exit code if the current state does not match the expected state.

  With two arguments (the expected and new status), the newly given status will be broadcast
  to the serf network, but only if the current status matches the expected one. On startup
  the default status is "alive".

Options:

  -rpc-addr=127.0.0.1:7373  RPC address of the Serf agent.
`
	return strings.TrimSpace(helpText)
}

func (c *StatusCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("status", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	rpcAddr := RPCAddrFlag(cmdFlags)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	var expectedStatus string;
	var newStatus string;

	args = cmdFlags.Args()
	if len(args) > 2 {
		c.Ui.Error("Too many command line arguments. Only an expectedStatus and newStatus may be specified.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	if len(args) >= 1 {
		expectedStatus=args[0];
	}

	if (len(args)>=2) {
		newStatus=args[1];
	}

	client, err := RPCClient(*rpcAddr)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Serf agent: %s", err))
		return 1
	}
	defer client.Close()

	if len(args)==0 {
		if status, err := client.Status(); err == nil {
			c.Ui.Output(status);
			return 0
		}
		c.Ui.Error(fmt.Sprintf("Error getting status: %s", err))
		return 1
	}

	if len(args)==1 {
		if status, err := client.Status(); err == nil {
			if strings.EqualFold(status, expectedStatus) {
				return 0
			} else {
				c.Ui.Output(status)
				return 1
			}
		} else {
			c.Ui.Error(fmt.Sprintf("Error getting status: %s", err))
		}
	}

	if len(args)==2 {
		if err := client.UpdateStatus(expectedStatus, newStatus); err != nil {
			c.Ui.Error(fmt.Sprintf("Error updating status: %s", err))
			return 1
		}
		c.Ui.Output("Graceful status update")
		return 0
	}

	c.Ui.Error("Error, reached end of function")
	return 1
}

func (c *StatusCommand) Synopsis() string {
	return "Gracefully Statuss the Serf cluster and shuts down"
}
