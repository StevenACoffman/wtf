package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/benbjohnson/wtf"
	"github.com/benbjohnson/wtf/http"
)

// DialSetCommand is a command for setting the WTF value for a membership.
type DialSetCommand struct {
	ConfigPath string
}

// Run executes the command.
func (c *DialSetCommand) Run(ctx context.Context, args []string, stdout io.Writer) error {
	// Create a flag set with parameters for the dial fields.
	fs := flag.NewFlagSet("wtf-dial-set", flag.ContinueOnError)
	attachConfigFlags(fs, &c.ConfigPath)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return errors.New("wtf dial set: dial ID required")
	} else if fs.NArg() == 1 {
		return errors.New("wtf dial set: WTF level required")
	} else if fs.NArg() > 2 {
		return errors.New("wtf dial set: specify only the dial ID and WTF level")
	}

	// Parse the dial ID from the first arg.
	id, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return errors.New("wtf dial set: invalid dial ID")
	}

	// Parse the WTF level from the second arg.
	value, err := strconv.Atoi(fs.Arg(1))
	if err != nil {
		return errors.New("wtf dial set: invalid WTF level")
	}

	// Load the configuration.
	config, err := ReadConfigFile(c.ConfigPath)
	if err != nil {
		return err
	}

	// Authenticate the user with the API key from the config.
	ctx = wtf.NewContextWithUser(ctx, &wtf.User{APIKey: config.APIKey})

	// Build dial from arguments and issue creation request over HTTP.
	svc := http.NewDialService(http.NewClient(config.URL))
	if err := svc.SetDialMembershipValue(ctx, id, value); err != nil {
		return err
	}

	// Notify user of the successful update.
	fmt.Fprintln(stdout, "Your WTF level has been updated.")

	return nil
}
