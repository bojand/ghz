#!/bin/sh

# build frontend
cd web/ui
npm install && npm run build

# back up
cd ../../

# build statik bundle
statik -src=web/ui/dist -dest=web/router
