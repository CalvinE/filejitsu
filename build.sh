COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_DATE=$(date '+%Y%m%d')

echo "building filejitsu version: Hash=$COMMIT_HASH Date=$BUILD_DATE"

go build -ldflags="-X main.commitHash=$COMMIT_HASH -X main.buildDate=$BUILD_DATE" -o "filejitsu" .

echo "finished"