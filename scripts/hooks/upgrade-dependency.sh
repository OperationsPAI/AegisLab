#!/bin/bash
git submodule foreach 'git checkout main && git pull origin main'
git add .
git commit -m "Update all submodules"