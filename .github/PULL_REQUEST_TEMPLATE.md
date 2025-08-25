<!--
Thanks for submitting a pull request! Please provide enough information so that others can review your pull request.

For more information about pull requests, please read:
1. The contributing guide: [CONTRIBUTING.md](../CONTRIBUTING.md)
2. The development guide: [IMPLEMENTATION_GUIDE.md](../IMPLEMENTATION_GUIDE.md)
-->

## PR Type
<!--
What kind of change does this PR introduce?
Please check the one that applies to this PR using "x".
-->
- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] 🧹 Code cleanup or refactor
- [ ] 🔧 Configuration change
- [ ] 🚀 Performance improvement
- [ ] ✅ Test update

## What does this PR do?
<!-- 
Provide a clear and concise description of what this PR accomplishes.
Include any technical details that would help reviewers understand your changes.
-->

## Related Issues
<!--
Link any related issues here using the GitHub issue syntax:
- Fixes #<issue_number>
- Related to #<issue_number>

If there are no related issues, write "N/A".
-->

## Breaking Changes
<!--
List any breaking changes and migration steps required when users update.
If there are no breaking changes, write "NONE".

Example:
- Changed API endpoint `/api/v1/servers` to `/api/v2/servers`
- Migration required: Update client configurations to use new endpoint
-->

## Implementation Notes
<!--
Include any important implementation details, architectural decisions,
or technical considerations that would help reviewers understand your changes.
-->

## Testing Done
<!--
Describe the testing you have performed:
1. Unit tests added/modified
2. Integration tests added/modified
3. Manual testing steps performed
-->

## Documentation
<!--
List any documentation updates required:
1. README.md changes
2. API documentation updates
3. New example files or updates
4. Architecture/ERD diagram updates
-->

## Screenshots/Recordings
<!--
For UI/UX changes, include before/after screenshots or recordings.
For API changes, include example requests/responses.
Delete this section if not applicable.
-->

## Security Considerations
<!--
If this PR has security implications, please address the following:
-->
- [ ] No sensitive data (passwords, keys, tokens) is exposed
- [ ] Input validation is implemented where necessary
- [ ] Error messages don't leak sensitive information
- [ ] SQL injection protection is maintained
- [ ] Authentication/authorization is properly handled
- [ ] Security tests have been run (`make security`)

## Checklist
<!--
Put an "x" in the boxes that apply. You can also fill these out after creating the PR.
-->
- [ ] I have read the [CONTRIBUTING.md](../CONTRIBUTING.md) document
- [ ] I have updated the documentation accordingly
- [ ] I have added tests to cover my changes
- [ ] All new and existing tests passed (`make test`)
- [ ] Code passes linting checks (`make lint`)
- [ ] Security scan passes (`make security`)
- [ ] My code follows the code style of this project
- [ ] I have updated the database schema if necessary
- [ ] I have added example usage if applicable
- [ ] Pre-commit hooks pass (`make precommit-all`)

## Special Notes for Reviewer
<!--
Add any additional notes for reviewers to consider while reviewing your changes.
Delete this section if not applicable.
-->

## Release Notes
<!--
Write user-facing release notes for this change.
If this PR requires action from users switching to the new release, include "action required".
If no release notes are required, write "NONE".

Example:
```release-note
Added virtual server feature to wrap non-MCP services as MCP-compatible servers
```
-->
```release-note

```

## Additional Documentation
<!--
Link to any additional documentation you've created or updated.
Use the following format:
- [Doc Name]: <link>

Example:
- [Virtual Servers Example]: examples/virtual_servers_example.md
- [Database ERD]: DATABASE_ERD.md#virtual-servers
-->
```docs

```