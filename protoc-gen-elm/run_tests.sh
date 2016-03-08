#!/bin/sh

cd "$(dirname "$0")"
set -e

elm-package install -y
elm-make --yes --output raw-test.js Test.elm
bash ./elm-stuff/packages/laszlopandy/elm-console/1.1.0/elm-io.sh raw-test.js test.js
node test.js
