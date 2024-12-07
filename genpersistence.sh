#!/bin/bash

DATABASE_LOCATION=$HOME/.go4ignition/go4ignition.db

TMP_DIR=$(mktemp -d)

# Define a cleanup function
cleanup() {
    rm ${TMP_DIR} -rf
}

# Set the trap to call cleanup on exit
trap cleanup EXIT

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