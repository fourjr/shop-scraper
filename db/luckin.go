package db

import (
	"context"
	"errors"
	"fmt"
	"shops/luckin"
	"shops/models"
	"strconv"
)

func (c *Postgres) GetLuckinAccount(ctx context.Context, consideration int) (*models.LuckinAccount, error) {
	var account models.LuckinAccount
	err := c.db.QueryRow(ctx,
		"SELECT email, password, token FROM luckin_login ORDER BY md5(email::text || ':' || $1::text) LIMIT 1",
		strconv.Itoa(consideration),
	).Scan(&account.Email, &account.Password, &account.Token)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("no luckin account found in table luckin_login")
		}
		return nil, err
	}

	if account.Token == nil {
		token, err := c.UpdateLuckinToken(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("failed to update login token for email %s: %v", account.Email, err)
		}
		account.Token = &token
	}
	return &account, nil
}

func (c *Postgres) UpdateLuckinToken(ctx context.Context, account models.LuckinAccount) (string, error) {
	token, err := luckin.RefreshLoginToken(account.Email, account.Password)
	if err != nil {
		return "", fmt.Errorf("failed to refresh login token for email %s: %v", account.Email, err)
	}
	_, err = c.db.Exec(ctx,
		"UPDATE luckin_login SET token = $1 WHERE email = $2",
		token, account.Email,
	)
	if err != nil {
		return "", fmt.Errorf("failed to update login token for email %s: %v", account.Email, err)
	}
	return token, nil
}
