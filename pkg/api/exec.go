package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/yuyaban/gitlab-comment/pkg/config"
	"github.com/yuyaban/gitlab-comment/pkg/execute"
	"github.com/yuyaban/gitlab-comment/pkg/expr"
	"github.com/yuyaban/gitlab-comment/pkg/gitlab"
	"github.com/yuyaban/gitlab-comment/pkg/option"
	"github.com/yuyaban/gitlab-comment/pkg/template"
)

type ExecController struct {
	Wd       string
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Getenv   func(string) string
	Reader   Reader
	Gitlab   Gitlab
	Renderer Renderer
	Executor Executor
	Expr     Expr
	Platform Platform
	Config   *config.Config
}

func (ctrl *ExecController) Exec(ctx context.Context, opts *option.ExecOptions) error { //nolint:funlen,cyclop
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementExec(opts); err != nil {
			return fmt.Errorf("complement opts with CI built in environment variables: %w", err)
		}
	}

	if opts.MRNumber == 0 && opts.SHA1 != "" {
		mrNum, err := ctrl.Gitlab.MRNumberWithSHA(ctx, opts.Org, opts.Repo, opts.SHA1)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"org":  opts.Org,
				"repo": opts.Repo,
				"sha":  opts.SHA1,
			}).Warn("list associated prs")
		}
		if mrNum > 0 {
			opts.MRNumber = mrNum
		}
	}

	cfg := ctrl.Config

	if cfg.Base != nil {
		if opts.Org == "" {
			opts.Org = cfg.Base.Org
		}
		if opts.Repo == "" {
			opts.Repo = cfg.Base.Repo
		}
	}
	result, execErr := ctrl.Executor.Run(ctx, &execute.Params{
		Cmd:   opts.Args[0],
		Args:  opts.Args[1:],
		Stdin: ctrl.Stdin,
	})

	if opts.SkipComment {
		if execErr != nil {
			return ecerror.Wrap(execErr, result.ExitCode)
		}
		return nil
	}

	execConfigs, err := ctrl.getExecConfigs(cfg, opts)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	if err := option.ValidateExec(opts); err != nil {
		return fmt.Errorf("validate command options: %w", err)
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]interface{}, len(opts.Vars))
	}
	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}

	ci := ""
	if ctrl.Platform != nil {
		ci = ctrl.Platform.CI()
	}
	joinCommand := strings.Join(opts.Args, " ")
	templates := template.GetTemplates(&template.ParamGetTemplates{
		Templates:      cfg.Templates,
		CI:             ci,
		JoinCommand:    joinCommand,
		CombinedOutput: result.CombinedOutput,
	})
	if err := ctrl.post(ctx, execConfigs, &ExecCommentParams{
		ExitCode:       result.ExitCode,
		Command:        result.Cmd,
		JoinCommand:    joinCommand,
		Stdout:         result.Stdout,
		Stderr:         result.Stderr,
		CombinedOutput: result.CombinedOutput,
		MRNumber:       opts.MRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		SHA1:           opts.SHA1,
		TemplateKey:    opts.TemplateKey,
		Template:       opts.Template,
		Vars:           cfg.Vars,
	}, templates); err != nil {
		if !opts.Silent {
			fmt.Fprintf(ctrl.Stderr, "gitlab-comment error: %+v\n", err)
		}
	}
	if execErr != nil {
		return ecerror.Wrap(execErr, result.ExitCode)
	}
	return nil
}

type ExecCommentParams struct {
	Stdout         string
	Stderr         string
	CombinedOutput string
	Command        string
	JoinCommand    string
	ExitCode       int
	// MRNumber is the merge request number where the comment is posted
	MRNumber int
	// Org is the GitHub Organization or User name
	Org string
	// Repo is the GitHub Repository name
	Repo string
	// SHA1 is the commit SHA1
	SHA1        string
	TemplateKey string
	Template    string
	Vars        map[string]interface{}
}

type Executor interface {
	Run(ctx context.Context, params *execute.Params) (*execute.Result, error)
}

type Expr interface {
	Match(expression string, params interface{}) (bool, error)
	Compile(expression string) (expr.Program, error)
}

func (ctrl *ExecController) getExecConfigs(cfg *config.Config, opts *option.ExecOptions) ([]*config.ExecConfig, error) {
	var execConfigs []*config.ExecConfig
	if opts.Template == "" && opts.TemplateKey != "" {
		a, ok := cfg.Exec[opts.TemplateKey]
		if !ok {
			if opts.TemplateKey != "default" {
				return nil, errors.New("template isn't found: " + opts.TemplateKey)
			}
			execConfigs = []*config.ExecConfig{
				{
					When:            "ExitCode != 0",
					UpdateCondition: `Comment.HasMeta && Comment.Meta.TemplateKey == "default"`,
					Template: `{{template "status" .}} {{template "link" .}}
{{template "join_command" .}}
{{template "hidden_combined_output" .}}`,
				},
			}
		} else {
			execConfigs = a
		}
	}
	return execConfigs, nil
}

// getExecConfig returns matched ExecConfig.
// If no ExecConfig matches, the second returned value is false.
func (ctrl *ExecController) getExecConfig(
	execConfigs []*config.ExecConfig, cmtParams *ExecCommentParams,
) (*config.ExecConfig, bool, error) {
	for _, execConfig := range execConfigs {
		f, err := ctrl.Expr.Match(execConfig.When, cmtParams)
		if err != nil {
			return nil, false, fmt.Errorf("test a condition is matched: %w", err)
		}
		if !f {
			continue
		}
		return execConfig, true, nil
	}
	return nil, false, nil
}

