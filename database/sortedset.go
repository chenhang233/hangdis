package database

import (
	SortedSet "hangdis/datastruct/sortedset"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"strconv"
	"strings"
)

func (db *DB) getAsSortedSet(key string) (SortedSet.SortedSet, redis.ErrorReply) {
	entity, exists := db.GetEntity(key)
	if !exists {
		return nil, nil
	}
	sortedSet, ok := entity.Data.(SortedSet.SortedSet)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return sortedSet, nil
}

func (db *DB) getOrInitSortedSet(key string) (SortedSet.SortedSet, bool, redis.ErrorReply) {
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return nil, false, err
	}
	isNew := false
	if set == nil {
		set = SortedSet.SimpleMake()
		db.PutEntity(key, &database.DataEntity{Data: set})
		isNew = true
	}
	return set, isNew, nil
}

func undoZAdd(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	size := (len(args) - 1) / 2
	fields := make([]string, size)
	for i := 0; i < size; i++ {
		fields[i] = string(args[2*i+2])
	}
	return rollbackZSetFields(db, key, fields...)
}

func undoZIncr(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	field := string(args[2])
	return rollbackZSetFields(db, key, field)
}

func undoZRem(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	fields := make([]string, len(args)-1)
	fieldArgs := args[1:]
	for i, v := range fieldArgs {
		fields[i] = string(v)
	}
	return rollbackZSetFields(db, key, fields...)
}

func execZAdd(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	tp := string(args[1])
	policy := upsertPolicy
	if tp == "NX" {
		policy = insertPolicy
	} else if tp == "XX" {
		policy = updatePolicy
	}
	var n, j int
	if policy != upsertPolicy {
		n = (len(args) - 2) / 2
		j = 1
	} else {
		n = (len(args) - 1) / 2
		j = 0
	}
	elements := make([]*SortedSet.Element, n)
	for i := 0; i < n; i++ {
		scoreValue := string(args[i*2+1+j])
		score, err := strconv.ParseFloat(scoreValue, 64)
		if err != nil {
			return protocol.MakeErrReply("ERR value is not a valid float")
		}
		member := string(args[i*2+2+j])
		elements[i] = &SortedSet.Element{
			Member: member,
			Score:  score,
		}
	}
	set, _, err := db.getOrInitSortedSet(key)
	if err != nil {
		return err
	}
	i := 0
	for _, element := range elements {
		if policy != upsertPolicy {
			_, ok := set.Get(element.Member)
			if policy == insertPolicy && ok {
				continue
			}
			if policy == updatePolicy && !ok {
				continue
			}
		}
		if set.Add(element.Member, element.Score) {
			i++
		}
	}
	db.addAof(utils.ToCmdLine3("zadd", args...))
	return protocol.MakeIntReply(int64(i))
}

func execZScore(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	member := string(args[1])
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	element, ok := set.Get(member)
	if !ok {
		return protocol.MakeEmptyMultiBulkReply()
	}
	value := strconv.FormatFloat(element.Score, 'f', -1, 64)
	return protocol.MakeBulkReply([]byte(value))
}

func execZIncrBy(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	raw := string(args[1])
	member := string(args[2])
	increment, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not a valid float")
	}
	sortedSet, _, err2 := db.getOrInitSortedSet(key)
	if err2 != nil {
		return err2
	}
	element, exists := sortedSet.Get(member)
	if !exists {
		sortedSet.Add(member, increment)
		db.addAof(utils.ToCmdLine3("zincrby", args...))
		return protocol.MakeBulkReply(args[1])
	}
	score := element.Score + increment
	sortedSet.Add(member, score)
	bytes := []byte(strconv.FormatFloat(score, 'f', -1, 64))
	db.addAof(utils.ToCmdLine3("zincrby", args...))
	return protocol.MakeBulkReply(bytes)
}

func execZRank(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	member := string(args[1])
	sortedSet, errReply := db.getAsSortedSet(key)
	if errReply != nil {
		return errReply
	}
	if sortedSet == nil {
		return &protocol.EmptyMultiBulkReply{}
	}

	rank := sortedSet.GetRank(member, false)
	if rank < 0 {
		return &protocol.EmptyMultiBulkReply{}
	}
	return protocol.MakeIntReply(rank)
}

