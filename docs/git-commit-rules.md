# Git Commit Message Rules

This project follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for all git commit messages. This ensures a consistent and readable project history, and allows for automated changelog generation and versioning.

## Structure

A commit message consists of a **header**, a **body**, and a **footer**. The header has a special format that includes a **type**, a **scope**, and a **subject**:

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Type

The type is mandatory and must be one of the following:

- **feat**: A new feature for the user, not a new feature for a build script.
- **fix**: A bug fix for the user, not a fix to a build script.
- **docs**: Changes to documentation.
- **style**: Formatting, missing semi colons, etc; no production code change.
- **refactor**: Refactoring production code, e.g. renaming a variable.
- **perf**: Code changes that improve performance.
- **test**: Adding missing tests, refactoring tests; no production code change.
- **build**: Changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm).
- **ci**: Changes to our CI configuration files and scripts (example scopes: Travis, Circle, BrowserStack, SauceLabs).
- **chore**: Other changes that don't modify src or test files.
- **revert**: Reverts a previous commit.

### Scope

The scope is optional and should be a noun describing a section of the codebase surrounded by parenthesis, e.g., `feat(parser): add support for new tags`.

### Description

The description is a short summary of the code changes.
- Use the imperative, present tense: "change", not "changed" nor "changes".
- Don't capitalize the first letter.
- No dot (.) at the end.

### Body

The body is optional and should include the motivation for the change and contrast this with previous behavior. It should be used for providing more context.

### Footer

The footer is optional and should be used for referencing GitHub issues or Pull Requests, or for breaking changes.

- **Breaking Changes**: All breaking changes must be explained in the footer. A breaking change is indicated by the text `BREAKING CHANGE:` followed by a space and the description.

## Examples

- `feat(api): add user authentication endpoint`
- `fix(ui): resolve button misalignment on mobile view`
- `docs: update installation instructions in README`
- `refactor: simplify database connection logic`
- `chore: update dependency versions`
