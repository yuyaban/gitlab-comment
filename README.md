# gitlab-comment

[![Build Status](https://github.com/yuyaban/gitlab-comment/workflows/test/badge.svg)](https://github.com/yuyaban/gitlab-comment/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yuyaban/gitlab-comment)](https://goreportcard.com/report/github.com/yuyaban/gitlab-comment)
[![GitHub last commit](https://img.shields.io/github/last-commit/yuyaban/gitlab-comment.svg)](https://github.com/yuyaban/gitlab-comment)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/yuyaban/gitlab-comment/main/LICENSE)

CLI to create GitLab Notes by GitLab REST API  
Fork of [suzuki-shunsuke/github-comment](https://github.com/suzuki-shunsuke/github-comment), supporting GitLab (dropped GitLab support).

## Document
### Prerequisites
* Create and store GitLab access token in  [project or group CI variables](https://docs.gitlab.com/ee/ci/variables/#add-a-cicd-variable-to-a-project) with key name `GITLAB_TOKEN` or `GITLAB_ACCESS_TOKEN`.  
  ref: [Project access tokens | GitLab](https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html)
* When using the post command, a default template is not provided, so you must provide configuration file.  
  The configuration file name must be one of the following.  
  * .gitlab-comment.yml (or gitlab-comment.yml)
  * .gitlab-comment.yaml (or gitlab-comment.yaml)

Basic Comamnds are follows:

```shell
# post
github-comment post -k hello
# post: Enable updateCondition
github-comment post -k hello -u 'Comment.HasMeta && Comment.Meta.TemplateKey == "hello"' --var target:"${CI_JOB_NAME}"

# exec
github-comment exec -k hello -- echo "this is comment"
# exec: Enable updateCondition
github-comment exec -k hello -u 'Comment.HasMeta && Comment.Meta.TemplateKey == "hello"' --var target:"${CI_JOB_NAME}" -- echo "this is comment"
```

A concrete example of gitlab-comment configuration running on GitLab CI can be found in [.gitlab-ci.yml](example.gitlab-ci.yml).

And, See also [the original documentation (suzuki-shunsuke/github-comment)](https://suzuki-shunsuke.github.io/github-comment/).
## License

[MIT](LICENSE)
