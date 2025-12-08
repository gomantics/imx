# Codecov Setup Guide

This document explains how to set up Codecov for the imx project.

## Prerequisites

- Repository owner/admin access
- GitHub account

## Setup Steps

### 1. Sign up for Codecov

1. Go to [https://codecov.io](https://codecov.io)
2. Click "Sign up with GitHub"
3. Authorize Codecov to access your GitHub account

### 2. Add the Repository

1. Once logged in, click "Add new repository"
2. Find `gomantics/imx` in the list
3. Click to activate it

### 3. Get the Upload Token

1. After activating, Codecov will show you a repository token
2. Copy this token (it looks like: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)

### 4. Add Token to GitHub Secrets

1. Go to your repository on GitHub: `https://github.com/gomantics/imx`
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `CODECOV_TOKEN`
5. Value: Paste the token from step 3
6. Click **Add secret**

### 5. Verify Setup

Once the token is added:

1. Push a commit to trigger CI
2. Wait for the coverage job to complete
3. Check Codecov dashboard: `https://codecov.io/gh/gomantics/imx`
4. You should see coverage reports and graphs

## What Codecov Provides

- **Coverage Reports**: Detailed line-by-line coverage
- **Pull Request Comments**: Automatic coverage diff comments on PRs
- **Historical Graphs**: Coverage trends over time
- **Badge**: Coverage badge for README

## Configuration

The project uses `codecov.yml` with these settings:

- **Target**: 100% coverage required
- **Ignored**: `cmd/`, `examples/`, test files
- **PR Comments**: Enabled with diff and file list

## Troubleshooting

### Upload Fails

If upload fails, check:
1. Token is correctly added to GitHub Secrets
2. Token name is exactly `CODECOV_TOKEN`
3. Check CI logs for error messages

### Coverage Shows Wrong Numbers

1. Ensure `codecov.yml` ignore patterns are correct
2. Verify `coverage.out` file is being generated
3. Check that only library code is being measured (not cmd/examples)

## Alternative: Skip Codecov

If you don't want to use Codecov:

1. The CI will still run coverage tests locally
2. 100% coverage is still enforced via `make coverage`
3. You can remove the Codecov upload step from `.github/workflows/ci.yml`

The coverage validation works independently of Codecov - it's just a nice-to-have for visualization.
