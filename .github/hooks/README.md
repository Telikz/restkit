# Git Hooks

This directory contains git hooks that can be used to enforce code quality before commits.

## Available Hooks

### pre-commit

Runs `go fmt` and `go test` before allowing commits.

## Installation

To use these hooks locally, copy them to your `.git/hooks` directory:

```bash
cp .github/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

Or use a symlink (recommended for keeping up to date):

```bash
ln -s ../../.github/hooks/pre-commit .git/hooks/pre-commit
chmod +x .github/hooks/pre-commit
```
