package deals

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
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

	m := map[searchKey][]string{}

	interests := map[itemKey][]string{}

	var userItems []*UserItem

	for {
		doc, err := documents.Next()
		if err != nil {
			fmt.Println(err)
			break
		}
		var user FirestoreUser
		err = doc.DataTo(&user)
		if err != nil {
			continue
		}

		for item, merchants := range user.Items {
			key := searchKey{
				postalCode: user.PostalCode,
				query:      item,
			}
			for _, merchant := range merchants {
				merchantString := strconv.Itoa(merchant)
				m[key] = append(m[key], merchantString)
				ik := itemKey{
					postalCode: user.PostalCode,
					query:      item,
					merchant:   merchant,
				}
				interests[ik] = append(interests[ik], user.DiscordUserID)
			}

		}

		fmt.Println(doc.Data())
	}

	for k, v := range m {
		resp, err := flipp.Search(flipp.SearchParams{
			PostalCode: k.postalCode,
			Query:      k.query,
			Merchants:  v,
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
	fmt.Println(userItems)
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
