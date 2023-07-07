package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"

	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type tokensQ struct {
	r *redis.Client
}

func (q *tokensQ) Put(ctx context.Context, chainID int64, tokens ...chains.TokenInfo) error {
	setKey := makeChainTokensKey(chainID)

	members := make([]redis.Z, len(tokens))

	for i, t := range tokens {
		encoded, err := json.Marshal(t)
		if err != nil {
			return errors.Wrap(err, "failed to marshal token", logan.F{
				"address": t.Address,
			})
		}

		tk := makeTokenKey(t.Address, chainID)

		err = q.r.Set(ctx, tk, encoded, 0).Err()
		if err != nil {
			return errors.Wrap(err, "failed to set token", logan.F{
				"key": tk,
			})
		}

		members[i] = redis.Z{
			Score:  float64(1),
			Member: tk,
		}
	}

	return q.r.ZAddArgs(ctx, setKey, redis.ZAddArgs{
		Members: members,
		NX:      true,
	}).Err()
}

func (q *tokensQ) Get(ctx context.Context, address string, chainID int64) (*chains.TokenInfo, error) {
	key := makeTokenKey(address, chainID)

	raw, err := q.r.Get(ctx, key).Result()
	if err != nil {
		if errors.Cause(err) == redis.Nil {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to get token", logan.F{
			"address":  address,
			"chain_id": chainID,
		})
	}

	if raw == "" {
		return nil, nil
	}

	var token chains.TokenInfo
	err = json.Unmarshal([]byte(raw), &token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal token", logan.F{
			"raw": raw,
		})
	}

	return &token, nil
}

func (q *tokensQ) Page(ctx context.Context, chainID int64, cursor string, limit int64) ([]chains.TokenInfo, error) {
	setKey := makeChainTokensKey(chainID)

	start := "-"
	if cursor != "" {
		start = fmt.Sprintf("(%s", cursor)
	}

	// zrange chain_tokens:56 ({cursor} + bylex limit 0 {limit}
	tokenKeys, err := q.r.ZRangeByLex(ctx, setKey, &redis.ZRangeBy{
		Min:   start,
		Max:   "+",
		Count: limit,
	}).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token keys", logan.F{
			"key": setKey,
		})
	}

	if len(tokenKeys) == 0 {
		return nil, nil
	}

	rawTokens, err := q.r.MGet(ctx, tokenKeys...).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tokens", logan.F{
			"keys": tokenKeys,
		})
	}

	tokens := make([]chains.TokenInfo, len(rawTokens))
	for i, raw := range rawTokens {
		rawS, ok := raw.(string)
		if !ok {
			return nil, errors.From(errors.New("failed to cast token to string"), logan.F{
				"raw": raw,
			})
		}

		var token chains.TokenInfo
		err = json.Unmarshal([]byte(rawS), &token)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal token", logan.F{
				"raw": raw,
			})
		}

		tokens[i] = token
	}

	// to make sure that tokens are sorted by address after mget
	sort.Slice(tokens, func(i, j int) bool {
		return strings.ToLower(tokens[i].Address) < strings.ToLower(tokens[j].Address)
	})

	return tokens, nil
}

func (q *tokensQ) All(ctx context.Context, chain int64) ([]chains.TokenInfo, error) {
	tokenKeys, err := q.r.ZRevRange(ctx, makeChainTokensKey(chain), 0, -1).Result()
	if err != nil {
		if errors.Cause(err) == redis.Nil {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to get token keys", logan.F{
			"chain_id": chain,
		})
	}

	if len(tokenKeys) == 0 {
		return nil, nil
	}

	tokens := make([]chains.TokenInfo, len(tokenKeys))
	for i, key := range tokenKeys {
		raw, err := q.r.Get(ctx, key).Result()
		if err != nil {
			if errors.Cause(err) == redis.Nil {
				continue
			}

			return nil, errors.Wrap(err, "failed to get token", logan.F{
				"key": key,
			})
		}

		var token chains.TokenInfo
		err = json.Unmarshal([]byte(raw), &token)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal token", logan.F{
				"raw": raw,
			})
		}

		tokens[i] = token
	}

	return tokens, nil
}

func makeTokenKey(address string, chainID int64) string {
	return strings.ToLower(fmt.Sprintf("token:%d:%s", chainID, address))
}

func makeChainTokensKey(chainID int64) string {
	return fmt.Sprintf("chain_tokens:%d", chainID)
}
