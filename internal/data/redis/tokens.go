package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
)

type tokensQ struct {
	r *redis.Client
}

func (q *tokensQ) Put(ctx context.Context, chainID int64, tokens ...data.Token) error {
	setKey := makeChainTokensKey(chainID)

	//currentSize, err := q.r.ZCard(ctx, setKey).Result()
	//if err != nil {
	//	return errors.Wrap(err, "failed to get chain tokens set size", logan.F{
	//		"key": setKey,
	//	})
	//}

	members := make([]redis.Z, len(tokens))

	for i, t := range tokens {
		if t.ChainID != chainID {
			return errors.From(errors.New("token chain id doesn't match chain id"), logan.F{
				"token_addr":     t.Address,
				"token_chain_id": t.ChainID,
				"chain_id":       chainID,
			})
		}

		tokenCursor := int64(1) // + currentSize + int64(i)
		t.Cursor = strconv.FormatInt(tokenCursor, 10)

		encoded, err := json.Marshal(t)
		if err != nil {
			return errors.Wrap(err, "failed to marshal token", logan.F{
				"address":  t.Address,
				"chain_id": t.ChainID,
			})
		}

		tk := makeTokenKey(t.Address, t.ChainID)

		err = q.r.Set(ctx, tk, encoded, 0).Err()
		if err != nil {
			return errors.Wrap(err, "failed to set token", logan.F{
				"address":  t.Address,
				"chain_id": t.ChainID,
			})
		}

		members[i] = redis.Z{
			Score:  float64(tokenCursor),
			Member: tk,
		}
	}

	return q.r.ZAddArgs(ctx, setKey, redis.ZAddArgs{
		Members: members,
		NX:      true,
	}).Err()
}

func (q *tokensQ) Get(ctx context.Context, address string, chainID int64) (*data.Token, error) {
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

	var token data.Token
	err = json.Unmarshal([]byte(raw), &token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal token", logan.F{
			"raw": raw,
		})
	}

	return &token, nil
}

func (q *tokensQ) Page(ctx context.Context, chainID int64, cursor string, limit int64) ([]data.Token, error) {
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

	tokens := make([]data.Token, len(rawTokens))
	for i, raw := range rawTokens {
		rawS, ok := raw.(string)
		if !ok {
			return nil, errors.From(errors.New("failed to cast token to string"), logan.F{
				"raw": raw,
			})
		}

		var token data.Token
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
		return tokens[i].Address < tokens[j].Address
	})

	return tokens, nil
}

func (q *tokensQ) All(ctx context.Context, chain int64) ([]data.Token, error) {
	tokenKeys, err := q.r.SMembers(ctx, makeChainTokensKey(chain)).Result()
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

	tokens := make([]data.Token, len(tokenKeys))
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

		var token data.Token
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
	return fmt.Sprintf("token:%d:%s", chainID, address)
}

func makeChainTokensKey(chainID int64) string {
	return fmt.Sprintf("chain_tokens:%d", chainID)
}
