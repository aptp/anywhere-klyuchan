package github

import (
	"context"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Repository struct {
	AccessToken string
}

func (r *Repository) GetContributions(ctx context.Context, userName string, from time.Time, to time.Time) ([]int, error) {

	// TODO: 現在のコード, トークのスコープが原因かパブリックなリポジトリのコミット数しか取れない。
	// なので修正する。
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.AccessToken},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	var q struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					Weeks []struct {
						ContributionDays []struct{ ContributionCount githubv4.Int }
					}
				}
			} `graphql:"contributionsCollection(from:$from,to:$to)"`
		} `graphql:"user(login:$userName)"`
	}

	variables := map[string]interface{}{
		"from":     githubv4.DateTime{from},
		"to":       githubv4.DateTime{to},
		"userName": githubv4.String(userName),
	}

	if err := client.Query(ctx, &q, variables); err != nil {
		return []int{}, err
	}

	// github end of slice maybe has int value 7 or less. So get week before last.
	// TODO: move this logic
	end := len(q.User.ContributionsCollection.ContributionCalendar.Weeks) - 1
	w := q.User.ContributionsCollection.ContributionCalendar.Weeks[end]

	c := make([]int, 7)
	if 7 > len(w.ContributionDays) {

		wbl := q.User.ContributionsCollection.ContributionCalendar.Weeks[end-1]
		for i := 0; i < 7-len(w.ContributionDays); i++ {
			c[i] = int(wbl.ContributionDays[len(w.ContributionDays)+i].ContributionCount)
		}

		for i := range w.ContributionDays {
			c[7-len(w.ContributionDays)+i] = int(w.ContributionDays[i].ContributionCount)
		}

	} else {

		for i := range w.ContributionDays {
			c[i] = int(w.ContributionDays[i].ContributionCount)
		}

	}

	return c, nil
}
