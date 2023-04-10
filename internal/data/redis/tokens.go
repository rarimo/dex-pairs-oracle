package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
)

type tokensQ struct {
	r *redis.Client
}

func (q *tokensQ) Put(ctx context.Context, tokens ...data.Token) error {
	for _, t := range tokens {
		err := q.r.SAdd(ctx, makeChainTokensKey(t.ChainID), makeTokenKey(t.Address, t.ChainID), 0).Err()
		if err != nil {
			return errors.Wrap(err, "failed to add token to chain tokens set", logan.F{
				"address":  t.Address,
				"chain_id": t.ChainID,
			})
		}

		encoded, err := json.Marshal(t)
		if err != nil {
			return errors.Wrap(err, "failed to marshal token", logan.F{
				"address":  t.Address,
				"chain_id": t.ChainID,
			})
		}

		err = q.r.Set(ctx, makeTokenKey(t.Address, t.ChainID), encoded, 0).Err()
		if err != nil {
			return errors.Wrap(err, "failed to set token", logan.F{
				"address":  t.Address,
				"chain_id": t.ChainID,
			})
		}
	}

	return nil
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
