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
)

type HideController struct {
	// Wd is a path to the working directory
	Wd string
	// Getenv returns the environment variable. os.Getenv
	Getenv func(string) string
	// HasStdin returns true if there is the standard input
	// If thre is the standard input, it is treated as the comment template
	HasStdin func() bool
	Stderr   io.Writer
	GitLab   GitLab
	Platform Platform
	Config   *config.Config
	Expr     Expr
}

func (ctrl *HideController) Hide(ctx context.Context, opts *option.HideOptions) error {
	logE := logrus.WithFields(logrus.Fields{
		"program": "gitlab-comment",
	})
	param, err := ctrl.getParamListHiddenComments(opts)
	if err != nil {
		return err
	}
	nodeIDs, err := listHiddenComments(
		ctrl.GitLab, ctrl.Expr, param, nil)
	if err != nil {
		return err
	}
	logE.WithFields(logrus.Fields{
		"count":    len(nodeIDs),
		"node_ids": nodeIDs,
	}).Debug("comments which would be hidden")
	hideComments(ctrl.GitLab, nodeIDs)
	return nil
}

func (ctrl *HideController) getParamListHiddenComments(opts *option.HideOptions) (*ParamListHiddenComments, error) { //nolint:cyclop,funlen
	param := &ParamListHiddenComments{}
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementHide(opts); err != nil {
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

	cfg := ctrl.Config

	if cfg.Base != nil {
		if opts.Org == "" {
			opts.Org = cfg.Base.Org
		}
		if opts.Repo == "" {
			opts.Repo = cfg.Base.Repo
		}
	}

	if err := option.ValidateHide(opts); err != nil {
		return param, fmt.Errorf("opts is invalid: %w", err)
	}

	hideCondition := opts.Condition
	if hideCondition == "" {
		a, ok := ctrl.Config.Hide[opts.HideKey]
		if !ok {
			return param, errors.New("invalid hide-key: " + opts.HideKey)
		}
		hideCondition = a
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]interface{}, len(opts.Vars))
	}
	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}

	return &ParamListHiddenComments{
		MRNumber:  opts.MRNumber,
		Org:       opts.Org,
		Repo:      opts.Repo,
		SHA1:      opts.SHA1,
		Condition: hideCondition,
		HideKey:   opts.HideKey,
		Vars:      cfg.Vars,
	}, nil
}

func hideComments(gl GitLab, nodeIDs []int) {
	logE := logrus.WithFields(logrus.Fields{
		"program": "gitlab-comment",
	})
	commentHidden := false
	for _, nodeID := range nodeIDs {
		if err := gl.HideComment(nodeID); err != nil {
			logE.WithError(err).WithFields(logrus.Fields{
				"node_id": nodeID,
			}).Error("hide an old comment")
			continue
		}
		commentHidden = true
		logE.WithFields(logrus.Fields{
			"node_id": nodeID,
		}).Info("hide an old comment")
	}
	if !commentHidden {
		logE.Info("no comment is hidden")
	}
}

type ParamListHiddenComments struct {
	Condition string
	HideKey   string
	Org       string
	Repo      string
	SHA1      string
	MRNumber  int
	Vars      map[string]interface{}
}

func listHiddenComments( //nolint:funlen
	gl GitLab, exp Expr,
	param *ParamListHiddenComments,
	paramExpr map[string]interface{},
) ([]int, error) {
	logE := logrus.WithFields(logrus.Fields{
		"program": "gitlab-comment",
	})
	if param.Condition == "" {
		logE.Debug("the condition to hide comments isn't set")
		return nil, nil
	}

	allnotes, err := gl.ListNote(&gitlab.MergeRequest{
		Org:      param.Org,
		Repo:     param.Repo,
		MRNumber: param.MRNumber,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	logE.WithFields(logrus.Fields{
		"count":     len(allnotes),
		"org":       param.Org,
		"repo":      param.Repo,
		"mr_number": param.MRNumber,
	}).Debug("get comments")

	nodeIDs := []int{}
	prg, err := exp.Compile(param.Condition)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	for _, note := range allnotes {
		nodeID := note.ID

		metadata := map[string]interface{}{}
		hasMeta := extractMetaFromComment(note.Body, &metadata)
		paramMap := map[string]interface{}{
			"Comment": map[string]interface{}{
				"Body": note.Body,
				// "CreatedAt": comment.CreatedAt,
				"Meta":    metadata,
				"HasMeta": hasMeta,
			},
			"Commit": map[string]interface{}{
				"Org":      param.Org,
				"Repo":     param.Repo,
				"MRNumber": param.MRNumber,
				"SHA1":     param.SHA1,
			},
			"HideKey": param.HideKey,
			"Vars":    param.Vars,
		}
		for k, v := range paramExpr {
			paramMap[k] = v
		}

		logE.WithFields(logrus.Fields{
			"node_id":   nodeID,
			"condition": param.Condition,
			"param":     paramMap,
		}).Debug("judge whether an existing note is hidden")
		f, err := prg.Run(paramMap)
		if err != nil {
			logE.WithError(err).WithFields(logrus.Fields{
				"node_id": nodeID,
			}).Error("judge whether an existing note is hidden")
			continue
		}
		if !f {
			continue
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, nil
}
