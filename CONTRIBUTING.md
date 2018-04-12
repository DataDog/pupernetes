# Contributing to pupernetes

First of all, thanks for contributing!

This document provides some basic guidelines for contributing to this repository.
To propose improvements, feel free to submit a PR.

## Submitting issues

  * You can open a Github issue.

## Pull Requests

Have you fixed a bug or written a new feature and want to share it? Many thanks!

In order to ease/speed up our review, here are some items you can check/improve
when submitting your PR:

  * have a [proper commit history](#commits) (we advise you to rebase if needed).
  * write tests for the code you wrote (if applicable)
  * preferably make sure that all tests pass locally.
  * summarize your PR with an explanatory title and a message describing your
    changes, cross-referencing any related bugs/PRs.

Your pull request must pass all CI tests before we will merge it. If you're seeing
an error and don't think it's your fault, it may not be! [Join us on Slack][slack]
or send  us an email, and together we'll get it sorted out.

### Commit Messages

Please don't be this person: `git commit -m "Fixed stuff"`. Take a moment to
write meaningful commit messages.

The commit message should describe the reason for the change and give extra details
that will allow someone later on to understand in 5 seconds the thing you've been
working on for a day.


### Squash your commits

Please rebase your changes on `master` and squash your commits whenever possible,
it keeps history cleaner and it's easier to revert things. It also makes developers
happier!


[slack]: http://datadoghq.slack.com