
## Creating a release

### Determine version number

Ensure you're on master and up to date with the remote, then git desccribe should returns the latest tag:

```bash
$ git checkout master 
$ git pull origin master
$ git describe
v0.15.0
```

Here our current version is `v0.15.0`, and we want to create the `v0.15.1` release. We'll use this number in the rest of the document.

### Update version

Create the tag:

```bash
git tag -a v0.15.1 -m v0.15.1

git push origin v0.15.1
```

### Create release on github

Now head to [the release page](https://github.com/fetchai/cosmos-sdk/releases) and you must see the tag you just pushed there.

Edit it and:

- set the release title to the version number (here `v0.15.1`)
- Update the description from the following template. Remember to update the ecosystem versions if they did change:

```markdown
## Changes in this release

* Main change 1
* Main change 2
* ...

## Ecosystem

| Component  | Version                                                                  | Baseline |
| ---------- | ------------------------------------------------------------------------ | -------- |
| Tendermint | [0.15.2](https://github.com/fetchai/cosmos-consensus/releases/tag/v0.15.2) | 0.33.6 |
| SDK        | -                                                                        | 0.39.1  |

## Pull Requests

* relevant PR 1
* relevant PR 2
* ...

```

- Tick the `This is a pre-release` box (until mainnet release)
- Hit `Publish release`


## Next steps

To use this new release, `fetchd` needs to be updated.

### Update fetchd

The [fetchd](https://github.com/fetchai/fetchd) `go.mod` file can be updated to change the `cosmos-sdk` version:

```diff
- replace github.com/cosmos/cosmos-sdk => github.com/fetchai/cosmos-sdk v0.15.0
+ replace github.com/cosmos/cosmos-sdk => github.com/fetchai/cosmos-sdk v0.15.1
```

then run `go mod tidy` to update the sum file.
