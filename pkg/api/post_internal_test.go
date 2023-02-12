package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yuyaban/gitlab-comment/pkg/config"
	"github.com/yuyaban/gitlab-comment/pkg/gitlab"
	"github.com/yuyaban/gitlab-comment/pkg/option"
	"github.com/yuyaban/gitlab-comment/pkg/template"
)

func TestPostController_getCommentParams(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		ctrl  *PostController
		exp   *gitlab.Note
		isErr bool
		opts  *option.PostOptions
	}{
		{
			title: "if there is a standard input, treat it as the template",
			ctrl: &PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(k string) string {
					return ""
				},
				Renderer: &template.Renderer{},
				Config:   &config.Config{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:      "yuyaban",
					Repo:     "gitlab-comment",
					Token:    "xxx",
					MRNumber: 1,
				},
				StdinTemplate: true,
			},
			exp: &gitlab.Note{
				Org:      "yuyaban",
				Repo:     "gitlab-comment",
				MRNumber: 1,
				Vars:     map[string]interface{}{"target": ""},
			},
		},
		{
			title: "if template is passed as argument, standard input is ignored",
			ctrl: &PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(k string) string {
					return ""
				},
				Renderer: &template.Renderer{},
				Config:   &config.Config{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:      "yuyaban",
					Repo:     "gitlab-comment",
					Token:    "xxx",
					MRNumber: 1,
					Template: "foo",
				},
			},
			exp: &gitlab.Note{
				Org:      "yuyaban",
				Repo:     "gitlab-comment",
				MRNumber: 1,
				Vars:     map[string]interface{}{"target": ""},
			},
		},
		{
			title: "read template from config",
			ctrl: &PostController{
				HasStdin: func() bool {
					return false
				},
				Getenv: func(k string) string {
					return ""
				},
				Config: &config.Config{
					Post: map[string]*config.PostConfig{
						"default": {
							Template: "hello",
						},
					},
				},
				Renderer: &template.Renderer{
					Getenv: func(k string) string {
						return ""
					},
				},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:         "yuyaban",
					Repo:        "gitlab-comment",
					Token:       "xxx",
					TemplateKey: "default",
					MRNumber:    1,
				},
			},
			exp: &gitlab.Note{
				Org:         "yuyaban",
				Repo:        "gitlab-comment",
				MRNumber:    1,
				TemplateKey: "default",
				Vars:        map[string]interface{}{"target": ""},
			},
		},
		{
			title: "template is rendered properly",
			ctrl: &PostController{
				HasStdin: func() bool {
					return false
				},
				Getenv: func(k string) string {
					return ""
				},
				Renderer: &template.Renderer{
					Getenv: func(k string) string {
						if k == "FOO" {
							return "BAR"
						}
						return ""
					},
				},
				Config: &config.Config{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:      "yuyaban",
					Repo:     "gitlab-comment",
					Token:    "xxx",
					MRNumber: 1,
					Template: `{{.Org}} {{.Repo}} {{.MRNumber}}`,
				},
			},
			exp: &gitlab.Note{
				Org:      "yuyaban",
				Repo:     "gitlab-comment",
				MRNumber: 1,
				Vars:     map[string]interface{}{"target": ""},
			},
		},
		{
			title: "config.base",
			ctrl: &PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(k string) string {
					return ""
				},
				Config: &config.Config{
					Base: &config.Base{
						Org:  "yuyaban",
						Repo: "gitlab-comment",
					},
				},
				Renderer: &template.Renderer{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Token:    "xxx",
					MRNumber: 1,
				},
				StdinTemplate: true,
			},
			exp: &gitlab.Note{
				Org:      "yuyaban",
				Repo:     "gitlab-comment",
				MRNumber: 1,
				Vars:     map[string]interface{}{"target": ""},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			cmt, err := d.ctrl.getCommentParams(d.opts)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			cmt.Body = ""
			cmt.BodyForTooLong = ""
			require.Equal(t, d.exp, cmt)
		})
	}
}

func TestPostController_readTemplateFromStdin(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		ctrl  PostController
		exp   string
		isErr bool
	}{
		{
			title: "no standard input",
			ctrl: PostController{
				HasStdin: func() bool {
					return false
				},
			},
		},
		{
			title: "standard input",
			ctrl: PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
			},
			exp: "hello",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			tpl, err := d.ctrl.readTemplateFromStdin()
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, tpl)
		})
	}
}