// getComment returns Comment.
// If the second returned value is false, no comment is posted.
func (ctrl *ExecController) getComment(ctx context.Context, execConfigs []*config.ExecConfig, cmtParams *ExecCommentParams, templates map[string]string) (*gitlab.Note, bool, error) { //nolint:funlen
	tpl := cmtParams.Template
	tplForTooLong := ""
	var embeddedVarNames []string
	var UpdateCondition string
	if tpl == "" {
		execConfig, f, err := ctrl.getExecConfig(execConfigs, cmtParams)
		if err != nil {
			return nil, false, err
		}
		if !f {
			return nil, false, nil
		}
		if execConfig.DontComment {
			return nil, false, nil
		}
		tpl = execConfig.Template
		tplForTooLong = execConfig.TemplateForTooLong
		embeddedVarNames = execConfig.EmbeddedVarNames
		UpdateCondition = execConfig.UpdateCondition
	}

	body, err := ctrl.Renderer.Render(tpl, templates, cmtParams)
	if err != nil {
		return nil, false, fmt.Errorf("render a comment template: %w", err)
	}
	bodyForTooLong, err := ctrl.Renderer.Render(tplForTooLong, templates, cmtParams)
	if err != nil {
		return nil, false, fmt.Errorf("render a comment template_for_too_long: %w", err)
	}

	noteCtrl := NoteController{
		Gitlab:   ctrl.Gitlab,
		Expr:     ctrl.Expr,
		Getenv:   ctrl.Getenv,
		Platform: ctrl.Platform,
	}

	embeddedMetadata := make(map[string]interface{}, len(embeddedVarNames))
	for _, name := range embeddedVarNames {
		if v, ok := cmtParams.Vars[name]; ok {
			embeddedMetadata[name] = v
		}
	}

	embeddedComment, err := noteCtrl.getEmbeddedComment(map[string]interface{}{
		"SHA1":        cmtParams.SHA1,
		"TemplateKey": cmtParams.TemplateKey,
		"Vars":        embeddedMetadata,
	})
	if err != nil {
		return nil, false, err
	}

	body += embeddedComment
	bodyForTooLong += embeddedComment

	note := gitlab.Note{
		MRNumber:       cmtParams.MRNumber,
		Org:            cmtParams.Org,
		Repo:           cmtParams.Repo,
		Body:           body,
		BodyForTooLong: bodyForTooLong,
		SHA1:           cmtParams.SHA1,
		Vars:           cmtParams.Vars,
		TemplateKey:    cmtParams.TemplateKey,
	}

	if UpdateCondition != "" && cmtParams.MRNumber != 0 {
		if err := ctrl.setUpdatedCommentID(ctx, &note, UpdateCondition); err != nil {
			return nil, false, fmt.Errorf("set updateCommentId: %w", err)
		}
	}

	return &note, true, nil
}

func (ctrl *ExecController) setUpdatedCommentID(ctx context.Context, note *gitlab.Note, updateCondition string) error { //nolint:funlen
	prg, err := ctrl.Expr.Compile(updateCondition)
	if err != nil {
		return err //nolint:wrapcheck
	}

	allnotes, err := ctrl.Gitlab.ListNote(ctx, &gitlab.MergeRequest{
		Org:      note.Org,
		Repo:     note.Repo,
		MRNumber: note.MRNumber,
	})

	if err != nil {
		return fmt.Errorf("list merge request notes: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"org":       note.Org,
		"repo":      note.Repo,
		"mr_number": note.MRNumber,
	}).Debug("get comments")

	for _, n := range allnotes {
		// if n.IsMinimized {
		// 	// ignore minimized comments
		// 	continue
		// }
		// if login != "" && comnt.Author.Login != login {
		// 	// ignore other users' comments
		// 	continue
		// }

		metadata := map[string]interface{}{}
		hasMeta := extractMetaFromComment(n.Body, &metadata)
		paramMap := map[string]interface{}{
			"Comment": map[string]interface{}{
				"Body":    n.Body,
				"Meta":    metadata,
				"HasMeta": hasMeta,
			},
			"Commit": map[string]interface{}{
				"Org":      note.Org,
				"Repo":     note.Repo,
				"MRNumber": note.MRNumber,
				"SHA1":     note.SHA1,
			},
			"Vars": note.Vars,
		}

		logrus.WithFields(logrus.Fields{
			"node_id":   n.ID,
			"condition": updateCondition,
			"param":     paramMap,
		}).Debug("judge whether an existing comment is ready for editing")
		f, err := prg.Run(paramMap)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"node_id": n.ID,
			}).Error("judge whether an existing comment is hidden")
			continue
		}
		if !f {
			continue
		}
		note.NoteID = n.ID
		break
	}
	return nil
}

func (ctrl *ExecController) post(
	ctx context.Context, execConfigs []*config.ExecConfig, cmtParams *ExecCommentParams,
	templates map[string]string,
) error {
	note, f, err := ctrl.getComment(ctx, execConfigs, cmtParams, templates)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"org":       note.Org,
		"repo":      note.Repo,
		"pr_number": note.MRNumber,
		"sha":       note.SHA1,
	}).Debug("comment meta data")

	noteCtrl := NoteController{
		Gitlab: ctrl.Gitlab,
		Expr:   ctrl.Expr,
		Getenv: ctrl.Getenv,
	}
	return noteCtrl.Post(ctx, note, map[string]interface{}{
		"Command": map[string]interface{}{
			"ExitCode":       cmtParams.ExitCode,
			"JoinCommand":    cmtParams.JoinCommand,
			"Command":        cmtParams.Command,
			"Stdout":         cmtParams.Stdout,
			"Stderr":         cmtParams.Stderr,
			"CombinedOutput": cmtParams.CombinedOutput,
		},
	})
}
