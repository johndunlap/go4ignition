#!/bin/bash

# TODO: Use a transient database for code generation
DATABASE_LOCATION=$HOME/.go4ignition/go4ignition.db
TMP_DIR=$(mktemp -d)
CWD=$(pwd)

# Always clean up after yourself
cleanup() {
    rm ${TMP_DIR} -rf
}

trap cleanup EXIT

SITES=$(ls sites)

for SITE in $SITES; do
  if [ ".skeleton" == "${SITE}" ]; then
    continue
  fi

  if [ ! -d "sites/${SITE}" ]; then
    continue
  fi

  FQDN=$(echo $SITE | sed 's/_/./g')
  cd "sites/${SITE}"

  echo "Generating persistence code for $FQDN"

  PERSISTENCE_GO="${TMP_DIR}/persistence.go"
  rm -f ${PERSISTENCE_GO}

  TABLES_TXT=$TMP_DIR/tables.txt
  rm -f ${TABLES_TXT}

  sqlite3 ${DATABASE_LOCATION} <<EOF > ${TABLES_TXT}
      select
          name
      from sqlite_master
      where name not like '%sqlite%'
  EOF

  for TABLE_NAME in $(cat ${TABLES_TXT}); do
    echo "TABLE: ${TABLE_NAME}"
  done

  cd "${CWD}"
done
