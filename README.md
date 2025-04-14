# Node Run Script Cloud Native Buildpack

## `docker.io/paketobuildpacks/node-run-script`

The Node Run Script CNB runs any arbitrary scripts desired for the given application. The scripts
run are determined by the environment variable `BP_NODE_RUN_SCRIPTS` and run scripts specified in
the contents of `package.json`. For example, given a `package.json` and `BP_NODE_RUN_SCRIPTS` var
with the following content:

```json
{
    "scripts": {
        "build": "<build-commands>",
        "deploy": "<deploy-commands>",
        "another-script": "<some-other-commands>"
    }
}
```

`BP_NODE_RUN_SCRIPTS="build,another-script"`

The scripts `build` and `another-script` will be run through `npm run-script` or `yarn run`.

## Integration

This CNB currently does not provide anything specific and is purposed primarily to run scripts in node framework apps, so there's no scenario we can imagine where you would need to require it as a dependency.

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh --version <version-number>
```

This will create a `buildpackage.cnb` file under the `build` directory which you
can use to build your app as follows:

```
pack build <app-name> -p <path-to-app> -b <path/to/node-engine.cnb> -b <path/to/yarn.cnb> -b build/buildpackage.cnb \
-b <path/to/node-and-yarn-requiring-cnb>
```

## Specifying a project path

To specify a project subdirectory to be used as the root of the app, please use
the `BP_NODE_PROJECT_PATH` environment variable at build time either directly
(e.g. `pack build my-app --env BP_NODE_PROJECT_PATH=./src/my-app`) or through a
[`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
This could be useful if your app is a part of a monorepo.

## Specifying the scripts to be run

To specify which scripts inside `package.json` you would like to run, please use the
`BP_NODE_RUN_SCRIPTS` environment variable at build time either directly or through a
[`project.toml` file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md). The value of the variable should be a comma separated list of events listed in the app's `package.json`

## Run Tests

To run all unit tests, run:

```
./scripts/unit.sh
```

To run all integration tests, run:

```
/scripts/integration.sh
```
