package commands

import (
	gocontext "context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/urfave/cli"
)

var (
	// SnapshotterFlags are cli flags specifying snapshotter names
	SnapshotterFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "snapshotter",
			Usage: "snapshotter name. Empty value stands for the daemon default value.",
			Value: containerd.DefaultSnapshotter,
		},
	}

	// LabelFlag is a cli flag specifying labels
	LabelFlag = cli.StringSliceFlag{
		Name:  "label",
		Usage: "labels to attach to the image",
	}

	// RegistryFlags are cli flags specifying registry options
	RegistryFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "skip-verify,k",
			Usage: "skip SSL certificate validation",
		},
		cli.BoolFlag{
			Name:  "plain-http",
			Usage: "allow connections using plain HTTP",
		},
		cli.StringFlag{
			Name:  "user,u",
			Usage: "user[:password] Registry user and password",
		},
		cli.StringFlag{
			Name:  "refresh",
			Usage: "refresh token for authorization server",
		},
	}
)

// AppContext returns the context for a command. Should only be called once per
// command, near the start.
//
// This will ensure the namespace is picked up and set the timeout, if one is
// defined.
func AppContext(clicontext *cli.Context) (gocontext.Context, gocontext.CancelFunc) {
	var (
		ctx       = gocontext.Background()
		timeout   = clicontext.GlobalDuration("timeout")
		namespace = clicontext.GlobalString("namespace")
		cancel    gocontext.CancelFunc
	)

	ctx = namespaces.WithNamespace(ctx, namespace)

	if timeout > 0 {
		ctx, cancel = gocontext.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = gocontext.WithCancel(ctx)
	}

	return ctx, cancel
}

// NewClient returns a new containerd client and app context for a command.
// Should only be called once per command, near the start.
func NewClient(context *cli.Context) (*containerd.Client, gocontext.Context, gocontext.CancelFunc, error) {
	client, err := containerd.New(context.GlobalString("address"))
	if err != nil {
		return nil, nil, nil, err
	}
	ctx, cancel := AppContext(context)
	return client, ctx, cancel, nil
}

// ObjectWithLabelArgs returns the first arg and a LabelArgs object
func ObjectWithLabelArgs(clicontext *cli.Context) (string, map[string]string) {
	var (
		first        = clicontext.Args().First()
		labelStrings = clicontext.Args().Tail()
	)

	return first, LabelArgs(labelStrings)
}

// LabelArgs returns a map of label key,value pairs
func LabelArgs(labelStrings []string) map[string]string {
	labels := make(map[string]string, len(labelStrings))
	for _, label := range labelStrings {
		parts := strings.SplitN(label, "=", 2)
		key := parts[0]
		value := "true"
		if len(parts) > 1 {
			value = parts[1]
		}

		labels[key] = value
	}

	return labels
}

// PrintAsJSON prints input in JSON format
func PrintAsJSON(x interface{}) {
	b, err := json.MarshalIndent(x, "", "    ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't marshal %+v as a JSON string: %v\n", x, err)
	}
	fmt.Println(string(b))
}
