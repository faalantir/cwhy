package data

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/supabase-community/supabase-go"
)

type FixEntry struct {
	ID          int    `json:"id"`
	ErrorHash   string `json:"error_hash"`
	ErrorText   string `json:"error_text"`
	FixSolution string `json:"fix_solution"`
}

type Client struct {
	db *supabase.Client
}

func NewClient(url, key string) (*Client, error) {
	client, err := supabase.NewClient(url, key, nil)
	if err != nil {
		return nil, err
	}
	return &Client{db: client}, nil
}

// GenerateHash creates a unique ID for an error message.
func GenerateHash(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(strings.TrimSpace(input)))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GetExistingFix checks if we have seen this error before
func (c *Client) GetExistingFix(hash string) (*FixEntry, error) {
	var results []FixEntry
	
	// FIX 1: Limit now requires two arguments: (count, foreignTable)
	// We pass "" as the second argument because we are querying the main table.
	_, err := c.db.From("fixes").
		Select("*", "exact", false).
		Eq("error_hash", hash).
		Limit(1, ""). 
		ExecuteTo(&results)
	
	if err != nil {
		return nil, err
	}
	
	if len(results) > 0 {
		return &results[0], nil
	}
	
	return nil, nil // Not found
}

// SaveFix stores the new AI solution for next time
func (c *Client) SaveFix(input string, fix string) error {
	hash := GenerateHash(input)
	
	entry := FixEntry{
		ErrorHash:   hash,
		ErrorText:   input,
		FixSolution: fix,
	}

	// Insert arguments: (value, upsert, onConflict, returning, count)
	_, _, err := c.db.From("fixes").
		Insert(entry, false, "", "", "exact").
		Execute()
		
	return err
}