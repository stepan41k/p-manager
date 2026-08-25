##  Description of changes
<!-- Briefly describe the changes made and the problem they solve -->

##  Related Issues
<!-- Specify the issue numbers that this PR resolves.
Using the keyword "Closes #" or "Fixes #" will automatically close the issue upon merging -->
Closes #

##  Type of changes
<!-- Mark the appropriate option with an X: [x] -->
- [ ]  Bug fix
- [ ]  Feature
- [ ]  Refactoring / Optimization
- [ ]  Documentation update
- [ ]  CI/CD / Dependency changes (Chore)
- [ ]  Breaking change

##  Screenshots / Terminal recording
<!-- If the PR changes the TUI interface (Bubbletea/Lipgloss), please include a screenshot or GIF showing it in action -->

##  Pre-PR submission checklist
<!-- Please check all items before requesting a review -->
- [ ] The project builds successfully locally (`go build ./...`)
- [ ] All unit tests pass without errors (`go test -race ./...`)
- [ ] The linter reports no warnings (`golangci-lint run`)
- [ ] I have added tests for the new code (if applicable)
- [ ] Documentation / TUI keybindings have been updated (if changed)
- [ ] **Security**: I have verified that the logs and debug output do NOT contain passwords, S3 tokens, or master keys.