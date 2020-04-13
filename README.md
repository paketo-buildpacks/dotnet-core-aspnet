# ASPNet Cloud Native Buildpack

## Integration

```toml
[[requires]]

  # The name of the ASPNet dependency is "dotnet-aspnetcore". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "dotnet-aspnetcore"

  # The version of the ASPNet dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "2.*", "2.1.*", or even
  # "2.1.15".
  version = "2.1.15"

  # The ASPNet buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the ASPNet
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run ASPNet
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the ASPNet
    # dependency is available on the $PATH for the running application. If you are
    # writing an application that needs to run ASPNet at runtime, this flag should
    # be set to true.
    launch = true
```

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## Development

### Generating a sample app

To generate a sample app (like the ones that live in `integration/testdata`:

```
app_name=my_sample_app
runtime_version=3.1

rm -rf "integration/testdata/$app_name"
mkdir "integration/testdata/$app_name"

docker run -v "$PWD/integration/testdata/$app_name:/app" -it mcr.microsoft.com/dotnet/core/sdk:"$runtime_version" \
  bash -c "
  mkdir /tmp/$app_name &&
    cd /tmp/$app_name &&
    dotnet new web &&
    dotnet build -c Release -o /app"
```
