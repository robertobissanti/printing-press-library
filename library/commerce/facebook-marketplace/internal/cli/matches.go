package cli

import (
	"database/sql"
	"encoding/json"

	"github.com/spf13/cobra"
)

func newMatchesCmd(flags *rootFlags) *cobra.Command {
	var onlyNew bool
	cmd := &cobra.Command{
		Use:   "matches",
		Short: "Show local Marketplace watch matches",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			db, err := openLocalDB()
			if err != nil {
				return err
			}
			defer db.Close()
			query := `SELECT m.id, m.watch_id, m.listing_id, l.title, l.price_cents, m.deterministic_ok, m.llm_relevant, m.reason, m.is_new, m.created_at
				FROM matches m JOIN listings l ON l.id = m.listing_id`
			if onlyNew {
				query += ` WHERE m.is_new = 1`
			}
			query += ` ORDER BY m.created_at DESC`
			rows, err := db.Query(query)
			if err != nil {
				return err
			}
			defer rows.Close()
			matches, err := scanMatches(rows)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(matches)
		},
	}
	cmd.Flags().BoolVar(&onlyNew, "new", false, "Only show new matches")
	return cmd
}

func scanMatches(rows *sql.Rows) ([]matchRow, error) {
	var matches []matchRow
	for rows.Next() {
		var m matchRow
		var deterministicOK, llmRelevant, isNew int
		if err := rows.Scan(&m.ID, &m.WatchID, &m.ListingID, &m.Title, &m.PriceCents, &deterministicOK, &llmRelevant, &m.Reason, &isNew, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.DeterministicOK = deterministicOK == 1
		m.LLMRelevant = llmRelevant == 1
		m.IsNew = isNew == 1
		matches = append(matches, m)
	}
	return matches, rows.Err()
}
