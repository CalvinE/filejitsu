COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_DATE=$(date '+%Y%m%d')
BUILD_TAG=$(git describe --tags $COMMIT_HASH)

echo "building filejitsu version: Hash=$COMMIT_HASH Date=$BUILD_DATE TAG=$BUILD_TAG"

go build -ldflags="-X main.commitHash=$COMMIT_HASH -X main.buildDate=$BUILD_DATE -X main.buildTag=$BUILD_TAG" -o "filejitsu" .

echo "finished"