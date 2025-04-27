//go:build monolith
// +build monolith

package main

func SetupTables() {
	SetupWorldgenTables()
	SetupGameEngineTables()
}
