package gomigrate

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// MigrationFile contains all queries to be executed from file
type MigrationFile struct {
	Name    string
	Full    string
	Major   int
	Minor   int
	Fix     int
	Action  rune
	IsProc  bool
	Hash    string
	Queries []string
}

// HashEntry contains file hashes
type HashEntry struct {
	Name string
}

// ReadMigrationFiles reads all migration files
func ReadMigrationFiles(filesPath string) ([]MigrationFile, error) {
	files, err := ioutil.ReadDir(filesPath)
	if err != nil {
		return nil, err
	}
	r, err := regexp.Compile("(--.*)")
	if err != nil {
		return nil, err
	}
	migrationFiles := make([]MigrationFile, len(files))
	for index, file := range files {
		mFile := &migrationFiles[index]
		name := file.Name()
		mFile.Name = name
		mFile.Full = filepath.Join(filesPath, name)
		split := strings.Split(name, "_")
		if len(split) < 2 {
			return nil, errors.New("Filename \"" + name + "\" is not formatted properly")
		}
		version := split[0]
		if strings.HasPrefix(version, "p") {
			mFile.Major = 0
			mFile.Minor = 0
			mFile.Fix = 0
			mFile.IsProc = true
			mFile.Action = 'p'
			mFile.Queries = make([]string, 1)
			data, err := ioutil.ReadFile(mFile.Full)
			if err != nil {
				return nil, errors.New("Error reading file \"" + name + "\"")
			}
			mFile.Queries[0] = string(data)
		} else if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "u") {
			split = strings.Split(version, ".")
			if len(split) != 3 && len(split) != 4 {
				return nil, errors.New(
					"Version requires at least a major and minor version for file \"" + name + "\"",
				)
			}
			mFile.Major, err = strconv.Atoi(split[1])
			if err != nil {
				return nil, errors.New("Error converting major version in \"" + name + "\"")
			}
			mFile.Minor, err = strconv.Atoi(split[2])
			if err != nil {
				return nil, errors.New("Error converting minor version in \"" + name + "\"")
			}
			if len(split) == 4 {
				mFile.Fix, err = strconv.Atoi(split[3])
				if err != nil {
					return nil, errors.New("Error converting major version in \"" + name + "\"")
				}
			} else {
				mFile.Fix = 0
			}
			mFile.IsProc = false
			if strings.HasPrefix(version, "v") {
				mFile.Action = 'v'
			} else {
				mFile.Action = 'u'
			}
			data, err := ioutil.ReadFile(mFile.Full)
			if err != nil {
				return nil, errors.New("Error reading file \"" + name + "\"")
			}
			data = r.ReplaceAll(data, make([]byte, 0))
			queries := strings.Split(string(data), ";")
			mFile.Queries = make([]string, 0, len(queries))
			for _, q := range queries {
				qs := strings.Split(q, "\n")
				for idx := range qs {
					qs[idx] = strings.Trim(qs[idx], " \t\r\n")
				}
				q = strings.Join(qs, "")
				if q == "" {
					continue
				}
				mFile.Queries = append(mFile.Queries, q)
			}

		} else {
			return nil, errors.New("Unknown prefix on file \"" + name + "\"")
		}
	}

	sort.SliceStable(
		migrationFiles,
		func(i int, j int) bool {
			m1 := migrationFiles[i]
			m2 := migrationFiles[j]
			if m1.Major == m2.Major {
				if m1.Minor == m2.Minor {
					if m1.Fix == m2.Fix {
						if m1.Action == m2.Action {
							return false
						}
						return m1.Action < m2.Action
					}
					return m1.Fix < m2.Fix
				}
				return m1.Minor < m2.Minor
			}
			return m1.Major < m2.Major
		},
	)
	return migrationFiles, nil
}

// FetchHashes fetches all the file hashes from the DB
func FetchHashes(conn *sql.DB) (map[string]HashEntry, error) {
	var entries map[string]HashEntry = make(map[string]HashEntry)
	stmt, err := conn.Prepare("SELECT name FROM public.migration_history")
	if err != nil {
		log.Println("Initializing creation history table")
		stmt, err := conn.Prepare("CREATE TABLE public.migration_history (name VARCHAR(100))")
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec()
		if err != nil {
			return nil, err
		}
		return entries, nil
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		entry := HashEntry{}
		rows.Scan(&entry.Name)
		entries[entry.Name] = entry
	}
	return entries, nil
}