func execZCount(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	min, err := SortedSet.ParseScoreBorder(string(args[1]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	max, err := SortedSet.ParseScoreBorder(string(args[2]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	sortedSet, errReply := db.getAsSortedSet(key)
	if errReply != nil {
		return errReply
	}
	if sortedSet == nil {
		return protocol.MakeIntReply(0)
	}
	return protocol.MakeIntReply(sortedSet.Count(min, max))
}

func execZRevRank(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	member := string(args[1])
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	rank := set.GetRank(member, true)
	if rank < 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	return protocol.MakeIntReply(rank)
}

func execZCard(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	return protocol.MakeIntReply(set.Len())

}

func zRangeParse0(db *DB, key string, start64 int64, stop64 int64, withScores bool, desc bool) redis.Reply {
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	start, end := utils.ConvertRange2(start64, stop64, set.Len())
	if start < 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	elements := set.Range(start, end, desc)
	if withScores {
		slice := make([][]byte, len(elements)*2)
		i := 0
		for _, e := range elements {
			slice[i] = []byte(e.Member)
			i++
			scoreStr := strconv.FormatFloat(e.Score, 'f', -1, 64)
			slice[i] = []byte(scoreStr)
			i++
		}
		return protocol.MakeMultiBulkReply(slice)
	}
	slice := make([][]byte, len(elements))
	for i, e := range elements {
		slice[i] = []byte(e.Member)
	}
	return protocol.MakeMultiBulkReply(slice)
}

func zRangeParse1(db *DB, key string, min *SortedSet.ScoreBorder, max *SortedSet.ScoreBorder, offset int64, limit int64, withScores bool, desc bool) redis.Reply {
	set, err2 := db.getAsSortedSet(key)
	if err2 != nil {
		return err2
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	elements := set.RangeByScore(min, max, offset, limit, desc)
	if withScores {
		slice := make([][]byte, len(elements)*2)
		i := 0
		for _, e := range elements {
			slice[i] = []byte(e.Member)
			i++
			scoreStr := strconv.FormatFloat(e.Score, 'f', -1, 64)
			slice[i] = []byte(scoreStr)
			i++
		}
		return protocol.MakeMultiBulkReply(slice)
	}
	slice := make([][]byte, len(elements))
	for i, e := range elements {
		slice[i] = []byte(e.Member)
	}
	return protocol.MakeMultiBulkReply(slice)
}

func execZRange(db *DB, args [][]byte) redis.Reply {
	withScores := false
	if len(args) == 4 {
		if strings.ToLower(string(args[3])) != "withscores" {
			return protocol.MakeSyntaxErrReply()
		}
		withScores = true
	}
	key := string(args[0])
	start64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	stop64, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	return zRangeParse0(db, key, start64, stop64, withScores, false)
}

func execZRangeByScore(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	min, err := SortedSet.ParseScoreBorder(string(args[1]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	max, err := SortedSet.ParseScoreBorder(string(args[2]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	withScores := false
	var offset int64 = 0
	var limit int64 = 0
	if len(args) > 3 {
		for i := 3; i < len(args); {
			s := string(args[i])
			if strings.ToUpper(s) == "WITHSCORES" {
				withScores = true
				i++
			} else if strings.ToUpper(s) == "LIMIT" {
				if len(args) < i+3 {
					return protocol.MakeErrReply("ERR syntax error")
				}
				offset, err = strconv.ParseInt(string(args[i+1]), 10, 64)
				if err != nil {
					return protocol.MakeErrReply("ERR value is not an integer or out of range")
				}
				limit, err = strconv.ParseInt(string(args[i+2]), 10, 64)
				if err != nil {
					return protocol.MakeErrReply("ERR value is not an integer or out of range")
				}
				i += 3
			} else {
				return protocol.MakeErrReply("ERR syntax error")
			}
		}
	}
	return zRangeParse1(db, key, min, max, offset, limit, withScores, false)
}

func execZRevRange(db *DB, args [][]byte) redis.Reply {
	withScores := false
	if len(args) == 4 {
		if string(args[3]) != "WITHSCORES" {
			return protocol.MakeErrReply("syntax error")
		}
		withScores = true
	}
	key := string(args[0])
	start, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	stop, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	return zRangeParse0(db, key, start, stop, withScores, true)
}

func execZRevRangeByScore(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	min, err := SortedSet.ParseScoreBorder(string(args[2]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	max, err := SortedSet.ParseScoreBorder(string(args[1]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	withScores := false
	var offset int64 = 0
	var limit int64 = -1
	if len(args) > 3 {
		for i := 3; i < len(args); {
			s := string(args[i])
			if strings.ToUpper(s) == "WITHSCORES" {
				withScores = true
				i++
			} else if strings.ToUpper(s) == "LIMIT" {
				if len(args) < i+3 {
					return protocol.MakeErrReply("ERR syntax error")
				}
				offset, err = strconv.ParseInt(string(args[i+1]), 10, 64)
				if err != nil {
					return protocol.MakeErrReply("ERR value is not an integer or out of range")
				}
				limit, err = strconv.ParseInt(string(args[i+2]), 10, 64)
				if err != nil {
					return protocol.MakeErrReply("ERR value is not an integer or out of range")
				}
				i += 3
			} else {
				return protocol.MakeErrReply("ERR syntax error")
			}
		}
	}
	return zRangeParse1(db, key, min, max, offset, limit, withScores, true)
}

func execZPopMin(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	count := 1
	if len(args) > 1 {
		var err error
		count, err = strconv.Atoi(string(args[1]))
		if err != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
	}

	sortedSet, errReply := db.getAsSortedSet(key)
	if errReply != nil {
		return errReply
	}
	if sortedSet == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}

	removed := sortedSet.PopMin(count)
	if len(removed) > 0 {
		db.addAof(utils.ToCmdLine3("zpopmin", args...))
	}
	result := make([][]byte, 0, len(removed)*2)
	for _, element := range removed {
		scoreStr := strconv.FormatFloat(element.Score, 'f', -1, 64)
		result = append(result, []byte(element.Member), []byte(scoreStr))
	}
	return protocol.MakeMultiBulkReply(result)
}

func execZRem(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	fields := make([]string, len(args)-1)
	fieldArgs := args[1:]
	for i, v := range fieldArgs {
		fields[i] = string(v)
	}
	sortedSet, errReply := db.getAsSortedSet(key)
	if errReply != nil {
		return errReply
	}
	if sortedSet == nil {
		return protocol.MakeIntReply(0)
	}
	var deleted int64 = 0
	for _, field := range fields {
		if sortedSet.Remove(field) {
			deleted++
		}
	}
	if deleted > 0 {
		db.addAof(utils.ToCmdLine3("zrem", args...))
	}
	return protocol.MakeIntReply(deleted)
}

func execZRemRangeByScore(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	min, err := SortedSet.ParseScoreBorder(string(args[1]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	max, err := SortedSet.ParseScoreBorder(string(args[2]))
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	sortedSet, errReply := db.getAsSortedSet(key)
	if errReply != nil {
		return errReply
	}
	if sortedSet == nil {
		return &protocol.EmptyMultiBulkReply{}
	}
	removed := sortedSet.RemoveByScore(min, max)
	if removed > 0 {
		db.addAof(utils.ToCmdLine3("zremrangebyscore", args...))
	}
	return protocol.MakeIntReply(removed)
}

func execZRemRangeByRank(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	start64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	stop64, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	// get data
	sortedSet, errReply := db.getAsSortedSet(key)
	if errReply != nil {
		return errReply
	}
	if sortedSet == nil {
		return protocol.MakeIntReply(0)
	}
	size := sortedSet.Len()
	start, stop := utils.ConvertRange(start64, stop64, size)
	removed := sortedSet.RemoveByRank(int64(start), int64(stop))
	if removed > 0 {
		db.addAof(utils.ToCmdLine3("zremrangebyrank", args...))
	}
	return protocol.MakeIntReply(removed)
}

func init() {
	RegisterCommand("ZADD", execZAdd, writeFirstKey, undoZAdd, -4, flagWrite)
	RegisterCommand("ZSCORE", execZScore, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("ZINCRBY", execZIncrBy, writeFirstKey, undoZIncr, 4, flagWrite)
	RegisterCommand("ZRANK", execZRank, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("ZCOUNT", execZCount, readFirstKey, nil, 4, flagReadOnly)
	RegisterCommand("ZREVRANK", execZRevRank, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("ZCARD", execZCard, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("ZRANGE", execZRange, readFirstKey, nil, -4, flagReadOnly)
	RegisterCommand("ZRANGEBYSCORE", execZRangeByScore, readFirstKey, nil, -4, flagReadOnly)
	RegisterCommand("ZREVRANGE", execZRevRange, readFirstKey, nil, -4, flagReadOnly)
	RegisterCommand("ZREVRANGEBYSCORE", execZRevRangeByScore, readFirstKey, nil, -4, flagReadOnly)
	RegisterCommand("ZPOPMIN", execZPopMin, writeFirstKey, rollbackFirstKey, -2, flagWrite)
	RegisterCommand("ZREM", execZRem, writeFirstKey, undoZRem, -3, flagWrite)
	RegisterCommand("ZREMRANGEBYSCORE", execZRemRangeByScore, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("ZREMRANGEBYRANK", execZRemRangeByRank, writeFirstKey, rollbackFirstKey, 4, flagWrite)
}
