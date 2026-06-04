package main

func HandleType(command *RedisCommand) string {
	if len(command.Args) != 1 {
		return "-ERR wrong number of arguments for 'type' command\r\n"
	}

	cache := GetInstance()
	value := cache.Get(command.Args[0])
	if value == nil {
		return "+none\r\n"
	}

	if _, isStream := value.(*Stream); isStream {
		return "+stream\r\n"
	}

	return "+string\r\n"
}
