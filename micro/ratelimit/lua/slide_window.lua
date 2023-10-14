
-- 1 2 3 4 5 6 7 这些元素
-- ZREMRANGEBYSCORE key 0 6
-- 执行完成之后  还剩7


-- 限流对象
local key = KEYS[1]


-- 窗口的大小
local window = tonumber(ARGV[1])
-- 阈值
local threshold = tonumber(ARGV[2])

local now = tonumber(ARGV[3])

-- 窗口的起始时间
local min = now - window


-- 先挪动窗口
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)

-- 两个是等价的
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
--local cnt = redis.call('ZCOUNT', key, min, '+inf')

if cnt >= threshold then
    -- 执行限流
    return "true"
else
    -- 把 score 和 member 都设置成now
    redis.call("ZADD", key, now, now)
    redis.call("PEXPORE", key, window)
    return "false"
end