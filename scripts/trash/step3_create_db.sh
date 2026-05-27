rm ../../data/modules/pathway/pathway-v3.db
cat tables.sql | sqlite3 ../../data/modules/pathway/pathway-v3.db
cat ../../data/modules/pathway/pathway.sql | sqlite3 ../../data/modules/pathway/pathway-v3.db