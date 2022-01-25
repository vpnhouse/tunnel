#!/bin/sh

containers="tunnel"

tags=$@
if [ ${#tags} -eq 0 ]
then
  tags=`git rev-parse --abbrev-ref HEAD`
fi

for tag in $tags
do
  for container in $containers
  do
    docker push codenameuranium/$container:$tag
  done
done

