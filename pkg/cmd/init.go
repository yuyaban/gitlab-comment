package cmd

import (
	"github.com/urfave/cli/v2"
	"github.com/yuyaban/gitlab-comment/pkg/api"
	"github.com/yuyaban/gitlab-comment/pkg/fsys"
)

// initAction is an entrypoint of the subcommand "init".
func (runner *Runner) initAction(c *cli.Context) error {
	ctrl := api.InitController{
		Fsys: &fsys.Fsys{},
	}
	return ctrl.Run(c.Context) //nolint:wrapcheck
}
