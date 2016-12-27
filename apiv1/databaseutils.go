package main

import "flag"

func GetDatabaseUtilFlags() *bool {

	sqlMapVerify := flag.Bool("sqlmapverify", false, "scan all maps in database and attempt to verify them all")
	return sqlMapVerify
}
