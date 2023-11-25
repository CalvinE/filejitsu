set BUILDDATE=%date:~10,4%%date:~7,2%%date:~4,2%
echo %BUILDDATE%

for /f %%i in ('git rev-parse --short HEAD') do set BUIDHASH=%%i
for /f %%i in ('git describe --tags %BUILDHASH%') do set BUILDTAG=%%i

echo "building filejitsu version: Hash=%BUIDHASH% Time=%BUILDDATE% Tag=%BUILDTAG%"

go build -ldflags="-X main.commitHash=%BUIDHASH% -X main.buildDate=%BUILDDATE% -X main.buildTag=%BUILDTAG%" -o "filejitsu.exe" .

echo "finished"
