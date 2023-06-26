package dict

type InstanceDict struct {
	m map[string]any
}

func MakeInstanceDict() *InstanceDict {
	return &InstanceDict{
		m: make(map[string]any),
	}
}

func (dict *InstanceDict) Get(key string) (val any, exists bool) {
	val, ok := dict.m[key]
	return val, ok
}

func (dict *InstanceDict) Len() int {
	if dict.m == nil {
		panic("m is nil")
	}
	return len(dict.m)
}

func (dict *InstanceDict) Put(key string, val any) (result int) {
	_, existed := dict.m[key]
	dict.m[key] = val
	if existed {
		return 0
	}
	return 1
}

func (dict *InstanceDict) PutIfAbsent(key string, val any) (result int) {
	_, existed := dict.m[key]
	if existed {
		return 0
	}
	dict.m[key] = val
	return 1
}

func (dict *InstanceDict) PutIfExists(key string, val any) (result int) {
	_, existed := dict.m[key]
	if existed {
		dict.m[key] = val
		return 1
	}
	return 0
}

func (dict *InstanceDict) Remove(key string) (result int) {
	_, existed := dict.m[key]
	delete(dict.m, key)
	if existed {
		return 1
	}
	return 0
}

func (dict *InstanceDict) Keys() []string {
	result := make([]string, len(dict.m))
	i := 0
	for k := range dict.m {
		result[i] = k
		i++
	}
	return result
}

func (dict *InstanceDict) ForEach(consumer Consumer) {
	for k, v := range dict.m {
		if !consumer(k, v) {
			break
		}
	}
}

func (dict *InstanceDict) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		for k := range dict.m {
			result[i] = k
			break
		}
	}
	return result
}

func (dict *InstanceDict) RandomDistinctKeys(limit int) []string {
	size := limit
	if size > len(dict.m) {
		size = len(dict.m)
	}
	result := make([]string, size)
	i := 0
	for k := range dict.m {
		if i == size {
			break
		}
		result[i] = k
		i++
	}
	return result
}

func (dict *InstanceDict) Clear() {
	*dict = *MakeInstanceDict()
}
