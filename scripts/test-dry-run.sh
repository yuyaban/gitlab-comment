#!/usr/bin/env bash

set -eux

cd "$(dirname "$0")/.."

export GITLAB_TOKEN=dummy
export HELLO=hello
export CI_PROJECT_NAMESPACE=yuyaban
export CI_PROJECT_NAME=gitlab-comment
export CI_MERGE_REQUEST_IID=1

go run ./cmd/gitlab-comment post --dry-run -k hello
HELLO=hello2 go run ./cmd/gitlab-comment post --dry-run -k hello -u 'Comment.HasMeta && Comment.Meta.TemplateKey == "hello"'
go run ./cmd/gitlab-comment exec --dry-run -k hello -- echo foo
go run ./cmd/gitlab-comment exec --dry-run -k hello -- false || true
