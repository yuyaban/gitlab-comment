package platform

import (
	"fmt"
	"os"
	"strconv"

	"github.com/yuyaban/gitlab-comment/pkg/option"
)

type Platform struct {
	platformId string
}

func (pt *Platform) getRepoOrg() (string, error) { //nolint:unparam
	if org := os.Getenv("CI_PROJECT_NAMESPACE"); org != "" {
		return org, nil
	}

	return "", nil
}

func (pt *Platform) getRepoName() (string, error) { //nolint:unparam
	if repo := os.Getenv("CI_PROJECT_NAME"); repo != "" {
		return repo, nil
	}
	return "", nil
}

func (pt *Platform) getSHA1() (string, error) { //nolint:unparam
	if sha1 := os.Getenv("CI_COMMIT_SHA"); sha1 != "" {
		return sha1, nil
	}
	return "", nil
}

func (pt *Platform) getMRNumber() (int, error) {

	if mr := os.Getenv("CI_MERGE_REQUEST_IID"); mr != "" {
		a, err := strconv.Atoi(mr)
		if err != nil {
			return 0, fmt.Errorf("parse CI_MERGE_REQUEST_IID %s: %w", mr, err)
		}
		return a, nil
	}
	return 0, fmt.Errorf("parse CI_MERGE_REQUEST_IID")
}

func (pt *Platform) complement(opts *option.Options) error {
	if opts.Org == "" {
		org, err := pt.getRepoOrg()
		if err != nil {
			return err
		}
		opts.Org = org
	}
	if opts.Repo == "" {
		repo, err := pt.getRepoName()
		if err != nil {
			return err
		}
		opts.Repo = repo
	}
	if opts.SHA1 == "" {
		sha1, err := pt.getSHA1()
		if err != nil {
			return err
		}
		opts.SHA1 = sha1
	}
	if opts.MRNumber > 0 {
		return nil
	}
	pr, err := pt.getMRNumber()
	if err != nil {
		return err
	}
	opts.MRNumber = pr
	return nil
}

func (pt *Platform) ComplementPost(opts *option.PostOptions) error {
	return pt.complement(&opts.Options)
}

func (pt *Platform) ComplementHide(opts *option.HideOptions) error {
	return pt.complement(&opts.Options)
}

func (pt *Platform) CI() string {
	return pt.platformId
}

func (pt *Platform) ComplementExec(opts *option.ExecOptions) error {
	return pt.complement(&opts.Options)
}

func Get() *Platform {
	return &Platform{
		platformId: "gitlab-ci",
	}
}
