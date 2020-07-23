package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type PostTemplateParams struct {
	PRNumber    int
	Org         string
	Repo        string
	SHA1        string
	TemplateKey string
}

type Commenter interface {
	Create(ctx context.Context, cmt comment.Comment) error
}

type Reader interface {
	Find(wd string) (string, bool, error)
	Read(p string, cfg *config.Config) error
}

type PostController struct {
	Wd        string
	Getenv    func(string) string
	HasStdin  func() bool
	Stdin     io.Reader
	Reader    Reader
	Commenter Commenter
}

func (ctrl PostController) Post(ctx context.Context, opts *option.PostOptions) error {
	cmt, err := ctrl.getCommentParams(ctx, opts)
	if err != nil {
		return err
	}
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}

func (ctrl PostController) getCommentParams(ctx context.Context, opts *option.PostOptions) (comment.Comment, error) {
	cmt := comment.Comment{}
	if option.IsCircleCI(ctrl.Getenv) {
		if err := option.ComplementPost(opts, ctrl.Getenv); err != nil {
			return cmt, fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}
	if opts.Template == "" {
		tpl, err := ctrl.readTemplateFromStdin()
		if err != nil {
			return cmt, err
		}
		opts.Template = tpl
	}

	if err := option.ValidatePost(opts); err != nil {
		return cmt, fmt.Errorf("opts is invalid: %w", err)
	}

	if opts.Template == "" {
		if err := ctrl.readTemplateFromConfig(opts); err != nil {
			return cmt, err
		}
	}

	if err := ctrl.render(opts); err != nil {
		return cmt, err
	}

	return comment.Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     opts.Template,
		SHA1:     opts.SHA1,
	}, nil
}

func (ctrl PostController) readTemplateFromStdin() (string, error) {
	if !ctrl.HasStdin() {
		return "", nil
	}
	b, err := ioutil.ReadAll(ctrl.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read standard input: %w", err)
	}
	return string(b), nil
}

func (ctrl PostController) readTemplateFromConfig(opts *option.PostOptions) error {
	cfg := &config.Config{}
	if opts.ConfigPath == "" {
		p, b, err := ctrl.Reader.Find(ctrl.Wd)
		if err != nil {
			return err
		}
		if !b {
			return errors.New("configuration file isn't found")
		}
		opts.ConfigPath = p
	}
	if err := ctrl.Reader.Read(opts.ConfigPath, cfg); err != nil {
		return err
	}
	if t, ok := cfg.Post[opts.TemplateKey]; ok {
		opts.Template = t
		return nil
	}
	return errors.New("the template " + opts.TemplateKey + " isn't found")
}

func (ctrl PostController) render(opts *option.PostOptions) error {
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"Env": ctrl.Getenv,
	}).Parse(opts.Template)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, &PostTemplateParams{
		PRNumber:    opts.PRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
	}); err != nil {
		return err
	}
	opts.Template = buf.String()
	return nil
}
