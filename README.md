# ASPNet Cloud Native Buildpack

The Dotnet Core ASPNet CNB provides a version of the [Dotnet Core
ASPNet Framework](https://github.com/aspnet) and sets an extension of the `$DOTNET_ROOT`
location.

A usage example can be found in the
[`samples` repository under the `dotnet-core/aspnet`
directory](https://github.com/paketo-buildpacks/samples/tree/main/dotnet-core/aspnet).

## Integration

The Dotnet Core ASPNet CNB provides `dotnet-aspnetcore` as a dependency.
Downstream buildpacks, like [Dotnet Core
Build](https://github.com/paketo-buildpacks/dotnet-core-build) and [Dotnet
Core SDK](https://github.com/paketo-buildpacks/dotnet-core-sdk) can require the
dotnet-aspnetcore dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the ASPNet dependency is "dotnet-aspnetcore". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "dotnet-aspnetcore"

  # The ASPNet buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the ASPNet
    # dependency is available to subsequent buildpacks during their build phase.
    # Currently we do not recommend having your application directly interface with
    # the framework, instead use the dotnet-core-sdk. However,
    # if you are writing a buildpack that needs to use the ASPNet during
    # its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the ASPNet
    # dependency is available on the $DOTNET_ROOT for the running application. If you are
    # writing an application that needs to run ASPNet at runtime, this flag should
    # be set to true.
    launch = true

    # The version of the ASPNet dependency is not required. In the case it
    # is not specified, the buildpack will provide the default version, which can
    # be seen in the buildpack.toml file.
    # If you wish to request a specific version, the buildpack supports
    # specifying a semver constraint in the form of "2.*", "2.1.*", or even
    # "2.1.15".
    version = "2.1.15"
```

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
