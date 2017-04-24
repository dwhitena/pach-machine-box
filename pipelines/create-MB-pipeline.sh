#!/bin/bash

pipelinefile=$1

sed 's/myMBKey/'"$MB_KEY"'/g' $pipelinefile > deploy.json
pachctl create-pipeline -f deploy.json
rm deploy.json
