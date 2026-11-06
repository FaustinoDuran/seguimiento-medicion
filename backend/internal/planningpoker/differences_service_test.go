package planningpoker

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRoundDifferencesAreOnlyReportedAfterReveal(t *testing.T) {
	cases := []struct {
		name   string
		points []int64
		want   bool
	}{
		{name: "different", points: []int64{5, 8}, want: true},
		{name: "equal", points: []int64{5, 5}},
		{name: "one vote", points: []int64{5}},
		{name: "no votes", points: []int64{}},
		{name: "difference after equal votes", points: []int64{5, 5, 8}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := testService()
			service.members = fakeMembers{members: map[int64]MemberRef{
				7: {ID: 7, ProjectID: 1},
				8: {ID: 8, ProjectID: 1},
				9: {ID: 9, ProjectID: 1},
			}}
			ctx := context.Background()
			round, err := service.OpenRound(ctx, 1, 4)
			if err != nil {
				t.Fatal(err)
			}
			for i, points := range tc.points {
				_, err := service.CastVote(ctx, CastVote{
					ProjectID: 1, RoundID: round.ID, MemberID: int64(7 + i), StoryPoints: points,
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			hidden, err := service.View(ctx, 1, round.ID, 7)
			if err != nil {
				t.Fatal(err)
			}
			if hidden.HasDifferences != nil {
				t.Fatalf("open round exposes comparison: %+v", hidden)
			}
			for _, vote := range hidden.Votes {
				if vote.MemberID != 7 && vote.StoryPoints != nil {
					t.Fatalf("open round exposes another vote: %+v", vote)
				}
			}
			revealed, err := service.Reveal(ctx, 1, round.ID, 7)
			if err != nil {
				t.Fatal(err)
			}
			if revealed.HasDifferences == nil || *revealed.HasDifferences != tc.want {
				t.Fatalf("reveal comparison: %+v, want %v", revealed.HasDifferences, tc.want)
			}
			view, err := service.View(ctx, 1, round.ID, 8)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(revealed, view) {
				t.Fatalf("later view differs: got %+v want %+v", view, revealed)
			}
			if _, err := service.OpenRound(ctx, 1, 4); err != nil {
				t.Fatalf("voluntary new round: %v", err)
			}
		})
	}
}

func TestNewRoundPreservesPreviousVotesAndHidesNewVotes(t *testing.T) {
	service, _ := testService()
	ctx := context.Background()
	first, err := service.OpenRound(ctx, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range []int64{7, 8} {
		if _, err := service.CastVote(ctx, CastVote{
			ProjectID: 1, RoundID: first.ID, MemberID: member, StoryPoints: 5,
		}); err != nil {
			t.Fatal(err)
		}
	}
	// Replacing a vote must affect the comparison, without adding another vote.
	if _, err := service.CastVote(ctx, CastVote{
		ProjectID: 1, RoundID: first.ID, MemberID: 8, StoryPoints: 8,
	}); err != nil {
		t.Fatal(err)
	}
	before, err := service.Reveal(ctx, 1, first.ID, 7)
	if err != nil {
		t.Fatal(err)
	}
	if before.HasDifferences == nil || !*before.HasDifferences || len(before.Votes) != 2 {
		t.Fatalf("replacement was not compared: %+v", before)
	}
	second, err := service.OpenRound(ctx, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == first.ID || second.Status != StatusOpen || second.StoryID != first.StoryID {
		t.Fatalf("new round: %+v", second)
	}
	empty, err := service.View(ctx, 1, second.ID, 7)
	if err != nil || len(empty.Votes) != 0 || empty.HasDifferences != nil {
		t.Fatalf("new round not empty and secret: %+v %v", empty, err)
	}
	for _, member := range []int64{7, 8} {
		if _, err := service.CastVote(ctx, CastVote{
			ProjectID: 1, RoundID: second.ID, MemberID: member, StoryPoints: 13,
		}); err != nil {
			t.Fatal(err)
		}
	}
	hidden, err := service.View(ctx, 1, second.ID, 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, other := splitVotes(t, hidden.Votes, 7); other.StoryPoints != nil || hidden.HasDifferences != nil {
		t.Fatalf("new round leaks comparison or points: %+v", hidden)
	}
	after, err := service.View(ctx, 1, first.ID, 8)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("old round changed: %+v %v", after, err)
	}
	if _, err := service.CastVote(ctx, CastVote{
		ProjectID: 1, RoundID: first.ID, MemberID: 7, StoryPoints: 21,
	}); !errors.Is(err, ErrRoundClosed) {
		t.Fatalf("old round accepts votes: %v", err)
	}
	if _, err := service.OpenRound(ctx, 1, 4); !errors.Is(err, ErrOpenRoundExists) {
		t.Fatalf("third round: %v", err)
	}
	if _, err := service.OpenRound(ctx, 2, 4); !errors.Is(err, ErrStoryNotFound) {
		t.Fatalf("cross-project round: %v", err)
	}
}
