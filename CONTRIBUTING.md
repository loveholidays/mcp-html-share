# Contributing to MCP HTML Share

Thank you for your interest in contributing to MCP HTML Share! We welcome contributions from the community.

## How to Contribute

### Reporting Issues

If you find a bug or have a feature request, please open an issue on GitHub. When reporting issues, please include:

- A clear description of the issue
- Steps to reproduce the problem
- Expected behavior
- Actual behavior
- Your environment (OS, Go version, etc.)

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes**:
   - Write clear, concise commit messages
   - Add tests for new functionality
   - Update documentation as needed
3. **Test your changes**:
   ```bash
   make test
   make check
   ```
4. **Submit a pull request**:
   - Provide a clear description of the changes
   - Reference any related issues
   - Ensure all CI checks pass

### Development Setup

1. Clone your fork:
   ```bash
   git clone https://github.com/your-username/mcp-html-share.git
   cd mcp-html-share
   ```

2. Install dependencies:
   ```bash
   go mod download
   make tools
   ```

3. Run tests:
   ```bash
   make test
   ```

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Run `make check` to ensure code quality
- Write meaningful variable and function names
- Add comments for complex logic

### Testing

- Write unit tests for new functionality
- Ensure existing tests pass
- Aim for good test coverage
- Use table-driven tests where appropriate

### Documentation

- Update the README.md if you change functionality
- Add godoc comments to exported functions and types
- Include examples in documentation where helpful

## Questions?

Feel free to open an issue for any questions about contributing.

Thank you for contributing to MCP HTML Share!