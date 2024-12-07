#!/bin/bash -e

TMP_DIR=$(mktemp -d)

# Define a cleanup function
cleanup() {
    rm ${TMP_DIR} -rf
}

# Set the trap to call cleanup on exit
trap cleanup EXIT

# Abort if we're in the wrong folder or if the expected folder does not exist
if [ ! -d 'static' ]; then
  echo "static directory does not exist"
  exit 1
fi

STATIC_GO="${TMP_DIR}/static.go"
rm -f ${STATIC_GO}

STATIC_MAP_GO="${TMP_DIR}/static-map.go"
rm -f ${STATIC_MAP_GO}

STATIC_CONTENT_TYPE_GO="${TMP_DIR}/static-content-type.go"
rm -f ${STATIC_CONTENT_TYPE_GO}

STATIC_NAME_MAP_GO="${TMP_DIR}/static-name-map.go"
rm -f ${STATIC_NAME_MAP_GO}

cat <<EOF >> ${STATIC_GO}
// Generated by genstatic.sh
package main

import (
	_ "embed"
)

EOF

FILES_TXT=${TMP_DIR}/files.txt
find static/ -type f > ${FILES_TXT}

echo "// StaticFiles generated bindings for static files" >> ${STATIC_MAP_GO}
echo "var StaticFiles = map[string][]byte{" >> ${STATIC_MAP_GO}
echo -e "\n// StaticFilesContentType mime types for static files" >> ${STATIC_CONTENT_TYPE_GO}
echo "var StaticFilesContentType = map[string]string{" >> ${STATIC_CONTENT_TYPE_GO}
echo -e "\n// StaticFileNames Resolve actual file names to md5 file names" >> ${STATIC_NAME_MAP_GO}
echo "var StaticFileNames = map[string]string{" >> ${STATIC_NAME_MAP_GO}

for FILE in `cat ${FILES_TXT}`; do
  FILE_EXTENSION="${FILE##*.}"
  DIRNAME=$(dirname ${FILE})
  BASENAME=$(basename ${FILE})
  MIME_TYPE=""

  # Skip editor swap files
  if [[ "${FILE: -1}" == "~" ]]; then
    continue
  fi

  # Skip markdown files. There's no need to embed them in the binary
  if [ "md" == "${FILE_EXTENSION}" ];then
    continue
  fi

  if [ "js" == "${FILE_EXTENSION}" ]; then
    MIME_TYPE="text/javascript"
  elif [ "css" == "${FILE_EXTENSION}" ]; then
    MIME_TYPE="text/css"
  elif [ "svg" == "${FILE_EXTENSION}" ]; then
    MIME_TYPE="image/svg+xml"
  else
    MIME_TYPE=$(file -b --mime-type $FILE)
  fi

  MD5=$(md5sum ${FILE} | awk '{print $1}')

  if [ "favicon.ico" == "${BASENAME}" ];then
    VARNAME="FaviconICO"
  else
    VARNAME="md5${MD5}"
  fi

  echo "  \"/${DIRNAME}/${MD5}.${FILE_EXTENSION}\": \"${MIME_TYPE}\", // ${FILE}" >> ${STATIC_CONTENT_TYPE_GO}
  echo "  \"/${FILE}\": \"/${DIRNAME}/${MD5}.${FILE_EXTENSION}\"," >> ${STATIC_NAME_MAP_GO}

  # Bind the files to go variables
  echo "//go:embed ${FILE}" >> ${STATIC_GO}
  echo "var ${VARNAME} []byte" >> ${STATIC_GO}
  echo >> ${STATIC_GO}

  # Expose the values of the go variables through a map
  echo "  \"/${DIRNAME}/${MD5}.${FILE_EXTENSION}\": ${VARNAME}, // ${FILE}" >> ${STATIC_MAP_GO}
done

echo "}" >> ${STATIC_MAP_GO}
echo "}" >> ${STATIC_CONTENT_TYPE_GO}
echo "}" >> ${STATIC_NAME_MAP_GO}


# Append the map to the final output
cat ${STATIC_MAP_GO} >> ${STATIC_GO}
cat ${STATIC_CONTENT_TYPE_GO} >> ${STATIC_GO}
cat ${STATIC_NAME_MAP_GO} >> ${STATIC_GO}
cp ${STATIC_GO} .