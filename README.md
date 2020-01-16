# ASPNet Cloud Native Buildpack
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
