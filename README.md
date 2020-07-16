![github-projector](https://i.imgur.com/T0Z3qWL.png)
# projector

A github bot for managing projects.

## Features

1. Automatically moves new issues & pull requests to a default project board.

1. Specify custom label rules to move issues or pull requests to a desired project board.

1. Generate label based reports.

[
    {
        "ProjectBoard": "Kanban",
        "IssuesClosed": 26,
        "LabelCounts": [
            {
                "Name": "effort: 3",
                "Count": 1
            },
            {
                "Name": "priority: soon",
                "Count": 1
            },
            {
                "Name": "type: enhancement",
                "Count": 4
            },
            {
                "Name": "work: obvious",
                "Count": 11
            },
            {
                "Name": "effort: 1",
                "Count": 8
            },
            {
                "Name": "priority: now",
                "Count": 16
            },
            {
                "Name": "type: bug",
                "Count": 6
            },
            {
                "Name": "type: feature",
                "Count": 12
            },
            {
                "Name": "effort: 2",
                "Count": 1
            },
            {
                "Name": "effort: 8",
                "Count": 1
            },
            {
                "Name": "work: complicated",
                "Count": 1
            },
            {
                "Name": "state: duplicate",
                "Count": 1
            }
        ]
      }
    ]
```

## Setup

The following environment variables need to be configured

Name | Value | Notes
-- | -- | --
PRJ_ORG_NAME | secberus | The GitHub organization name
PRJ_DEFAULT_PROJECT | Kanban | The default project to send new PRS and issues.
PRJ_DEFAULT_COLUMN | To Do | The default column to place new issue and PRs on the project board.
PRJ_HOOK_URL | http://projector.your.domain.com/webhook | The public url of your service [WARNING: YOUR PRIVATE DATA WILL BE SENT HERE].
PRJ_HOOK_SECRET | Your_secret_key | The secret used to validate your webhook payloads.
PRJ_GITHUB_TOKEN | You_Github_Token | A service account token with organization admin privilege. Used to create organization webhook.
