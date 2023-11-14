set BUILDDATE=%date:~10,4%%date:~7,2%%date:~4,2%
echo %BUILDDATE%

for /f %%i in ('git rev-parse --short HEAD') do set BUIDHASH=%%i

echo "building filejitsu version: Hash=%BUIDHASH% Time=%BUILDDATE%"

go build -ldflags="-X main.commitHash=%BUIDHASH% -X main.buildTime=%BUILDDATE%" -o "filejitsu" .

echo "finished"
