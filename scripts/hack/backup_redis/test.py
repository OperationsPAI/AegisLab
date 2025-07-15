import redis

# 连接 Redis
r = redis.Redis(host="127.0.0.1", port=6379, db=0, decode_responses=True)

# 方法1：直接插入
hash_key = "injection:algorithms"
field = "9d1569bb-cd37-4acf-91e3-78e4a067ec83"
value = '[{"name":"traceback","image":"","tag":""}]'

r.hset(hash_key, field, value)
