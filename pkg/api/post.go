package api

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/yuyaban/gitlab-comment/pkg/config"
	"github.com/yuyaban/gitlab-comment/pkg/gitlab"
	"github.com/yuyaban/gitlab-comment/pkg/option"
	"github.com/yuyaban/gitlab-comment/pkg/template"
)

type PostController struct {
	// Wd is a path to the working directory
	Wd string
	// Getenv returns the environment variable. os.Getenv
	Getenv func(string) string
	// HasStdin returns true if there is the standard input
	// If thre is the standard input, it is treated as the comment template
	HasStdin func() bool
	Stdin    io.Reader
	Stderr   io.Writer
	GitLab   GitLab
	Renderer Renderer
	Platform Platform
	Config   *config.Config
	Expr     Expr
}

func (ctrl *PostController) Post(ctx context.Context, opts *option.PostOptions) error {
	note, err := ctrl.getCommentParams(opts)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"org":       note.Org,
		"repo":      note.Repo,
		"mr_number": note.MRNumber,
		"sha":       note.SHA1,
	}).Debug("note meta data")

	noteCtrl := NoteController{
		GitLab: ctrl.GitLab,
		Expr:   ctrl.Expr,
		Getenv: ctrl.Getenv,
	}
	return noteCtrl.Post(ctx, note, nil)
}

func (ctrl *PostController) setUpdatedCommentID(note *gitlab.Note, updateCondition string) error {
	customUpdateCondition := fmt.Sprintf("%s && Comment.Meta.Vars.target == \"%s\"", updateCondition, ctrl.Config.Vars["target"])
	prg, err := ctrl.Expr.Compile(customUpdateCondition)
	if err != nil {
		return err //nolint:wrapcheck
	}

	allnotes, err := ctrl.GitLab.ListNote(&gitlab.MergeRequest{
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
		note.ID = n.ID
		break
	}
	return nil
}

type Reader interface {
	FindAndRead(cfgPath, wd string) (config.Config, error)
}

type Renderer interface {
	Render(tpl string, templates map[string]string, params interface{}) (string, error)
}

type PostTemplateParams struct {
	// MRNumber is the merge request number where the comment is posted
	MRNumber int
	// Org is the GitHub Organization or User name
	Org string
	// Repo is the GitHub Repository name
	Repo string
	// SHA1 is the commit SHA1
	SHA1        string
	TemplateKey string
	Vars        map[string]interface{}
}

type Platform interface {
	ComplementPost(opts *option.PostOptions) error
	ComplementExec(opts *option.ExecOptions) error
	ComplementHide(opts *option.HideOptions) error
	CI() string
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func (ctrl *PostController) getCommentParams(opts *option.PostOptions) (*gitlab.Note, error) { //nolint:funlen,cyclop,gocognit
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementPost(opts); err != nil {
			return nil, fmt.Errorf("failed to complement opts with platform built in environment variables: %w", err)
		}
	}

	if opts.MRNumber == 0 && opts.SHA1 != "" {
		mrNum, err := ctrl.GitLab.MRNumberWithSHA(opts.Org, opts.Repo, opts.SHA1)
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

	if opts.Template == "" && opts.StdinTemplate {
		tpl, err := ctrl.readTemplateFromStdin()
		if err != nil {
			return nil, err
		}
		opts.Template = tpl
	}

	cfg := ctrl.Config

	if opts.Org == "" {
		opts.Org = cfg.Base.Org
	}
	if opts.Repo == "" {
		opts.Repo = cfg.Base.Repo
	}

	if err := option.ValidatePost(opts); err != nil {
		return nil, fmt.Errorf("opts is invalid: %w", err)
	}

	if opts.Template == "" {
		tpl, err := ctrl.readTemplateFromConfig(cfg, opts.TemplateKey)
		if err != nil {
			return nil, err
		}
		opts.Template = tpl.Template
		opts.TemplateForTooLong = tpl.TemplateForTooLong
		opts.EmbeddedVarNames = tpl.EmbeddedVarNames
		if !contains(opts.EmbeddedVarNames, "target") {
			opts.EmbeddedVarNames = append(opts.EmbeddedVarNames, "target")
		}
		if opts.UpdateCondition == "" {
			opts.UpdateCondition = tpl.UpdateCondition
		}
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]interface{}, len(opts.Vars))
	}
	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}
	if cfg.Vars["target"] == nil {
		cfg.Vars["target"] = ""
	}

	ci := ""
	if ctrl.Platform != nil {
		ci = ctrl.Platform.CI()
	}
	templates := template.GetTemplates(&template.ParamGetTemplates{
		Templates: cfg.Templates,
		CI:        ci,
	})
	tpl, err := ctrl.Renderer.Render(opts.Template, templates, PostTemplateParams{
		MRNumber:    opts.MRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
		Vars:        cfg.Vars,
	})
	if err != nil {
		return nil, fmt.Errorf("render a template for post: %w", err)
	}
	tplForTooLong, err := ctrl.Renderer.Render(opts.TemplateForTooLong, templates, PostTemplateParams{
		MRNumber:    opts.MRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
		Vars:        cfg.Vars,
	})
	if err != nil {
		return nil, fmt.Errorf("render a template template_for_too_long for post: %w", err)
	}

	noteCtrl := NoteController{
		GitLab:   ctrl.GitLab,
		Expr:     ctrl.Expr,
		Getenv:   ctrl.Getenv,
		Platform: ctrl.Platform,
	}
	embeddedMetadata := make(map[string]interface{}, len(opts.EmbeddedVarNames))
	for _, name := range opts.EmbeddedVarNames {
		if v, ok := cfg.Vars[name]; ok {
			embeddedMetadata[name] = v
		}
	}
	embeddedComment, err := noteCtrl.getEmbeddedComment(map[string]interface{}{
		"SHA1":        opts.SHA1,
		"TemplateKey": opts.TemplateKey,
		"Vars":        embeddedMetadata,
	})
	if err != nil {
		return nil, err
	}

	tpl += embeddedComment
	tplForTooLong += embeddedComment

	note := gitlab.Note{
		MRNumber:       opts.MRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		Body:           tpl,
		BodyForTooLong: tplForTooLong,
		SHA1:           opts.SHA1,
		HideOldComment: opts.HideOldComment,
		Vars:           cfg.Vars,
		TemplateKey:    opts.TemplateKey,
	}
	if opts.UpdateCondition != "" && opts.MRNumber != 0 {
		if err := ctrl.setUpdatedCommentID(&note, opts.UpdateCondition); err != nil {
			return nil, err
		}
	}
	return &note, nil
}

func (ctrl *PostController) readTemplateFromStdin() (string, error) {
	if !ctrl.HasStdin() {
		return "", nil
	}
	b, err := io.ReadAll(ctrl.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read standard input: %w", err)
	}
	return string(b), nil
}

func (ctrl *PostController) readTemplateFromConfig(cfg *config.Config, key string) (*config.PostConfig, error) {
	if t, ok := cfg.Post[key]; ok {
		return t, nil
	}
	return nil, errors.New("the template " + key + " isn't found")
}
