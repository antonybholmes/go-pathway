PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE pathway (
	id INTEGER PRIMARY KEY ASC,
	dataset TEXT NOT NULL,
	name TEXT NOT NULL,
	source TEXT NOT NULL,
	genes TEXT NOT NULL);
CREATE INDEX pathway_dataset_name_idx ON pathway (dataset, name);
CREATE INDEX pathway_name_idx ON pathway (name); 

CREATE TABLE genes (
	id INTEGER PRIMARY KEY ASC,
	gene_symbol TEXT NOT NULL);
CREATE INDEX genes_gene_symbol_idx ON genes (gene_symbol);

 