package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// parseHideOptions parses the command line arguments of the subcommand "hide".
// func parseHideOptions(opts *option.HideOptions, c *cli.Context) error {
// 	opts.Org = c.String("org")
// 	opts.Repo = c.String("repo")
// 	opts.Token = c.String("token")
// 	opts.ConfigPath = c.String("config")
// 	opts.MRNumber = c.Int("mr")
// 	opts.DryRun = c.Bool("dry-run")
// 	opts.SkipNoToken = c.Bool("skip-no-token")
// 	opts.Silent = c.Bool("silent")
// 	opts.LogLevel = c.String("log-level")
// 	opts.HideKey = c.String("hide-key")
// 	opts.Condition = c.String("condition")
// 	opts.SHA1 = c.String("sha1")
// 	vars, err := parseVarsFlag(c.StringSlice("var"))
// 	if err != nil {
// 		return err
// 	}
// 	varFiles, err := parseVarFilesFlag(c.StringSlice("var-file"))
// 	if err != nil {
// 		return err
// 	}
// 	for k, v := range varFiles {
// 		vars[k] = v
// 	}
// 	opts.Vars = vars

// 	return nil
// }

// hideAction is an entrypoint of the subcommand "hide".
func (runner *Runner) hideAction(c *cli.Context) error {
	return fmt.Errorf("gitlab does not support the hide option for note :(")

	// if a := os.Getenv("GITLAB_COMMENT_SKIP"); a != "" {
	// 	skipComment, err := strconv.ParseBool(a)
	// 	if err != nil {
	// 		return fmt.Errorf("parse the environment variable GITLAB_COMMENT_SKIP as a bool: %w", err)
	// 	}
	// 	if skipComment {
	// 		return nil
	// 	}
	// }
	// opts := &option.HideOptions{}
	// if err := parseHideOptions(opts, c); err != nil {
	// 	return err
	// }

	// setLogLevel(opts.LogLevel)
	// wd, err := os.Getwd()
	// if err != nil {
	// 	return fmt.Errorf("get a current directory path: %w", err)
	// }

	// cfgReader := config.Reader{
	// 	ExistFile: existFile,
	// }

	// cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	// if err != nil {
	// 	return fmt.Errorf("find and read a configuration file: %w", err)
	// }
	// opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken

	// var pt api.Platform = platform.Get()

	// gl, err := getGitlab(c.Context, &opts.Options, cfg)
	// if err != nil {
	// 	return fmt.Errorf("initialize commenter: %w", err)
	// }

	// ctrl := api.HideController{
	// 	Wd:     wd,
	// 	Getenv: os.Getenv,
	// 	HasStdin: func() bool {
	// 		return !term.IsTerminal(0)
	// 	},
	// 	Stderr:   runner.Stderr,
	// 	Gitlab:   gl,
	// 	Platform: pt,
	// 	Config:   cfg,
	// 	Expr:     &expr.Expr{},
	// }
	// return ctrl.Hide(c.Context, opts) //nolint:wrapcheck
}
