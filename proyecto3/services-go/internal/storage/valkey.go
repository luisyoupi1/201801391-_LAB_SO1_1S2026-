package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
	"github.com/redis/go-redis/v9"
)

type Store struct{ Client *redis.Client }

func New(addr string) *Store {
	return &Store{Client: redis.NewClient(&redis.Options{Addr: addr})}
}

func (s *Store) Close() error { return s.Client.Close() }

var saveScript = redis.NewScript(`
local country = ARGV[1]
local planes = tonumber(ARGV[2])
local ships = tonumber(ARGV[3])
local timestamp = tonumber(ARGV[4])
local payload = ARGV[5]
local id = ARGV[6]
redis.call('INCR', 'mum:reports:total')
redis.call('HINCRBY', 'mum:country:counts', country, 1)
redis.call('HINCRBY', 'mum:country:planes_sum', country, planes)
redis.call('HINCRBY', 'mum:country:ships_sum', country, ships)
redis.call('HSET', 'mum:country:planes_latest', country, planes)
redis.call('HSET', 'mum:country:ships_latest', country, ships)
local maxp = redis.call('HGET', 'mum:global', 'planes_max')
local minp = redis.call('HGET', 'mum:global', 'planes_min')
local maxs = redis.call('HGET', 'mum:global', 'ships_max')
local mins = redis.call('HGET', 'mum:global', 'ships_min')
if (not maxp) or planes > tonumber(maxp) then redis.call('HSET', 'mum:global', 'planes_max', planes) end
if (not minp) or planes < tonumber(minp) then redis.call('HSET', 'mum:global', 'planes_min', planes) end
if (not maxs) or ships > tonumber(maxs) then redis.call('HSET', 'mum:global', 'ships_max', ships) end
if (not mins) or ships < tonumber(mins) then redis.call('HSET', 'mum:global', 'ships_min', ships) end
redis.call('ZINCRBY', 'mum:freq:planes', 1, tostring(planes))
redis.call('ZINCRBY', 'mum:freq:ships', 1, tostring(ships))
redis.call('ZADD', 'mum:timeseries:' .. country .. ':planes', timestamp, id .. ':' .. planes)
redis.call('ZADD', 'mum:timeseries:' .. country .. ':ships', timestamp, id .. ':' .. ships)
redis.call('LPUSH', 'mum:reports:recent', payload)
redis.call('LTRIM', 'mum:reports:recent', 0, 9999)
return redis.call('GET', 'mum:reports:total')
`)

func (s *Store) Save(ctx context.Context, report model.WarReport) error {
	if err := report.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}
	_, err = saveScript.Run(ctx, s.Client, nil,
		report.Country, report.WarplanesInAir, report.WarshipsInWater,
		report.Timestamp.UnixMilli(), string(payload), report.ID).Result()
	if err != nil {
		return fmt.Errorf("save report: %w", err)
	}
	return nil
}
