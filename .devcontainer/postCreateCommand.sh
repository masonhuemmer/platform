#!/bin/bash

# Configure Pre-Commit Hook
DIR=~/.git-template
git config --global init.templateDir ${DIR}
pre-commit init-templatedir -t pre-commit ${DIR}

pre-commit install
pre-commit run -a
