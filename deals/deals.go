package deals

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/m-brady/go-flipp/flipp"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type UserItem struct {
	DiscordUserID string
	Item          *flipp.ItemDetails
}

func Scan(firestoreClient *firestore.Client) []*UserItem {
	documents := firestoreClient.Collection("users").Documents(context.Background())

	searchRequests := make(map[searchKey]map[int]struct{})

	interests := make(map[itemKey][]string)

	var userItems []*UserItem

	for {
		doc, err := documents.Next()
		if err != nil {
			log.Error(err)
			break
		}
		var user FirestoreUser
		err = doc.DataTo(&user)
		if err != nil {
			log.Error(err)
			continue
		}

		log.Infof("Looking at user %v", user)

		for _, item := range user.Items {
			key := searchKey{
				postalCode: user.PostalCode,
				query:      item,
			}
			if searchRequests[key] == nil {
				searchRequests[key] = make(map[int]struct{})
			}
			for _, m := range user.Merchants {
				searchRequests[key][m] = struct{}{}

				ik := itemKey{
					postalCode: user.PostalCode,
					query:      item,
					merchant:   m,
				}
				interests[ik] = append(interests[ik], user.DiscordUserID)

			}

		}
	}

	log.Info(searchRequests)
	log.Info(interests)

	for k, v := range searchRequests {

		merchants := make([]string, len(v))
		for m := range v {
			merchants = append(merchants, strconv.Itoa(m))
		}

		resp, err := flipp.Search(flipp.SearchParams{
			PostalCode: k.postalCode,
			Query:      k.query,
			Merchants:  merchants,
		})
		if err != nil {
			log.Errorf("Error searching for %v %v. %v", k, v, err)
			continue
		}
		for _, item := range resp.ItemDetails {
			users := interests[itemKey{
				postalCode: k.postalCode,
				query:      k.query,
				merchant:   item.MerchantID,
			}]
			for _, user := range users {
				userItems = append(userItems, &UserItem{
					DiscordUserID: user,
					Item:          item,
				})
			}
		}

	}
	return userItems

}

type searchKey struct {
	postalCode string
	query      string
}

type itemKey struct {
	postalCode string
	query      string
	merchant   int
}
