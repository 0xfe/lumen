#!/bin/bash

TAG=$1

if [ "$TAG" == "" ]; then
    echo Usage: $0 tag
    echo Current tag: `git tag | tail -n 1`
    exit 1
fi

if [ "$GITHUB_TOKEN" == "" ]; then
    echo error: GITHUB_TOKEN not set.
    exit 1
fi

git pull --tags
git status
echo
echo Current tag: `git tag | tail -n 1`
echo Hit return to build and push release: $TAG
read

make
git tag $TAG
git push --tags
gothub -v release -u 0xfe -r lumen --tag $TAG --name "release: $TAG"

for r in lumen.linux.amd64 lumen.linux.arm lumen.linux.arm64 lumen.macos lumen.windows; do
    gothub -v upload -u 0xfe -r lumen --tag $TAG --name $r --file dist/$r
done

