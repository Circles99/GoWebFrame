local val = redis.call('get', KEYS[1])

local expiration = ARGV[1]
-- 获取有多少个可过
local limit = tonumber(ARGV[2])

-- 当没有这个key的时候
if val == false then
    -- limit 小于1 代表全部都不能过
    if limit < 1 then
        -- 执行限流
        return "true"
    else
        -- key 不存在 设置初始值 1，并设置过期时间
        redis.call('set', KEYS[1], 1, 'px', expiration)
        -- 不执行限流
        return "false"
    end
elseif tonumber(val) < limit then
    -- 自增1
    redis.call('incr', KEYS[1])
    -- 不需要限流
    return "false"
else
    return "true"
end