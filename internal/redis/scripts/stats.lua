local likesKey = KEYS[1]
local dislikesKey = KEYS[2]
local visitsKey = KEYS[3]

local likes = redis.call("GET", likesKey)
local dislikes = redis.call("GET", dislikesKey)
local visits = redis.call("GET", visitsKey)

return {
    likes or "0",
    dislikes or "0",
    visits or "0"
}