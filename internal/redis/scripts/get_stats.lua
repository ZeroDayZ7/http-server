-- keys: statsKeyLike, statsKeyDislike, statsKeyVisit
-- argv: (opcjonalnie) TTL do ustawienia przy fallbacku
local likes = redis.call("GET", KEYS[1])
local dislikes = redis.call("GET", KEYS[2])
local visits = redis.call("GET", KEYS[3])

if not likes or not dislikes or not visits then
    return nil
end

return {likes, dislikes, visits}