local cooldownKey = KEYS[1]
local statsKey = KEYS[2]
local ttl = ARGV[1]

if redis.call("EXISTS", cooldownKey) == 0 then
    redis.call("SET", cooldownKey, "1", "PX", ttl)
    redis.call("INCR", statsKey)
    return 1
end

return 0