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

// DialDeleteCommand represents a command for deleting dials.
type DialDeleteCommand struct {
	ConfigPath string
}

// Run executes the command.
func (c *DialDeleteCommand) Run(ctx context.Context, args []string, stdout io.Writer) error {
	// Create flag set to parse the config path & read the ID.
	fs := flag.NewFlagSet("wtf-dial-delete", flag.ContinueOnError)
	attachConfigFlags(fs, &c.ConfigPath)
	if err := fs.Parse(args); err != nil {
		return err
	} else if fs.NArg() == 0 {
		return errors.New("wtf dial delete: dial ID required")
	} else if fs.NArg() > 1 {
		return errors.New("wtf dial delete: only one dial ID allowed")
	}

	// Parse the dial ID from the first arg.
	id, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return errors.New("wtf dial delete: invalid dial ID")
	}

	// Load configuration file.
	config, err := ReadConfigFile(c.ConfigPath)
	if err != nil {
		return err
	}

	// Authenticate user using the API key.
	ctx = wtf.NewContextWithUser(ctx, &wtf.User{APIKey: config.APIKey})

	// Instantiate HTTP service and issue delete.
	svc := http.NewDialService(http.NewClient(config.URL))
	if err := svc.DeleteDial(ctx, id); err != nil {
		return err
	}

	// Notify user that dial is gone.
	fmt.Fprintln(stdout, "Your dial has been deleted.")

	return nil
}
