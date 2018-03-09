#!/bin/bash

TAG=$1

if [ "$TAG" == "" ]; then
    echo Usage: $0 tag
    echo Current tag: `git tag`
fi

git status
echo Make sure changes are committed, and new releases built. Hit return to continue.
read

git tag $TAG
git push --tags
gothub release -u 0xfe -r lumen --tag $TAG --name "release: $TAG" -v

for r in lumen.linux.amd64 lumen.linux.arm lumen.linux.arm64 lumen.macos lumen.windows
    gothub upload -u 0xfe -r lumen --tag $TAG --name $r --file dist/$r -v
done

