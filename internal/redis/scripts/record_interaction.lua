-- keys: cooldownKey, statsKey
-- argv: cooldownTTL (ms)
local cooldownKey = KEYS[1]
local statsKey = KEYS[2]
local ttl = tonumber(ARGV[1])

-- atomowe SETNX + PEXPIRE + INCR
if redis.call("SETNX", cooldownKey, "1") == 1 then
    redis.call("PEXPIRE", cooldownKey, ttl)
    redis.call("INCR", statsKey)
    return 1
end

return 0