# Contributing to FlowForge

First off, thank you for considering contributing to FlowForge! It's people like you that make FlowForge such a great platform for distributed workflow orchestration.

## 🧭 How Can I Contribute?

### Reporting Bugs
Bugs are tracked as GitHub issues. When creating an issue, please provide the following information:
* A clear and descriptive title.
* FlowForge version, OS, and runtime environment.
* Steps to reproduce the behavior.
* Expected and actual behavior.
* Relevant logs, error messages, or screenshots.

### Suggesting Enhancements
Enhancement suggestions are also tracked as GitHub issues. Please provide:
* A clear description of the proposed feature.
* The specific use case and why this feature would be useful for the broader community.
* Proposed architectural implementation (if you have one).

### Pull Requests
1. **Fork the repo** and create your branch from `main`.
2. **If you've added code**, add tests to verify your implementation.
3. **Ensure the test suite passes** locally (`make test`).
4. **Format your code** using the project's linters (`make lint`).
5. **Update documentation** (README.md, or the `/docs` directory) if your change modifies user-facing behavior.
6. **Issue a Pull Request**. Ensure the PR description clearly describes the problem and solution.

## 🛠️ Development Setup

FlowForge is primarily written in Go, with the Web UI in React.

### Prerequisites
* Go 1.21+
* Node.js 18+ (for UI development)
* Docker & Docker Compose (for local database/queue infrastructure)
* `make`

### Local Environment Setup
1. Clone your fork:
   ```bash
   git clone [https://github.com/GauravKarakoti/flowforge.git](https://github.com/GauravKarakoti/flowforge.git)
   cd flowforge
   ```
2. Start the local infrastructure dependencies (PostgreSQL, Redis):
```bash
make up-infra
```
3. Run the backend tests:
```bash
make test-backend
```
4. Build and run the server locally:
```bash
make run
```

## 📝 Commit Guidelines
We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This leads to more readable messages that are easy to follow when looking through the project history and allows us to auto-generate changelogs.

Format: `<type>(<scope>): <subject>`

Examples:
- `feat(worker): add exponential backoff for task retries`
- `fix(dashboard): resolve memory leak in live log viewer`
- `docs(readme): update quickstart python example`
- `chore(deps): update prometheus client library`

## 🎨 Styleguides
Code Style
- Go: We adhere strictly to standard `gofmt`. Run `make lint` before committing to ensure `golangci-lint` passes.
- React/TypeScript: We use Prettier and ESLint. Run `npm run lint` in the `/ui` directory.
- Documentation: Keep sentences concise. Use Markdown formatting for commands, file names, and code blocks.

## Architecture Decisions
For any major architectural change (e.g., changing the database schema, altering the API contract between control plane and workers), please submit an RFC (Request for Comments) issue before writing code. This saves everyone time and ensures alignment with the project roadmap.

## ❓ Need Help?
If you need help setting up your development environment or figuring out how to approach a bug, open a Draft PR and ask questions there. We're happy to help!