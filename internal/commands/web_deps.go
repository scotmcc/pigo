package commands

import "github.com/scotmcc/pigo/internal/search"

// searchClient is set at startup by SetSearchClient.
var searchClient *search.Client

// SetSearchClient wires the search client into web commands.
func SetSearchClient(client *search.Client) {
	searchClient = client
}
