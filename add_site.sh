#!/bin/bash

DOMAIN_NAME=$1
SKELETON_DIR="sites/.skeleton"
OWD=$(pwd)

# Always clean up after yourself
cleanup() {
    cd "${OWD}"
}

trap cleanup EXIT

if [ ! -d "${SKELETON_DIR}" ]; then
  echo "Skeleton site does not exist: ${SKELETON_DIR}"
  exit 1
fi

if [[ ! "${DOMAIN_NAME}" =~ ^[.a-zA-Z0-9-]+$ ]]; then
  echo "Not a valid domain name: ${DOMAIN_NAME}"
  exit 2
fi

# Strip out periods
SITE_NAME=$(echo "${DOMAIN_NAME}" | sed 's/\./_/g' )
SITE_DIR="sites/${SITE_NAME}"

if [ -d "${SITE_DIR}" ];then
  echo "Site ${SITE_DIR} already exists"
  exit 3
fi

# Create a directory for the site
mkdir -p "${SITE_DIR}"

# Generate a skeleton site
cp -r "${SKELETON_DIR}"/* "${SITE_DIR}"

# Remove the skeleton readme file from the new site
rm -f "${SITE_DIR}/README.md"

# Update package names in the copied go source files
cd "${SITE_DIR}"
find . -type f -name "*.go" -exec sed -i "s/package _skeleton/package ${SITE_NAME}/g" {} +

cd "${OWD}"

# Regenerate site imports to ensure that the new site's package init method can be called
./genimport.sh
