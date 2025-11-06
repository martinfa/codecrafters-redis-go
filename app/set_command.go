package main

func HandleSet(cmd *RedisCommand) string {
	cache := GetInstance()
	cache.Set(cmd.Args[0], cmd.Args[1])
	return "+OK\r\n"
}
