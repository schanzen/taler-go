// This file is part of taler-go, the Taler Go implementation.
// Copyright (C) 2026 Martin Schanzenbach
//
// Taler Go is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Taler Go is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later

package util

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckVersioning(db *sql.DB) (bool, error) {
	rows, err := db.Query(`SELECT schema_name FROM information_schema.schemata WHERE schema_name='_v';`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		fmt.Println("Versioning applied")
		return true, nil
	}
	return false, nil
}

func CheckPatch(db *sql.DB, patchName string) (bool, error) {
	rows, err := db.Query(`SELECT applied_by FROM _v.patches WHERE patch_name=$1 LIMIT 1;`, patchName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func RunSQL(db *sql.DB, patchName string, dbName string) error {
	path, err := exec.LookPath("psql")
	if err != nil {
		return err
	}
	_, err = exec.Command(path, dbName, "-f", patchName, "-q", "--set", "ON_ERROR_STOP=1").Output()
	fmt.Printf("Running: %s %s %s %s %s %s %s\n", path, dbName, "-f", patchName, "-q", "--set", "ON_ERROR_STOP=1")
	if err != nil {
		return err
	}
	return nil

}

func DBInit(db *sql.DB, datahome string, dbName string, patchesPrefix string) error {
	applied, err := CheckVersioning(db)
	loadSuffix := filepath.Join(datahome, "sql")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	if !applied {
		err := RunSQL(db, filepath.Join(loadSuffix, "versioning.sql"), dbName)
		if err != nil {
			return err
		}
	}
	for i := range 10000 {
		patchName := fmt.Sprintf("%s-%04d", patchesPrefix, i+1)
		applied, err := CheckPatch(db, patchName)
		if err != nil {
			return err
		}
		if applied {
			fmt.Printf("Patch %s already applied\n", patchName)
			continue
		}
		patchFile := fmt.Sprintf("%s.sql", filepath.Join(loadSuffix, patchName))
		if _, err := os.Stat(patchFile); err != nil {
			fmt.Printf("Patch %s not found, up-to-date.\n", patchFile)
			break
		}
		fmt.Printf("Applying patch %s\n", patchName)
		err = RunSQL(db, patchFile, dbName)
		if err != nil {
			return err
		}
	}
	return nil
}
