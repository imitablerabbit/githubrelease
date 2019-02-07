# githubrelease

This repo contains a script that can create a GitHub release and upload tar.gz files from a specified directory.

## Table of Contents

- [Example](#example)
- [Command Line arguments](#command-line-arguments)

## Example

```bash 
./githubrelease \
    --api-url="https://api.github.com" --pat=${PAT} \
    --user="imitablerabbit" --repo="counterifficwebsite" \
    --release-tag="v0.0.1" --target="0.0.1" --name="v0.0.1" \
    --body="Initial version" --draft --prerelease=false \
    --uploads="./uploads/"
```

## Command Line arguments

| Name          | Type    | Description                                                                                                                                                                                                             |
|---------------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `api-url`     | string  | GitHub api base url. Typically this can be left off the list of arguments so that the latest api version is used.                                                                                                       |
| `pat`         | string  | GitHub personal access token to be used in the api requests.                                                                                                                                                            |
| `user`        | string  | This is actually the user namespace that the repo is located under, e.g. githubrelease is imitablerabbit/githubrelease, so the user is imitablerabbit.                                                                  |
| `repo`        | string  | The name of the repo as it appears on GitHub.                                                                                                                                                                           |
| `release-tag` | string  | This is the tag name for the release. This does not have to be the same as an actual git tag.                                                                                                                           |
| `target`      | string  | This is the `target_commitish` value that the api request requires. Essentially this is the commit, branch or tag that the release represents.                                                                          |
| `name`        | string  | The name of the release                                                                                                                                                                                                 |
| `body`        | string  | A description of the release, should probably include changelog information.                                                                                                                                            |
| `draft`       | boolean | Whether or not the release should be created as a draft. It is recommended that this is set to true. If a draft release is created, it remains invisible to the public but can checked and edited before making public. |
| `prerelease`  | boolean | Whether or not the release should be listed as a pre-release.                                                                                                                                                           |
| `uploads`     | string  | This is the directory that should contain the `.tar.gz` files to upload as part of the release. There should be nothing else in the folder other than the files to upload.                                              |