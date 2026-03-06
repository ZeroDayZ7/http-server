local cooldownKey = KEYS[1]
local statsKey = KEYS[2]
local ttl = ARGV[1]

if redis.call("EXISTS", cooldownKey) == 0 then
    redis.call("SET", cooldownKey, "1", "PX", ttl)
    
    if redis.call("EXISTS", statsKey) == 1 then
        redis.call("INCR", statsKey)
    end
    
    return 1
end

return 0