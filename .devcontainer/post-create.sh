#!/bin/sh
set -eu

(cd backend && go mod download)
go install golang.org/x/tools/gopls@v0.22.0
go install github.com/bufbuild/buf/cmd/buf@v1.71.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.20.0
(cd web && npm ci)
curl -fsSL https://claude.ai/install.sh | bash
