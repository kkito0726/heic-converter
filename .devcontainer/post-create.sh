#!/bin/sh
set -eu

go mod download
go install golang.org/x/tools/gopls@v0.22.0
curl -fsSL https://claude.ai/install.sh | bash
