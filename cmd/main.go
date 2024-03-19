package main

import (
	"database/sql"
	"flag"
	"gomigrate"
	"log"
)

func updateHistory(conn *sql.DB, entries map[string]gomigrate.HashEntry, file gomigrate.MigrationFile) error {
	stmt, err := conn.Prepare("INSERT INTO public.migration_history (name) VALUES ($1)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(file.Name)
	if err != nil {
		return err
	}
	entries[file.Name] = gomigrate.HashEntry{Name: file.Name}
	return nil
}

func main() {
	fNamePtr := flag.String("config-file", "", "Specify a non-default config file")
	flag.Parse()
	configs := gomigrate.Migration{}
	if *fNamePtr != "" {
		if err := configs.Init(*fNamePtr); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := configs.Init(); err != nil {
			log.Fatal(err)
		}
	}
	for _, config := range configs.DBParams {
		migrationFiles, err := gomigrate.ReadMigrationFiles(config.Queries)
		if err != nil {
			log.Fatal(err)
		}
		conn, err := gomigrate.ConnectDB(config.DBType, config)
		if err != nil {
			log.Fatal(err)
		}
		entries, err := gomigrate.FetchHashes(conn)
		if err != nil {
			log.Fatal(err)
		}
		for _, mFile := range migrationFiles {
			if _, ok := entries[mFile.Name]; ok {
				log.Println("Already processed \"" + mFile.Name + "\"")
				continue
			}
			for _, f := range mFile.Queries {
				stmt, err := conn.Prepare(f)
				if err != nil {
					if mFile.IsProc {
						log.Println("Error creating stored procedure:", mFile.Name)
					} else {
						log.Println("Error creating the following query:", f)
					}
					log.Println("Error in file \"" + mFile.Name + "\"")
					log.Fatal(err)
				}
				_, err = stmt.Exec()
				if err != nil {
					if mFile.IsProc {
						log.Println("Error creating stored procedure:", mFile.Name)
					} else {
						log.Println("Error executing the following query:", f)
					}
					log.Fatal(err)
				}
			}
			if err := updateHistory(conn, entries, mFile); err != nil {
				log.Println("Error updating history table on file \"" + mFile.Name + "\"")
				log.Fatal(err)
			}
		}
		conn.Close()
	}
}
