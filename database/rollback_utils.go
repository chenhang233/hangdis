package database

import "hangdis/utils"

func rollbackFirstKey(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	return rollbackGivenKeys(db, key)
}

func rollbackGivenKeys(db *DB, keys ...string) []CmdLine {
	var undoCmdLines [][][]byte
	for _, key := range keys {
		//entity, ok := db.GetEntity(key)
		undoCmdLines = append(undoCmdLines, utils.ToCmdLine("DEL", key))
	}
	return undoCmdLines
}

func writeFirstKey(args [][]byte) ([]string, []string) {
	key := string(args[0])
	return []string{key}, nil
}
